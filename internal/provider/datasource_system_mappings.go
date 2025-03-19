package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-cloudconnector/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &SystemMappingsDataSource{}

func NewSystemMappingsDataSource() datasource.DataSource {
	return &SystemMappingsDataSource{}
}

type SystemMappingsDataSource struct {
	client *api.RestApiClient
}

func (d *SystemMappingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mappings"
}

func (r *SystemMappingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cloud Connector System Mappings Data Source",
		Attributes: map[string]schema.Attribute{
			"credentials": schema.SingleNestedAttribute{
				MarkdownDescription: "Input parameters required to configure the subaccount connected to cloud connector.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"region_host": schema.StringAttribute{
						Required: true,
					},
					"subaccount": schema.StringAttribute{
						Required: true,
					},
				},
			},
			"system_mappings": schema.ListNestedAttribute{
				MarkdownDescription: "List of System Mappings between Virtual and Internal System.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"virtual_host": schema.StringAttribute{
							MarkdownDescription: "Virtual host used on the cloud side.",
							Required:            true,
						},
						"virtual_port": schema.StringAttribute{
							MarkdownDescription: "Virtual port used on the cloud side.",
							Required:            true,
						},
						"local_host": schema.StringAttribute{
							MarkdownDescription: "Host on the on-premise side.",
							Computed:            true,
						},
						"local_port": schema.StringAttribute{
							MarkdownDescription: "Port on the on-premise side.",
							Computed:            true,
						},
						"creation_date": schema.StringAttribute{
							MarkdownDescription: "Date of creation of system mapping.",
							Computed:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "Protocol used when sending requests and receiving responses, which must be one of the following values:" +
								getFormattedValueAsTableRow("protocol", "description") +
								getFormattedValueAsTableRow("---", "---") +
								getFormattedValueAsTableRow("HTTP", "") +
								getFormattedValueAsTableRow("HTTPS", "") +
								getFormattedValueAsTableRow("RFC", "") +
								getFormattedValueAsTableRow("RFCS", "") +
								getFormattedValueAsTableRow("LDAP", "") +
								getFormattedValueAsTableRow("LDAPS", "") +
								getFormattedValueAsTableRow("TCP", "") +
								getFormattedValueAsTableRow("TCPS", ""),
							Computed: true,
						},
						"backend_type": schema.StringAttribute{
							MarkdownDescription: "Type of the backend system. Valid values are:" +
								getFormattedValueAsTableRow("protocol", "description") +
								getFormattedValueAsTableRow("---", "---") +
								getFormattedValueAsTableRow("abapSys", "") +
								getFormattedValueAsTableRow("netweaverCE", "") +
								getFormattedValueAsTableRow("netweaverGW", "") +
								getFormattedValueAsTableRow("applServerJava", "") +
								getFormattedValueAsTableRow("PI", "") +
								getFormattedValueAsTableRow("hana", "") +
								getFormattedValueAsTableRow("otherSAPsys", "") +
								getFormattedValueAsTableRow("nonSAPsys", ""),
							Computed: true,
						},
						"authentication_mode": schema.StringAttribute{
							MarkdownDescription: "Authentication mode to be used on the backend side, which must be one of the following:" +
								getFormattedValueAsTableRow("protocol", "description") +
								getFormattedValueAsTableRow("---", "---") +
								getFormattedValueAsTableRow("NONE", "") +
								getFormattedValueAsTableRow("NONE_RESTRICTED", "") +
								getFormattedValueAsTableRow("X509_GENERAL", "") +
								getFormattedValueAsTableRow("X509_RESTRICTED", "") +
								getFormattedValueAsTableRow("KERBEROS", ""),
							Computed: true,
						},
						"host_in_header": schema.StringAttribute{
							MarkdownDescription: "Policy for setting the host in the response header. This property is applicable to HTTP(S) protocols only.",
							Computed:            true,
						},
						"sid": schema.StringAttribute{
							MarkdownDescription: "The ID of the system.",
							Computed:            true,
						},
						"total_resources_count": schema.Int64Attribute{
							MarkdownDescription: "The total number of resources.",
							Computed:            true,
						},
						"enabled_resources_count": schema.Int64Attribute{
							MarkdownDescription: "The number of enabled resources.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the system mapping.",
							Computed:            true,
						},
						"sap_router": schema.StringAttribute{
							MarkdownDescription: "SAP router route, required only if an SAP router is used.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SystemMappingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SystemMappingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemMappingsData
	var respObj apiobjects.SystemMappings
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := data.Credentials.RegionHost.ValueString()
	subaccount := data.Credentials.Subaccount.ValueString()
	endpoint := endpoints.GetSystemMappingBaseEndpoint(region_host, subaccount)

	err := requestAndUnmarshal(d.client, &respObj.SystemMappings, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector system mappings", err.Error())
		return
	}

	responseModel, err := SystemMappingsValueFrom(ctx, data, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mappings value", fmt.Sprintf("%s", err))
		return
	}
	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
