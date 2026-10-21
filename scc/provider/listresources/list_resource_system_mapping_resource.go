package listresources

import (
	"context"
	"fmt"
	"regexp"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ list.ListResourceWithConfigure = &SystemMappingResourceListResource{}

type SystemMappingResourceListResource struct {
	Client *api.RestApiClient
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

	r.Client = client
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
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
			},
			"virtual_host": schema.StringAttribute{
				MarkdownDescription: "The virtual host name used in the system mapping.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9.-]+$`),
						"must contain only lowercase letters, numbers, dots, and dashes",
					),
				},
			},
			"virtual_port": schema.StringAttribute{
				MarkdownDescription: "The virtual port used in the system mapping.",
				Required:            true,
				Validators: []validator.String{
					// Validate port format without requiring a "protocol" context.
					// The protocol-aware PortValidator reads protocol from config, which
					// list resource configs do not have.
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d{1,5}$|^sapms[A-Z0-9]{3}$|^sapgw\d{2}s?$`),
						"virtual_port must be a numeric port (1-65535), sapmsSID, sapgwXX, or sapgwXXs",
					),
				},
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
		filter   model.SystemMappingResourceListResourceFilterModel
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

	diags := helpers.RequestAndUnmarshal(r.Client, &respObj.SystemMappingResources, "GET", endpoint, nil, true)
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
				resDm, dgs := model.MapToSystemMappingResourceListModel(ctx, filter, sm)
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
