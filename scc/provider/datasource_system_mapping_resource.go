package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &SystemMappingResourceDataSource{}

func NewSystemMappingResourceDataSource() datasource.DataSource {
	return &SystemMappingResourceDataSource{}
}

type SystemMappingResourceDataSource struct {
	client *api.RestApiClient
}

func (d *SystemMappingResourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping_resource"
}

func (r *SystemMappingResourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector System Mapping Resource Data Source.
				
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
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
			},
			"virtual_host": schema.StringAttribute{
				MarkdownDescription: "Virtual host used on the cloud side.",
				Required:            true,
			},
			"virtual_port": schema.StringAttribute{
				MarkdownDescription: "Virtual port used on the cloud side.",
				Required:            true,
			},
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
	}
}

func (d *SystemMappingResourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SystemMappingResourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()
	virtualHost := data.VirtualHost.ValueString()
	virtualPort := data.VirtualPort.ValueString()
	resourceID := CreateEncodedResourceID(data.ID.ValueString())

	endpoint := endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resourceID)

	err := requestAndUnmarshal(d.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSystemMappingResourceFailed, err.Error())
		return
	}

	responseModel, err := SystemMappingResourceValueFrom(ctx, data, respObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapSystemMappingResourceFailed, fmt.Sprintf("%s", err))
		return
	}
	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
