package listresources

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &SubjectPatternRuleListResource{}

type SubjectPatternRuleListResource struct {
	Client *api.RestApiClient
}

func NewSubjectPatternRuleListResource() list.ListResource {
	return &SubjectPatternRuleListResource{}
}

func (r *SubjectPatternRuleListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subject_pattern_rule" // must match managed resource
}

func (r *SubjectPatternRuleListResource) Configure(ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.RestApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.RestApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.Client = client
}

// ListResourceConfigSchema defines the schema for the 'config' block in a list query.
// The optional index filter narrows results to the single rule at that position.
func (r *SubjectPatternRuleListResource) ListResourceConfigSchema(
	ctx context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `SAP Cloud Connector **Subject Pattern Rule** list resource.

Retrieves all subject pattern rules configured in the SAP Cloud Connector instance.
Optionally filter to a single rule by its positional index.`,
		Attributes: map[string]schema.Attribute{
			"index": schema.Int64Attribute{
				MarkdownDescription: "Filter results to the rule at this positional index. Omit to return all rules.",
				Optional:            true,
			},
		},
	}
}

type subjectPatternRuleListFilterModel struct {
	Index types.Int64 `tfsdk:"index"`
}

// List streams subject pattern rules from the API to the results stream.
// When index is set in config, only the rule at that position is streamed.
func (r *SubjectPatternRuleListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var (
		respObj apiobjects.SubjectPatternRules
		filter  subjectPatternRuleListFilterModel
	)

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	diags := helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoints.GetSubjectPatternRulesBaseEndpoint(), nil, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for i, spr := range respObj {
			// If an index filter is set, skip rules that don't match.
			if !filter.Index.IsNull() && !filter.Index.IsUnknown() {
				if int64(i) != filter.Index.ValueInt64() {
					continue
				}
			}

			result := req.NewListResult(ctx)

			// Identity matches the managed resource's IdentitySchema: index (Int64).
			result.Diagnostics.Append(
				result.Identity.SetAttribute(ctx, path.Root("index"), types.Int64Value(int64(i)))...,
			)

			if req.IncludeResource && !result.Diagnostics.HasError() {
				sp := model.SubjectPattern{
					CommonName:       helpers.StringValueOrNull(spr.SubjectPattern.CommonName),
					Email:            helpers.StringValueOrNull(spr.SubjectPattern.Email),
					Locality:         helpers.StringValueOrNull(spr.SubjectPattern.Locality),
					OrganizationUnit: helpers.StringValueOrNull(spr.SubjectPattern.OrganizationUnit),
					Organization:     helpers.StringValueOrNull(spr.SubjectPattern.Organization),
					State:            helpers.StringValueOrNull(spr.SubjectPattern.State),
					Country:          helpers.StringValueOrNull(spr.SubjectPattern.Country),
				}

				subjectPattern, d := types.ObjectValueFrom(ctx, model.SubjectPatternType, sp)
				result.Diagnostics.Append(d...)

				condition, d := model.ParseSubjectPatternRuleCondition(ctx, spr.Condition)
				result.Diagnostics.Append(d...)

				if !result.Diagnostics.HasError() {
					resSpr := &model.SubjectPatternRuleConfig{
						Index: types.Int64Value(int64(i)),
						SubjectPatternRule: model.SubjectPatternRule{
							Description:    types.StringValue(spr.Description),
							SubjectPattern: subjectPattern,
							Condition:      condition,
						},
					}
					result.Diagnostics.Append(result.Resource.Set(ctx, resSpr)...)
				}
			}

			// Always push — errors on a result are surfaced by the framework.
			// Returning false from push means the caller wants no more results.
			if !push(result) {
				return
			}
		}
	}
}
