package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &SystemMappingListResource{}

type SystemMappingListResource struct {
	client *api.RestApiClient
}

type systemMappingListResourceFilterModel struct {
	RegionHost  types.String `tfsdk:"region_host"`
	Subaccount  types.String `tfsdk:"subaccount"`
	VirtualHost types.String `tfsdk:"virtual_host"`
	VirtualPort types.String `tfsdk:"virtual_port"`
}

func NewSystemMappingListResource() list.ListResource {
	return &SystemMappingListResource{}
}

func (r *SystemMappingListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping" // must match managed resource
}

func (r *SystemMappingListResource) Configure(ctx context.Context,
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

func (r *SystemMappingListResource) ListResourceConfigSchema(
	ctx context.Context,
	req list.ListResourceSchemaRequest,
	resp *list.ListResourceSchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
SAP Cloud Connector **System Mapping** list resource.
This list resource retrieves system mappings for a specific region host and subaccount. 
It can be filtered to a specific system by providing both the virtual host and virtual port.
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
				MarkdownDescription: "The virtual host name used in the system mapping. If provided, `virtual_port` must also be specified.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("virtual_port")),
				},
			},
			"virtual_port": schema.StringAttribute{
				MarkdownDescription: "The virtual port used in the system mapping. If provided, `virtual_host` must also be specified.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("virtual_host")),
				},
			},
		},
	}
}

func (r *SystemMappingListResource) List(
	ctx context.Context,
	req list.ListRequest,
	stream *list.ListResultsStream,
) {
	var (
		respObj  apiobjects.SystemMappings
		filter   systemMappingListResourceFilterModel
		endpoint string
	)

	if diags := req.Config.Get(ctx, &filter); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	if !filter.VirtualHost.IsNull() && !filter.VirtualPort.IsNull() {
		endpoint = endpoints.GetSystemMappingEndpoint(
			filter.RegionHost.ValueString(),
			filter.Subaccount.ValueString(),
			filter.VirtualHost.ValueString(),
			filter.VirtualPort.ValueString(),
		)
	} else {
		endpoint = endpoints.GetSystemMappingBaseEndpoint(
			filter.RegionHost.ValueString(),
			filter.Subaccount.ValueString(),
		)
	}

	diags := requestAndUnmarshal(r.client, &respObj.SystemMappings, "GET", endpoint, nil, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// 4. Stream Results
	stream.Results = func(push func(list.ListResult) bool) {
		for _, sm := range respObj.SystemMappings {
			result := req.NewListResult(ctx)

			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("subaccount"), filter.Subaccount)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("region_host"), filter.RegionHost)...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("virtual_host"), types.StringValue(sm.VirtualHost))...)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("virtual_port"), types.StringValue(sm.VirtualPort))...)

			if req.IncludeResource {
				resDm, dgs := MapToSystemMappingListModel(ctx, filter, sm)
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
