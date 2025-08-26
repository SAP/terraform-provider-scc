package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &SubaccountABAPServiceChannelsDataSource{}

func NewSubaccountABAPServiceChannelsDataSource() datasource.DataSource {
	return &SubaccountABAPServiceChannelsDataSource{}
}

type SubaccountABAPServiceChannelsDataSource struct {
	client *api.RestApiClient
}

func (d *SubaccountABAPServiceChannelsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_abap_service_channels"
}

func (r *SubaccountABAPServiceChannelsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount ABAP Service Channels Data Source.
				
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator
	* Display
	* Support

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/subaccount-service-channels>`,
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
			"subaccount_abap_service_channels": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"abap_cloud_tenant_host": schema.StringAttribute{
							MarkdownDescription: "Host name to access the Host of ABAP Cloud Tenant.",
							Computed:            true,
						},

						"instance_number": schema.Int64Attribute{
							MarkdownDescription: "Local Instance number under which the ABAP Cloud system is reachable for the client systems.",
							Computed:            true,
							Validators: []validator.Int64{
								int64validator.Between(00, 99),
							},
						},
						"id": schema.Int64Attribute{
							MarkdownDescription: "Unique identifier for the subaccount service channel (a positive integer number, starting with 1). This identifier is unique across all types of subaccount service channels.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of Subaccount Service Channel.",
							Computed:            true,
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "Port of the subaccount service channel for the ABAP Cloud System. The port numbers result from the following pattern: `33<LocalInstanceNumber>`, for activated SNC (Secure Network Connection) `48<LocalInstanceNumber>`.",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Boolean flag indicating whether the channel is enabled and therefore should be open.",
							Computed:            true,
						},
						"connections": schema.Int64Attribute{
							MarkdownDescription: "Maximal number of open connections.",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment or short description; this property is not supplied if no comment was provided.",
							Computed:            true,
						},
						"state": schema.SingleNestedAttribute{
							MarkdownDescription: "Current connection state; this property is only available if the channel is enabled.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"connected": schema.BoolAttribute{
									MarkdownDescription: "A Boolean flag indicating whether the channel is connected.",
									Computed:            true,
								},
								"opened_connections": schema.Int64Attribute{
									MarkdownDescription: "The number of open, possibly idle connections.",
									Computed:            true,
								},
								"connected_since_time_stamp": schema.Int64Attribute{
									MarkdownDescription: "The time stamp, a UTC long number, for the first time the channel was opened/connected.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *SubaccountABAPServiceChannelsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubaccountABAPServiceChannelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubaccountABAPServiceChannelsConfig
	var respObj apiobjects.SubaccountABAPServiceChannels
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()

	endpoint := endpoints.GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, "ABAPCloud")

	diags = requestAndUnmarshal(d.client, &respObj.SubaccountABAPServiceChannels, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountABAPServiceChannelsValueFrom(ctx, data, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
