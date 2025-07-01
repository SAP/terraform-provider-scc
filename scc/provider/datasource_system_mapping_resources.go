package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &SystemMappingResourcesDataSource{}

func NewSystemMappingResourcesDataSource() datasource.DataSource {
	return &SystemMappingResourcesDataSource{}
}

type SystemMappingResourcesDataSource struct {
	client *api.RestApiClient
}

func (d *SystemMappingResourcesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping_resources"
}

func (r *SystemMappingResourcesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector System Mapping Resources Data Source.
				
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator
	* Display
	* Support

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/system-mapping-resources>`,
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				MarkdownDescription: "Region Host Name.",
				Required:            true,
			},
			"subaccount": schema.StringAttribute{
				MarkdownDescription: "The ID of the subaccount.",
				Required:            true,
			},
			"virtual_host": schema.StringAttribute{
				MarkdownDescription: "Virtual host used on the cloud side.",
				Required:            true,
			},
			"virtual_port": schema.StringAttribute{
				MarkdownDescription: "Virtual port used on the cloud side.",
				Required:            true,
			},
			"system_mapping_resources": schema.ListNestedAttribute{
				MarkdownDescription: "A list of system mapping resource. ",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The resource itself, which, depending on the owning system mapping, is either a URL path (or the leading section of it), or a RFC function name.",
							Required:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Boolean flag indicating whether the resource is enabled.",
							Computed:            true,
						},
						"exact_match_only": schema.BoolAttribute{
							MarkdownDescription: "Boolean flag determining whether access is granted only if the requested resource is an exact match.",
							Computed:            true,
						},
						"websocket_upgrade_allowed": schema.BoolAttribute{
							MarkdownDescription: "Boolean flag indicating whether websocket upgrade is allowed. This property is of relevance only if the owning system mapping employs protocol HTTP or HTTPS.",
							Computed:            true,
						},
						"creation_date": schema.StringAttribute{
							MarkdownDescription: "Date of creation of system mapping resource.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description. This property is not available unless explicitly set.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SystemMappingResourcesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *SystemMappingResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemMappingResourcesConfig
	var respObj apiobjects.SystemMappingResources
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()
	virtualHost := data.VirtualHost.ValueString()
	virtualPort := data.VirtualPort.ValueString()
	endpoint := endpoints.GetSystemMappingResourceBaseEndpoint(regionHost, subaccount, virtualHost, virtualPort)

	err := requestAndUnmarshal(d.client, &respObj.SystemMappingResources, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSystemMappingResourcesFailed, err.Error())
		return
	}

	responseModel, err := SystemMappingResourcesValueFrom(ctx, data, respObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapSystemMappingResourcesFailed, fmt.Sprintf("%s", err))
		return
	}
	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
