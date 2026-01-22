package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &SubaccountListResource{}

type SubaccountListResource struct {
	client *api.RestApiClient
}

type subaccountListFilterModel struct {
	RegionHost types.String `tfsdk:"region_host"`
}

func NewSubaccountListResource() list.ListResource {
	return &SubaccountListResource{}
}

func (r *SubaccountListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount" // must match managed resource
}

func (r *SubaccountListResource) Configure(ctx context.Context,
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

func (r *SubaccountListResource) ListResourceConfigSchema(
	ctx context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
SAP Cloud Connector **Subaccounts** list resource.

This list resource retrieves all subaccounts accessible via the configured
SAP Cloud Connector instance.
`,
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: `
Filter subaccounts by region host.

**Note:** If this attribute is omitted or set to an empty value, subaccounts
from all regions are returned.
`,
			},
		},
	}
}

// List streams all subaccounts from the API
func (r *SubaccountListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var (
		respObj apiobjects.SubaccountsListResource
		filter  subaccountListFilterModel
	)

	// Read optional filters
	if req.Config.Raw.IsFullyKnown() && !req.Config.Raw.IsNull() {
		if diags := req.Config.Get(ctx, &filter); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	endpoint := endpoints.GetSubaccountBaseEndpoint()

	diags := requestAndUnmarshal(r.client, &respObj.Subaccounts, "GET", endpoint, nil, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		warned := false

		for _, sa := range respObj.Subaccounts {

			if !filter.RegionHost.IsNull() && filter.RegionHost.ValueString() != "" {
				if sa.RegionHost != filter.RegionHost.ValueString() {
					continue
				}
			}

			result := req.NewListResult(ctx)

			_ = result.Identity.SetAttribute(ctx, path.Root("subaccount"), types.StringValue(sa.Subaccount))
			_ = result.Identity.SetAttribute(ctx, path.Root("region_host"), types.StringValue(sa.RegionHost))

			if !warned && req.IncludeResource {
				result.Diagnostics.AddWarning(
					"include_resource Not Supported",
					"The include_resource option is not supported for this list resource.",
				)

				warned = true
			}

			if !push(result) {
				return
			}
		}
	}
}
