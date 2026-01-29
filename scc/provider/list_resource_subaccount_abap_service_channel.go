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

var _ list.ListResourceWithConfigure = &SubaccountABAPServiceChannelListResource{}

type SubaccountABAPServiceChannelListResource struct {
	client *api.RestApiClient
}

type subaccountABAPServiceChannelListResourceFilterModel struct {
	RegionHost types.String `tfsdk:"region_host"`
	Subaccount types.String `tfsdk:"subaccount"`
}

func NewSubaccountABAPServiceChannelListResource() list.ListResource {
	return &SubaccountABAPServiceChannelListResource{}
}

func (r *SubaccountABAPServiceChannelListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_abap_service_channel" // must match managed resource
}

func (r *SubaccountABAPServiceChannelListResource) Configure(ctx context.Context,
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

func (r *SubaccountABAPServiceChannelListResource) ListResourceConfigSchema(
	ctx context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
SAP Cloud Connector **Subaccount ABAP service channel** list resource.

This list resource retrieves Subaccount ABAP service channel for a specific region host and subaccount.
`,
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				MarkdownDescription: "The host URL of the region (e.g., `cf.eu12.hana.ondemand.com`).",
				Required:            true,
			},
			"subaccount": schema.StringAttribute{
				MarkdownDescription: "The GUID of the SAP subaccount.",
				Required:            true,
			},
		},
	}
}

func (r *SubaccountABAPServiceChannelListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var (
		respObj  apiobjects.SubaccountABAPServiceChannels
		filter   subaccountABAPServiceChannelListResourceFilterModel
		endpoint string
	)

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	endpoint = endpoints.GetSubaccountServiceChannelBaseEndpoint(filter.RegionHost.ValueString(), filter.Subaccount.ValueString(), "ABAPCloud")

	diags := requestAndUnmarshal(r.client, &respObj.SubaccountABAPServiceChannels, "GET", endpoint, nil, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// 4. Stream Results
	stream.Results = func(push func(list.ListResult) bool) {
		for _, sm := range respObj.SubaccountABAPServiceChannels {
			result := req.NewListResult(ctx)

			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("subaccount"), filter.Subaccount)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("region_host"), filter.RegionHost)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("id"), types.Int64Value(sm.ID))...)

			if req.IncludeResource {
				resDm, dgs := SubaccountABAPServiceChannelListValueFrom(ctx, filter, sm)
				result.Diagnostics.Append(dgs...)
				if !dgs.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, resDm)...)
				}
			}

			if !push(result) {
				return
			}
		}
	}
}
