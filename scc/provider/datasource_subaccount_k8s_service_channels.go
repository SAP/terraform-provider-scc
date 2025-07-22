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

var _ datasource.DataSource = &SubaccountK8SServiceChannelsDataSource{}

func NewSubaccountK8SServiceChannelsDataSource() datasource.DataSource {
	return &SubaccountK8SServiceChannelsDataSource{}
}

type SubaccountK8SServiceChannelsDataSource struct {
	client *api.RestApiClient
}

func (d *SubaccountK8SServiceChannelsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_k8s_service_channels"
}

func (r *SubaccountK8SServiceChannelsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount K8S Service Channels Data Source.
				
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
			"subaccount_k8s_service_channels": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"k8s_cluster": schema.StringAttribute{
							MarkdownDescription: "Host name to access the Kubernetes cluster.",
							Computed:            true,
						},

						"k8s_service": schema.StringAttribute{
							MarkdownDescription: "Host name providing the service inside of Kubernetes cluster.",
							Computed:            true,
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
							MarkdownDescription: "Port of the subaccount service channel for the Kubernetes Cluster.",
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

func (d *SubaccountK8SServiceChannelsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubaccountK8SServiceChannelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubaccountK8SServiceChannelsConfig
	var respObj apiobjects.SubaccountK8SServiceChannels
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()

	endpoint := endpoints.GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, "K8S")

	err := requestAndUnmarshal(d.client, &respObj.SubaccountK8SServiceChannels, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSubaccountK8SServiceChannelsFailed, err.Error())
		return
	}

	responseModel, diags := SubaccountK8SServiceChannelsValueFrom(ctx, data, respObj)
	if diags.HasError() {
		resp.Diagnostics.AddError(errMsgMapSubaccountK8SServiceChannelsFailed, fmt.Sprintf("%s", diags))
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
