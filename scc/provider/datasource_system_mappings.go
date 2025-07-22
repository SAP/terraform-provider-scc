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
		MarkdownDescription: `Cloud Connector System Mappings Data Source.
				
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator
	* Display
	* Support

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/system-mappings>`,
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
								getFormattedValueAsTableRow("HTTP", "HTTP protocol") +
								getFormattedValueAsTableRow("HTTPS", "Secure HTTP protocol") +
								getFormattedValueAsTableRow("RFC", "Remote Function Call protocol") +
								getFormattedValueAsTableRow("RFCS", "Secure RFC protocol") +
								getFormattedValueAsTableRow("LDAP", "Lightweight Directory Access Protocol") +
								getFormattedValueAsTableRow("LDAPS", "Secure LDAP") +
								getFormattedValueAsTableRow("TCP", "Transmission Control Protocol") +
								getFormattedValueAsTableRow("TCPS", "Secure TCP"),
							Computed: true,
						},
						"backend_type": schema.StringAttribute{
							MarkdownDescription: "Type of the backend system. Valid values are:" +
								getFormattedValueAsTableRow("backend", "description") +
								getFormattedValueAsTableRow("---", "---") +
								getFormattedValueAsTableRow("abapSys", "ABAP-based SAP system") +
								getFormattedValueAsTableRow("netweaverCE", "SAP NetWeaver Composition Environment") +
								getFormattedValueAsTableRow("netweaverGW", "SAP NetWeaver Gateway") +
								getFormattedValueAsTableRow("applServerJava", "Java-based application server") +
								getFormattedValueAsTableRow("PI", "SAP Process Integration system") +
								getFormattedValueAsTableRow("hana", "SAP HANA system") +
								getFormattedValueAsTableRow("otherSAPsys", "Other SAP system") +
								getFormattedValueAsTableRow("nonSAPsys", "Non-SAP system"),
							Computed: true,
						},
						"authentication_mode": schema.StringAttribute{
							MarkdownDescription: "Authentication mode to be used on the backend side, which must be one of the following:" +
								getFormattedValueAsTableRow("authentication mode", "description") +
								getFormattedValueAsTableRow("---", "---") +
								getFormattedValueAsTableRow("NONE", "No authentication") +
								getFormattedValueAsTableRow("NONE_RESTRICTED", "No authentication; system certificate will never be sent") +
								getFormattedValueAsTableRow("X509_GENERAL", "X.509 certificate-based authentication, system certificate may be sent") +
								getFormattedValueAsTableRow("X509_RESTRICTED", "X.509 certificate-based authentication, system certificate never sent") +
								getFormattedValueAsTableRow("KERBEROS", "Kerberos-based authentication") +
								"The authentication modes NONE_RESTRICTED and X509_RESTRICTED prevent the Cloud Connector from sending the system certificate in any case, whereas NONE and X509_GENERAL will send the system certificate if the circumstances allow it.",
							Computed: true,
						},
						"host_in_header": schema.StringAttribute{
							MarkdownDescription: "Policy for setting the host in the response header. This property is applicable to HTTP(S) protocols only. If set, it must be one of the following strings:" +
								getFormattedValueAsTableRow("policy", "description") +
								getFormattedValueAsTableRow("---", "---") +
								getFormattedValueAsTableRow("internal/INTERNAL", "Use internal (local) host for HTTP headers") +
								getFormattedValueAsTableRow("virtual/VIRTUAL", "Use virtual host (default) for HTTP headers") + "The default is virtual.",
							Computed: true,
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
	var data SystemMappingsConfig
	var respObj apiobjects.SystemMappings
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()
	endpoint := endpoints.GetSystemMappingBaseEndpoint(regionHost, subaccount)

	err := requestAndUnmarshal(d.client, &respObj.SystemMappings, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSystemMappingsFailed, err.Error())
		return
	}

	responseModel, err := SystemMappingsValueFrom(ctx, data, respObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapSystemMappingsFailed, fmt.Sprintf("%s", err))
		return
	}
	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
