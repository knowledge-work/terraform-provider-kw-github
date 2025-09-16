package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/knowledge-work/knowledgework/terraform-provider/github-repository-rule/internal/githubclient"
)

func NewRulesetAllowedMergeMethodsResource() resource.Resource {
	return &rulesetAllowedMergeMethodsResource{}
}

type rulesetAllowedMergeMethodsResource struct {
	client *githubclient.Client
}

type rulesetAllowedMergeMethodsResourceModel struct {
	Repository          types.String `tfsdk:"repository"`
	RulesetID           types.String `tfsdk:"ruleset_id"`
	AllowedMergeMethods types.Set    `tfsdk:"allowed_merge_methods"`
	ForceUpdate         types.String `tfsdk:"force_update"`
	ID                  types.String `tfsdk:"id"`
}

func (r *rulesetAllowedMergeMethodsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_ruleset_allowed_merge_methods"
}

func (r *rulesetAllowedMergeMethodsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repository": schema.StringAttribute{
				Required: true,
			},
			"ruleset_id": schema.StringAttribute{
				Required: true,
			},
			"allowed_merge_methods": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"force_update": schema.StringAttribute{
				Optional:    true,
				Description: "Timestamp to force update when dependent resources change. Set this to a new value (e.g., timestamp) when you want to force an update.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *rulesetAllowedMergeMethodsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*githubclient.Client)
}

func (r *rulesetAllowedMergeMethodsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan rulesetAllowedMergeMethodsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.upsert(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ruleset", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.Repository.ValueString(), plan.RulesetID.ValueString()))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *rulesetAllowedMergeMethodsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state rulesetAllowedMergeMethodsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner, repo, err := parseRepo(state.Repository.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid repository format", err.Error())
		return
	}

	rulesetID, err := parseID(state.RulesetID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset ID", err.Error())
		return
	}

	ruleset, _, err := r.client.Repositories.GetRuleset(ctx, owner, repo, rulesetID, true)
	if err != nil {
		resp.Diagnostics.AddError("Error reading a ruleset", err.Error())
		return
	}

	var currentMethods []string
	if ruleset.Rules != nil && ruleset.Rules.PullRequest != nil {
		for _, method := range ruleset.Rules.PullRequest.AllowedMergeMethods {
			currentMethods = append(currentMethods, string(method))
		}
	}

	// Check if current methods differ from expected methods
	expectedMethods := extractMethodsFromSet(state.AllowedMergeMethods)
	if !methodsEqual(currentMethods, expectedMethods) {
		// Methods have been reset by GitHub, restore them
		err = r.restoreMergeMethods(ctx, owner, repo, rulesetID, expectedMethods)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Merge methods were reset",
				fmt.Sprintf("GitHub reset the merge methods, attempted to restore but failed: %v", err),
			)
		} else {
			// Successfully restored, use expected methods
			currentMethods = expectedMethods
		}
	}

	state.AllowedMergeMethods = convertToSet(currentMethods)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *rulesetAllowedMergeMethodsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan rulesetAllowedMergeMethodsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.upsert(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating ruleset", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *rulesetAllowedMergeMethodsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state rulesetAllowedMergeMethodsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner, repo, err := parseRepo(state.Repository.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid repository format", err.Error())
		return
	}

	rulesetID, err := parseID(state.RulesetID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ruleset ID", err.Error())
		return
	}

	ruleset, _, err := r.client.Repositories.GetRuleset(ctx, owner, repo, rulesetID, true)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ruleset", err.Error())
		return
	}

	// Reset to default merge methods (all methods allowed)
	defaultMethods := []github.PullRequestMergeMethod{
		github.PullRequestMergeMethodMerge,
		github.PullRequestMergeMethodSquash,
		github.PullRequestMergeMethodRebase,
	}

	if ruleset.Rules == nil {
		ruleset.Rules = &github.RepositoryRulesetRules{}
	}
	if ruleset.Rules.PullRequest == nil {
		ruleset.Rules.PullRequest = &github.PullRequestRuleParameters{}
	}
	ruleset.Rules.PullRequest.AllowedMergeMethods = defaultMethods

	_, _, err = r.client.Repositories.UpdateRuleset(ctx, owner, repo, rulesetID, *ruleset)
	if err != nil {
		resp.Diagnostics.AddError("Error resetting merge methods", err.Error())
		return
	}
}

func (r *rulesetAllowedMergeMethodsResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: owner/repo:ruleset_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ruleset_id"), parts[1])...)
}

func (r *rulesetAllowedMergeMethodsResource) upsert(
	ctx context.Context,
	plan *rulesetAllowedMergeMethodsResourceModel,
) error {
	owner, repo, err := parseRepo(plan.Repository.ValueString())
	if err != nil {
		return err
	}

	rulesetID, err := parseID(plan.RulesetID.ValueString())
	if err != nil {
		return err
	}

	ruleset, _, err := r.client.Repositories.GetRuleset(ctx, owner, repo, rulesetID, true)
	if err != nil {
		return err
	}

	var methods []github.PullRequestMergeMethod
	for _, v := range plan.AllowedMergeMethods.Elements() {
		s, _ := v.(types.String)
		methods = append(methods, github.PullRequestMergeMethod(s.ValueString()))
	}

	if ruleset.Rules == nil {
		ruleset.Rules = &github.RepositoryRulesetRules{}
	}
	if ruleset.Rules.PullRequest == nil {
		ruleset.Rules.PullRequest = &github.PullRequestRuleParameters{}
	}
	ruleset.Rules.PullRequest.AllowedMergeMethods = methods

	_, _, err = r.client.Repositories.UpdateRuleset(ctx, owner, repo, rulesetID, *ruleset)
	return err
}

func parseRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format")
	}
	return parts[0], parts[1], nil
}

func parseID(id string) (int64, error) {
	var i int64
	n, err := fmt.Sscanf(id, "%d", &i)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, fmt.Errorf("invalid ID format")
	}
	// Check if the entire string was consumed
	if fmt.Sprintf("%d", i) != id {
		return 0, fmt.Errorf("invalid ID format")
	}
	return i, nil
}

func convertToSet(methods []string) types.Set {
	var elems []types.String
	for _, m := range methods {
		elems = append(elems, types.StringValue(m))
	}
	set, _ := types.SetValueFrom(context.Background(), types.StringType, elems)
	return set
}

// extractMethodsFromSet extracts string slice from Terraform Set
func extractMethodsFromSet(set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return []string{}
	}

	var methods []string
	for _, elem := range set.Elements() {
		if str, ok := elem.(types.String); ok {
			methods = append(methods, str.ValueString())
		}
	}
	return methods
}

// methodsEqual compares two string slices for equality (order-independent)
func methodsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison
	mapA := make(map[string]bool)
	mapB := make(map[string]bool)

	for _, method := range a {
		mapA[method] = true
	}
	for _, method := range b {
		mapB[method] = true
	}

	// Compare maps
	for method := range mapA {
		if !mapB[method] {
			return false
		}
	}
	return true
}

// restoreMergeMethods restores the merge methods to the expected values
func (r *rulesetAllowedMergeMethodsResource) restoreMergeMethods(
	ctx context.Context,
	owner, repo string,
	rulesetID int64,
	expectedMethods []string,
) error {
	plan := &rulesetAllowedMergeMethodsResourceModel{
		Repository:          types.StringValue(fmt.Sprintf("%s/%s", owner, repo)),
		RulesetID:           types.StringValue(fmt.Sprintf("%d", rulesetID)),
		AllowedMergeMethods: convertToSet(expectedMethods),
	}

	return r.upsert(ctx, plan)
}
