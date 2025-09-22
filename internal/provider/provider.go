package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/knowledge-work/terraform-provider-kw-github/internal/githubclient"
)

func New() provider.Provider {
	return &kwgithubProvider{}
}

type kwgithubProvider struct{}

func (p *kwgithubProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kwgithub"
}

func (p *kwgithubProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The KW GitHub provider is used to interact with GitHub resources specific to Knowledge Work organization.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "GitHub personal access token. Can also be set via GITHUB_TOKEN environment variable.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Description: "GitHub owner name to manage. Can also be set via GITHUB_OWNER environment variable.",
			},
			"github_base_url": schema.StringAttribute{
				Optional:    true,
				Description: "GitHub base URL. Defaults to https://api.github.com. Can also be set via GITHUB_BASE_URL environment variable.",
			},
		},
		Blocks: map[string]schema.Block{
			"app_auth": schema.ListNestedBlock{
				Description: "GitHub App authentication configuration. Conflicts with token.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:    true,
							Description: "GitHub App ID. Can also be set via GITHUB_APP_ID environment variable.",
						},
						"installation_id": schema.StringAttribute{
							Optional:    true,
							Description: "GitHub App installation ID. Can also be set via GITHUB_APP_INSTALLATION_ID environment variable.",
						},
						"pem_file": schema.StringAttribute{
							Optional:    true,
							Sensitive:   true,
							Description: "GitHub App private key PEM file contents. Can also be set via GITHUB_APP_PEM_FILE environment variable.",
						},
					},
				},
			},
		},
	}
}

func (p *kwgithubProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config struct {
		Token         string `tfsdk:"token"`
		Owner         string `tfsdk:"owner"`
		GithubBaseURL string `tfsdk:"github_base_url"`
		AppAuth       []struct {
			ID             string `tfsdk:"id"`
			InstallationID string `tfsdk:"installation_id"`
			PemFile        string `tfsdk:"pem_file"`
		} `tfsdk:"app_auth"`
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := "https://api.github.com"
	if config.GithubBaseURL != "" {
		baseURL = config.GithubBaseURL
	} else if envBaseURL := os.Getenv("GITHUB_BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	var token string
	var client *githubclient.Client

	if len(config.AppAuth) > 0 {
		appAuth := config.AppAuth[0]
		appID := appAuth.ID
		if appID == "" {
			appID = os.Getenv("GITHUB_APP_ID")
		}
		installationID := appAuth.InstallationID
		if installationID == "" {
			installationID = os.Getenv("GITHUB_APP_INSTALLATION_ID")
		}
		pemFile := appAuth.PemFile
		if pemFile == "" {
			pemFile = os.Getenv("GITHUB_APP_PEM_FILE")
		}

		if appID == "" || installationID == "" || pemFile == "" {
			resp.Diagnostics.AddError(
				"Incomplete GitHub App configuration",
				"app_auth.id, app_auth.installation_id, and app_auth.pem_file must all be set either in configuration or via environment variables (GITHUB_APP_ID, GITHUB_APP_INSTALLATION_ID, GITHUB_APP_PEM_FILE)",
			)
			return
		}

		pemFile = strings.Replace(pemFile, `\n`, "\n", -1)
		client = githubclient.NewClientWithApp(appID, installationID, pemFile, baseURL)
		if client == nil {
			resp.Diagnostics.AddError(
				"Failed to create GitHub App client",
				"Unable to generate access token from GitHub App credentials",
			)
			return
		}
	} else {
		token = config.Token
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		if token == "" {
			resp.Diagnostics.AddError("Missing authentication", "Either token or app_auth must be configured")
			return
		}
		client = githubclient.NewClient(token, baseURL)
	}

	resp.ResourceData = client
}

func (p *kwgithubProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *kwgithubProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRulesetAllowedMergeMethodsResource,
	}
}
