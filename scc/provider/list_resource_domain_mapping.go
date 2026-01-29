package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &DomainMappingListResource{}

type DomainMappingListResource struct {
	client *api.RestApiClient
}

type domainMappingListFilterModel struct {
	RegionHost types.String `tfsdk:"region_host"`
	Subaccount types.String `tfsdk:"subaccount"`
}

func NewDomainMappingListResource() list.ListResource {
	return &DomainMappingListResource{}
}

func (r *DomainMappingListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_mapping" // must match managed resource
}

func (r *DomainMappingListResource) Configure(ctx context.Context,
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

	r.client = client
}

// ListResourceConfigSchema defines the schema for the 'config' block in a list query.
func (r *DomainMappingListResource) ListResourceConfigSchema(
	ctx context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
SAP Cloud Connector **Domain mapping** list resource.

This list resource retrieves all domain mappings for a specific region_host and subaccount 
accessible via the configured SAP Cloud Connector instance. 
`,
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				MarkdownDescription: "The host URL of the region (e.g., cf.us10.hana.ondemand.com).",
				Required:            true,
			},
			"subaccount": schema.StringAttribute{
				MarkdownDescription: "The GUID of the SAP subaccount.",
				Required:            true,
			},
		},
	}
}

// List streams all domain mappings for a specific subaccount and region host from the API to the results stream.
func (r *DomainMappingListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var (
		respObj apiobjects.DomainMappings
		filter  domainMappingListFilterModel
	)

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	endpoint := fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/domainMappings", filter.RegionHost.ValueString(), filter.Subaccount.ValueString())

	diags := requestAndUnmarshal(r.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {

		for _, dm := range respObj.DomainMappings {

			result := req.NewListResult(ctx)

			_ = result.Identity.SetAttribute(ctx, path.Root("subaccount"), types.StringValue(filter.Subaccount.ValueString()))
			_ = result.Identity.SetAttribute(ctx, path.Root("region_host"), types.StringValue(filter.RegionHost.ValueString()))
			_ = result.Identity.SetAttribute(ctx, path.Root("internal_domain"), types.StringValue(dm.InternalDomain))

			resDm := &DomainMappingConfig{
				Subaccount:     filter.Subaccount,
				RegionHost:     filter.RegionHost,
				InternalDomain: types.StringValue(dm.InternalDomain),
				VirtualDomain:  types.StringValue(dm.VirtualDomain),
			}

			if req.IncludeResource {
				// Set the resource information on the result
				result.Diagnostics.Append(result.Resource.Set(ctx, resDm)...)
			}

			if !push(result) {
				return
			}
		}
	}
}
