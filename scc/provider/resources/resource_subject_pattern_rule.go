package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	subjectpatternrules "github.com/SAP/terraform-provider-scc/validation/subjectPatternRules"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &SubjectPatternRuleResource{}

// subjectPatternRuleMu serializes all mutating API calls because the SCC API
// does not support concurrent modifications (returns ConcurrentModificationException).
var subjectPatternRuleMu sync.Mutex

func NewSubjectPatternRuleResource() resource.Resource {
	return &SubjectPatternRuleResource{}
}

type SubjectPatternRuleResource struct {
	Client *api.RestApiClient
}

type subjectPatternRuleResourceIdentityModel struct {
	Index types.Int64 `tfsdk:"index"`
}

func (d *SubjectPatternRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subject_pattern_rule"
}

func (d *SubjectPatternRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Subject Pattern Rule in the SAP Cloud Connector instance.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Associate Administrator
	* Subaccount Administrator
	* Display
	* Support
	* Monitoring
* The system-level rule with condition "always true" cannot be managed by this resource. Attempting to import it will return an error.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/subject-pattern-rules>`,
		Attributes: map[string]schema.Attribute{
			"index": schema.Int64Attribute{
				MarkdownDescription: "Index of the subject pattern rule to retrieve.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the subject pattern rule.",
				Required:            true,
			},
			"condition": schema.SingleNestedAttribute{
				MarkdownDescription: "Condition of the subject pattern rule.",
				Required:            true,
				Validators: []validator.Object{
					subjectpatternrules.SubjectPatternRuleCondition(),
				},
				Attributes: map[string]schema.Attribute{
					"variable": schema.StringAttribute{
						MarkdownDescription: "Variable of the condition to be evaluated.",
						Required:            true,
					},
					"operator": schema.StringAttribute{
						MarkdownDescription: "Operator of the condition.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"exist",
								"exist_not",
								"is",
								"is_not",
							),
						},
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Value of the condition. Required when operator is \"is\" or \"is_not\".",
						Optional:            true,
					},
				},
			},
			"subject_pattern": schema.SingleNestedAttribute{
				MarkdownDescription: "Subject pattern of the subject pattern rule.",
				Required:            true,
				Validators: []validator.Object{
					subjectpatternrules.SubjectPatternAtLeastOneField(),
				},
				Attributes: map[string]schema.Attribute{
					"cn": schema.StringAttribute{
						MarkdownDescription: "Common Name (CN) of the subject pattern.",
						Optional:            true,
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Email (EMAIL) of the subject pattern.",
						Optional:            true,
						Computed:            true,
					},
					"l": schema.StringAttribute{
						MarkdownDescription: "Locality (L) of the subject pattern.",
						Optional:            true,
						Computed:            true,
					},
					"ou": schema.StringAttribute{
						MarkdownDescription: "Organization Unit (OU) of the subject pattern.",
						Optional:            true,
						Computed:            true,
					},
					"o": schema.StringAttribute{
						MarkdownDescription: "Organization (O) of the subject pattern.",
						Optional:            true,
						Computed:            true,
					},
					"st": schema.StringAttribute{
						MarkdownDescription: "State (ST) of the subject pattern.",
						Optional:            true,
						Computed:            true,
					},
					"c": schema.StringAttribute{
						MarkdownDescription: "Country (C) of the subject pattern.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}

func (rs *SubjectPatternRuleResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"index": identityschema.Int64Attribute{
				RequiredForImport: true,
			},
		},
	}
}

func (d *SubjectPatternRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.RestApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.RestApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.Client = client
}

func (r *SubjectPatternRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	subjectPatternRuleMu.Lock()
	defer subjectPatternRuleMu.Unlock()

	var plan model.SubjectPatternRuleConfig
	var respObj apiobjects.SubjectPatternRules
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSubjectPatternRulesBaseEndpoint()

	var condition model.SubjectPatternCondition
	diags = plan.Condition.As(ctx, &condition, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var subjectPattern model.SubjectPattern
	diags = plan.SubjectPattern.As(ctx, &subjectPattern, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planBody := map[string]any{
		"description": plan.Description.ValueString(),
		"condition":   buildConditionBody(condition),
		"subjectPattern": apiobjects.SubjectPattern{
			CommonName:       subjectPattern.CommonName.ValueString(),
			Email:            subjectPattern.Email.ValueString(),
			Locality:         subjectPattern.Locality.ValueString(),
			OrganizationUnit: subjectPattern.OrganizationUnit.ValueString(),
			Organization:     subjectPattern.Organization.ValueString(),
			State:            subjectPattern.State.ValueString(),
			Country:          subjectPattern.Country.ValueString(),
		},
	}

	// Create the subject pattern rule
	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Find all subject pattern rule in the response
	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Find the created subject pattern rule in the response
	rule, index, err := findSubjectPatternRule(respObj, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to locate created subject pattern rule",
			err.Error(),
		)
		return
	}

	plan.Index = types.Int64Value(int64(index))

	model, diags := model.SubjectPatternRuleValueFrom(ctx, plan, *rule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subjectPatternRuleResourceIdentityModel{
		Index: plan.Index,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SubjectPatternRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if req.State.Raw.IsNull() {
		return
	}

	var state model.SubjectPatternRuleConfig
	var respObj apiobjects.SubjectPatternRules

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = helpers.RequestAndUnmarshal(
		r.Client,
		&respObj,
		"GET",
		endpoints.GetSubjectPatternRulesBaseEndpoint(),
		nil,
		true,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		rule     *apiobjects.SubjectPatternRule
		newIndex int
		err      error
	)

	// Import case: only index is known.
	if state.Description.IsNull() || state.Description.IsUnknown() {
		idx := int(state.Index.ValueInt64())

		if idx < 0 || idx >= len(respObj) {
			resp.State.RemoveResource(ctx)
			return
		}

		rule = &respObj[idx]
		newIndex = idx
	} else {
		// Normal CRUD case.
		rule, newIndex, err = findSubjectPatternRule(respObj, state)
		if err != nil {
			resp.State.RemoveResource(ctx)
			return
		}
	}

	state.Index = types.Int64Value(int64(newIndex))

	if strings.TrimSpace(rule.Condition) == "always true" {
		resp.Diagnostics.AddError(
			"Rule cannot be managed",
			fmt.Sprintf(
				"The subject pattern rule at index %d has an \"always true\" condition. "+
					"This is a system-level rule that cannot be managed by Terraform.",
				newIndex,
			),
		)
		return
	}

	model, diags := model.SubjectPatternRuleValueFrom(ctx, state, *rule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)

	identity := subjectPatternRuleResourceIdentityModel{
		Index: state.Index,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SubjectPatternRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	subjectPatternRuleMu.Lock()
	defer subjectPatternRuleMu.Unlock()

	var plan, state model.SubjectPatternRuleConfig
	var respObj apiobjects.SubjectPatternRules

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If nothing has changed, return early
	if plan.Description.Equal(state.Description) &&
		plan.Condition.Equal(state.Condition) &&
		plan.SubjectPattern.Equal(state.SubjectPattern) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	index := int(state.Index.ValueInt64())
	endpoint := endpoints.GetSubjectPatternRuleByIndexEndpoint(index)

	var condition model.SubjectPatternCondition
	diags = plan.Condition.As(ctx, &condition, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var subjectPattern model.SubjectPattern
	diags = plan.SubjectPattern.As(ctx, &subjectPattern, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateBody := map[string]any{
		"description": plan.Description.ValueString(),
		"condition":   buildConditionBody(condition),
		"subjectPattern": apiobjects.SubjectPattern{
			CommonName:       subjectPattern.CommonName.ValueString(),
			Email:            subjectPattern.Email.ValueString(),
			Locality:         subjectPattern.Locality.ValueString(),
			OrganizationUnit: subjectPattern.OrganizationUnit.ValueString(),
			Organization:     subjectPattern.Organization.ValueString(),
			State:            subjectPattern.State.ValueString(),
			Country:          subjectPattern.Country.ValueString(),
		},
	}

	// Update the subject pattern rule
	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "PUT", endpoint, updateBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint = endpoints.GetSubjectPatternRulesBaseEndpoint()

	// Find all subject pattern rule in the response
	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, newIndex, err := findSubjectPatternRule(respObj, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to locate updated subject pattern rule",
			err.Error(),
		)
		return
	}

	plan.Index = types.Int64Value(int64(newIndex))

	model, diags := model.SubjectPatternRuleValueFrom(ctx, plan, *rule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subjectPatternRuleResourceIdentityModel{
		Index: plan.Index,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SubjectPatternRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to delete
	if req.State.Raw.IsNull() {
		return
	}

	var state model.SubjectPatternRuleConfig
	var respObj apiobjects.SubjectPatternRules

	subjectPatternRuleMu.Lock()
	defer subjectPatternRuleMu.Unlock()

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-fetch the live list to resolve the current index, since parallel
	// deletions shift indices of subsequent rules.
	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoints.GetSubjectPatternRulesBaseEndpoint(), nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, liveIndex, err := findSubjectPatternRule(respObj, state)
	if err != nil {
		// Already gone — treat as success.
		resp.State.RemoveResource(ctx)
		return
	}

	endpoint := endpoints.GetSubjectPatternRuleByIndexEndpoint(liveIndex)

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (rs *SubjectPatternRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		idParts := strings.Split(req.ID, ",")

		if len(idParts) != 1 || idParts[0] == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: index. Got: %q", req.ID),
			)
			return
		}

		idx, err := strconv.ParseInt(idParts[0], 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import Identifier",
				fmt.Sprintf("Expected a numeric index, got %q: %v", idParts[0], err),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("index"), idx)...)

		return
	}

	var identity subjectPatternRuleResourceIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("index"), identity.Index)...)
}

func buildConditionBody(condition model.SubjectPatternCondition) map[string]any {
	m := map[string]any{
		"variable": condition.Variable.ValueString(),
		"operator": condition.Operator.ValueString(),
	}
	op := condition.Operator.ValueString()
	if op != "exist" && op != "exist_not" {
		m["value"] = condition.Value.ValueString()
	}
	return m
}

func findSubjectPatternRule(rules apiobjects.SubjectPatternRules, plan model.SubjectPatternRuleConfig) (*apiobjects.SubjectPatternRule, int, error) {
	var condition model.SubjectPatternCondition
	if diags := plan.Condition.As(context.Background(), &condition, basetypes.ObjectAsOptions{}); diags.HasError() {
		return nil, -1, fmt.Errorf("failed to decode condition: %v", diags)
	}

	var subjectPattern model.SubjectPattern
	if diags := plan.SubjectPattern.As(context.Background(), &subjectPattern, basetypes.ObjectAsOptions{}); diags.HasError() {
		return nil, -1, fmt.Errorf("failed to decode subject pattern: %v", diags)
	}

	for i, rule := range rules {
		if rule.Description != plan.Description.ValueString() {
			continue
		}

		if rule.SubjectPattern.CommonName != subjectPattern.CommonName.ValueString() ||
			rule.SubjectPattern.Email != subjectPattern.Email.ValueString() ||
			rule.SubjectPattern.Locality != subjectPattern.Locality.ValueString() ||
			rule.SubjectPattern.OrganizationUnit != subjectPattern.OrganizationUnit.ValueString() ||
			rule.SubjectPattern.Organization != subjectPattern.Organization.ValueString() ||
			rule.SubjectPattern.State != subjectPattern.State.ValueString() ||
			rule.SubjectPattern.Country != subjectPattern.Country.ValueString() {
			continue
		}

		parsedCondition, diags := model.ParseSubjectPatternRuleCondition(context.Background(), rule.Condition)
		if diags.HasError() {
			continue
		}

		var parsed model.SubjectPatternCondition
		if diags := parsedCondition.As(context.Background(), &parsed, basetypes.ObjectAsOptions{}); diags.HasError() {
			continue
		}

		if parsed.Variable.ValueString() != condition.Variable.ValueString() ||
			parsed.Operator.ValueString() != condition.Operator.ValueString() ||
			parsed.Value.ValueString() != condition.Value.ValueString() {
			continue
		}

		return &rule, i, nil
	}

	return nil, -1, fmt.Errorf("subject pattern rule not found")
}
