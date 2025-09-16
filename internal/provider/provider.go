package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/knowledge-work/knowledgework/terraform-provider/github-repository-rule/internal/githubclient"
)

func New() provider.Provider {
	return &kwgithubProvider{}
}

type kwgithubProvider struct{}

func (p *kwgithubProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kw_github"
}

func (p *kwgithubProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"github_base_url": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *kwgithubProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config struct {
		Token         string `tfsdk:"token"`
		GithubBaseURL string `tfsdk:"github_base_url"`
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := config.Token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		resp.Diagnostics.AddError("Missing GITHUB_TOKEN", "")
		return
	}

	baseURL := "https://api.github.com"
	if config.GithubBaseURL != "" {
		baseURL = config.GithubBaseURL
	}

	client := githubclient.NewClient(token, baseURL)
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
