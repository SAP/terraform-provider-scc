package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-cloudconnector/internal/api/endpoints"
	"github.com/SAP/terraform-provider-cloudconnector/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &SubaccountServiceChannelK8SDataSource{}

func NewSubaccountServiceChannelK8SDataSource() datasource.DataSource {
	return &SubaccountServiceChannelK8SDataSource{}
}

type SubaccountServiceChannelK8SDataSource struct {
	client *api.RestApiClient
}

func (d *SubaccountServiceChannelK8SDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_service_channel_k8s"
}

func (r *SubaccountServiceChannelK8SDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cloud Connector Subaccount Subaccount Service Channel K8S Data Source",
		Attributes: map[string]schema.Attribute{
			"credentials": schema.SingleNestedAttribute{
				MarkdownDescription: "Input parameters required to configure the subaccount connected to cloud connector.",
				Required:            true,
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
				},
			},
			"subaccount_service_channel_k8s": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"k8s_cluster": schema.StringAttribute{
						MarkdownDescription: "Host name to access the Kubernetes cluster.",
						Computed:            true,
					},

					"k8s_service": schema.StringAttribute{
						MarkdownDescription: "Host name providiing the service inside of Kubernetes cluster.",
						Computed:            true,
					},
					"id": schema.Int64Attribute{
						MarkdownDescription: "Unique identifier for the subaccount service channel (a positive integer number, starting with 1). This identifier is unique across all types of subaccount service channels.",
						Required:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of Subaccount Service Channel.",
						Computed:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "Port of the subaccount service channel for the virtual machine.",
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
	}
}

func (d *SubaccountServiceChannelK8SDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubaccountServiceChannelK8SDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubaccountServiceChannelK8SConfig
	var respObj apiobjects.SubaccountServiceChannelK8S
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := data.SubaccountServiceChannelK8SCredentials.RegionHost.ValueString()
	subaccount := data.SubaccountServiceChannelK8SCredentials.Subaccount.ValueString()
	id := data.SubaccountServiceChannelK8SData.ID.ValueInt64()

	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(region_host, subaccount, "K8S", id)

	err := requestAndUnmarshal(d.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector subaccount service channel", err.Error())
		return
	}

	responseModel, err := SubaccountServiceChannelK8SValueFrom(ctx, data, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount service channel value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
