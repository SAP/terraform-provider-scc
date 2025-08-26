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

var _ datasource.DataSource = &SubaccountConfigurationDataSource{}

func NewSubaccountConfigurationDataSource() datasource.DataSource {
	return &SubaccountConfigurationDataSource{}
}

type SubaccountConfigurationDataSource struct {
	client *api.RestApiClient
}

func (d *SubaccountConfigurationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_configuration"
}

func (r *SubaccountConfigurationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount Configuration Data Source.
				
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator
	* Display
	* Support

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/subaccount>`,
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
			"location_id": schema.StringAttribute{
				MarkdownDescription: "Location identifier for the Cloud Connector instance. This property is not available if the default location ID is in use.",
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name of the subaccount.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the subaccount.",
				Computed:            true,
			},
			"tunnel": schema.SingleNestedAttribute{
				MarkdownDescription: "Array of connection tunnels used by the subaccount.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						MarkdownDescription: "State of the tunnel. Possible values are: \n" +
							getFormattedValueAsTableRow("state", "description") +
							getFormattedValueAsTableRow("---", "---") +
							getFormattedValueAsTableRow("`Connected`", "The tunnel is active and functioning properly.") +
							getFormattedValueAsTableRow("`ConnectFailure`", "The tunnel failed to establish a connection due to an issue.") +
							getFormattedValueAsTableRow("`Disconnected`", "The tunnel was previously connected but is now intentionally or unintentionally disconnected."),
						Computed: true,
					},
					"connected_since_time_stamp": schema.Int64Attribute{
						MarkdownDescription: "Timestamp of the start of the connection.",
						Computed:            true,
					},
					"connections": schema.Int64Attribute{
						MarkdownDescription: "Number of subaccount connections.",
						Computed:            true,
					},
					"subaccount_certificate": schema.SingleNestedAttribute{
						MarkdownDescription: "Information on the subaccount certificate such as validity period, issuer and subject DN.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"not_after_time_stamp": schema.Int64Attribute{
								MarkdownDescription: "Timestamp of the end of the validity period.",
								Computed:            true,
							},
							"not_before_time_stamp": schema.Int64Attribute{
								MarkdownDescription: "Timestamp of the beginning of the validity period.",
								Computed:            true,
							},
							"subject_dn": schema.StringAttribute{
								MarkdownDescription: "The subject distinguished name.",
								Computed:            true,
							},
							"issuer": schema.StringAttribute{
								MarkdownDescription: "Certificate authority (CA) that issued this certificate.",
								Computed:            true,
							},
							"serial_number": schema.StringAttribute{
								MarkdownDescription: "Unique identifier for the certificate, typically assigned by the CA.",
								Computed:            true,
							},
						},
					},
					"application_connections": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"connection_count": schema.Int64Attribute{
									MarkdownDescription: "Number of active connections to the specified application instance.",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of the connected application instance.",
									Computed:            true,
								},
								"type": schema.StringAttribute{
									MarkdownDescription: "Type of the connected application instance.",
									Computed:            true,
								},
							},
						},
					},
					"service_channels": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Type of the service channel (e.g., HANA, VM, or RFC).",
									Computed:            true,
								},
								"state": schema.StringAttribute{
									MarkdownDescription: "Current operational state of the service channel.",
									Computed:            true,
								},
								"details": schema.StringAttribute{
									MarkdownDescription: "Technical details about the service channel.",
									Computed:            true,
								},
								"comment": schema.StringAttribute{
									MarkdownDescription: "Optional user-provided comment or annotation regarding the service channel.",
									Computed:            true,
								},
							},
						},
					},
					"user": schema.StringAttribute{
						MarkdownDescription: "User for the specified region host and subaccount.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *SubaccountConfigurationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubaccountConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubaccountData
	var respObj apiobjects.Subaccount
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount)

	diags = requestAndUnmarshal(d.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountDataSourceValueFrom(ctx, respObj)
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
