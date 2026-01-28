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

var _ list.ListResourceWithConfigure = &SystemMappingResourceListResource{}

type SystemMappingResourceListResource struct {
	client *api.RestApiClient
}

type systemMappingResourceListResourceFilterModel struct {
	RegionHost  types.String `tfsdk:"region_host"`
	Subaccount  types.String `tfsdk:"subaccount"`
	VirtualHost types.String `tfsdk:"virtual_host"`
	VirtualPort types.String `tfsdk:"virtual_port"`
}

func NewSystemMappingResourceListResource() list.ListResource {
	return &SystemMappingResourceListResource{}
}

func (r *SystemMappingResourceListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping_resource" // must match managed resource
}

func (r *SystemMappingResourceListResource) Configure(ctx context.Context,
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

func (r *SystemMappingResourceListResource) ListResourceConfigSchema(
	ctx context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
SAP Cloud Connector **System Mapping Resource** list resource.
This list resource retrieves system mappings resource for a specific region host, subaccount, virtual_host and virtual_port.
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
			"virtual_host": schema.StringAttribute{
				MarkdownDescription: "The virtual host name used in the system mapping.",
				Required:            true,
			},
			"virtual_port": schema.StringAttribute{
				MarkdownDescription: "The virtual port used in the system mapping.",
				Required:            true,
			},
		},
	}
}

func (r *SystemMappingResourceListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var (
		respObj  apiobjects.SystemMappingResources
		filter   systemMappingResourceListResourceFilterModel
		endpoint string
	)

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	endpoint = endpoints.GetSystemMappingResourceBaseEndpoint(
		filter.RegionHost.ValueString(),
		filter.Subaccount.ValueString(),
		filter.VirtualHost.ValueString(),
		filter.VirtualPort.ValueString(),
	)

	diags := requestAndUnmarshal(r.client, &respObj.SystemMappingResources, "GET", endpoint, nil, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// 4. Stream Results
	stream.Results = func(push func(list.ListResult) bool) {
		for _, sm := range respObj.SystemMappingResources {
			result := req.NewListResult(ctx)

			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("subaccount"), filter.Subaccount)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("region_host"), filter.RegionHost)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("virtual_host"), filter.VirtualHost)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("virtual_port"), filter.VirtualPort)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("url_path"), types.StringValue(sm.URLPath))...)

			if req.IncludeResource {
				resDm, dgs := MapToSystemMappingResourceListModel(ctx, filter, sm)
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
