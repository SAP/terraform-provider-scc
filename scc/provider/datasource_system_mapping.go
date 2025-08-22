package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/systemMapping"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SystemMappingDataSource{}

func NewSystemMappingDataSource() datasource.DataSource {
	return &SystemMappingDataSource{}
}

type SystemMappingDataSource struct {
	client *api.RestApiClient
}

func (d *SystemMappingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping"
}

func (r *SystemMappingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector System Mapping Data Source.
				
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
			"virtual_host": schema.StringAttribute{
				MarkdownDescription: `Virtual host used on the cloud side.
				Cannot be updated after creation (changing it requires a resource replacement).
				Host names with underscore ('_') may cause problems. We recommend refraining from using underscore in host names.
				
Note: In the UI, this attribute may appear with different names depending on the protocol used:
* **HTTP(S), TCP, LDAP** → "Virtual Host"
* **RFC** → "Virtual Message Server/ Virtual Application Server"`,
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9.-]+$`),
						"must contain only letters, numbers, dots, and dashes (no underscores recommended)",
					),
				},
			},
			"virtual_port": schema.StringAttribute{
				MarkdownDescription: `Port on the cloud (virtual) side.  
Cannot be updated after creation (changing this value requires resource replacement).

__UI Note:__ This attribute appears under different names depending on the protocol:
* **HTTP(S), TCP, LDAP** → "Virtual Port"
* **RFC** → "Virtual Instance Number/ Virtual System ID"

__Allowed formats:__
* **Numeric (0–65535)** → for HTTP(S), TCP/TCPS, LDAP/LDAPS
* **sapgwXX** or **sapgwXXs** → for RFC without load balancing
* **33XX** → Classic RFC Port
* **48XX** → Secure RFC Port`,
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([0-9]{1,5}|sapgw[0-9]{2}|sapgw[0-9]{2}s|33[0-9]{2}|48[0-9]{2})$`),
						"must be numeric (0–65535), sapgwXX, sapgwXXs, 33XX, or 48XX",
					),
				},
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: `Host on the on-premise side.
				Host names with underscore ('_') may cause problems. We recommend refraining from using underscore in host names.
Note: In the UI, this attribute may appear with different names depending on the protocol used:
* **HTTP(S), TCP, LDAP** → "Internal Host"
* **RFC** → "Message Server/ Application Server"`,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9.-]+$`),
						"must contain only letters, numbers, dots, and dashes (no underscores recommended)",
					),
				},
			},
			"internal_port": schema.StringAttribute{
				MarkdownDescription: `Port on the on-premise side.
__UI Note:__ This field may appear under different names in the Cloud Connector UI depending on the protocol:
* **HTTP(S), TCP, LDAP** → "Internal Port / Port Range"
* **RFC** → "System ID / Instance Number"
				
				
__Allowed formats:__
* **Numeric (0–65535)** → for HTTP(S), TCP/TCPS, LDAP/LDAPS
* **sapgwXX** or **sapgwXXs** → for RFC without load balancing
* **33XX** → Classic RFC Port
* **48XX** → Secure RFC Port`,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([0-9]{1,5}|sapgw[0-9]{2}|sapgw[0-9]{2}s|33[0-9]{2}|48[0-9]{2})$`),
						"must be numeric (0–65535), sapgwXX, sapgwXXs, 33XX, or 48XX",
					),
				},
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
				Validators: []validator.String{
					stringvalidator.OneOf("HTTP", "HTTPS", "RFC", "RFCS", "LDAP", "LDAPS", "TCP", "TCPS"),
					systemMapping.ValidateProtocolBackend(),
				},
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
				Validators: []validator.String{
					stringvalidator.OneOf("abapSys", "netweaverCE", "netweaverGW", "applServerJava", "PI", "hana", "otherSAPsys", "nonSAPsys"),
				},
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
				Validators: []validator.String{
					stringvalidator.OneOf("NONE", "NONE_RESTRICTED", "X509_GENERAL", "X509_RESTRICTED", "KERBEROS"),
					systemMapping.ValidateAuthenticationMode(),
				},
			},
			"host_in_header": schema.StringAttribute{
				MarkdownDescription: "Policy for setting the host in the response header. This property is applicable to HTTP(S) protocols only. If set, it must be one of the following strings:" +
					getFormattedValueAsTableRow("policy", "description") +
					getFormattedValueAsTableRow("---", "---") +
					getFormattedValueAsTableRow("internal/INTERNAL", "Use internal (local) host for HTTP headers") +
					getFormattedValueAsTableRow("virtual/VIRTUAL", "Use virtual host (default) for HTTP headers") + "The default is virtual.",
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("internal", "INTERNAL", "virtual", "VIRTUAL"),
					systemMapping.ValidateProtocolString([]string{"HTTP", "HTTPS"}),
				},
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
				MarkdownDescription: `SAP router string (only applicable if an SAP router is used). Only applicable for RFC-based communication.
__Format rules:__
* Sequence of hops separated by */H/* and */S/*
* Each hop must contain a host and a port
* Host can be a hostname, FQDN, or IPv4
* Port must be numeric (0–65535)`,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(?:/H/([a-zA-Z0-9.-]+|(?:\d{1,3}\.){3}\d{1,3})/S/[0-9]{1,5})+$`),
						"must be a valid SAProuter string like /H/host/S/3299 or /H/host1/S/3299/H/host2/S/3299",
					),
					systemMapping.ValidateProtocolString([]string{"RFC", "RFCS"}),
				},
			},
			"snc_partner_name": schema.StringAttribute{
				MarkdownDescription: "Distinguished name of the SNC partner in the format 'p:<Distinguished_Name>' (RFCS only).",
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^p:.+`), "snc_partner_name must match syntax: 'p:<Distinguished_Name>'"),
					systemMapping.ValidateProtocolString([]string{"RFCS"}),
				},
			},
			"allowed_clients": schema.ListAttribute{
				MarkdownDescription: "List of allowed SAP clients (3 characters each). Only applicable for RFC-based communication.",
				ElementType:         types.StringType,
				Computed:            true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(3, 3),
					),
					systemMapping.ValidateProtocolList([]string{"RFC", "RFCS"}),
				},
			},
			"blacklisted_users": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of users that are not allowed to execute the call, even if the client is listed under allowed clients. If not specified, no users are blacklisted. Only applicable for RFC-based communication.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client": schema.StringAttribute{
							MarkdownDescription: "Client ID of the user (3 characters).",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(3, 3),
							},
						},
						"user": schema.StringAttribute{
							MarkdownDescription: "User ID of the user.",
							Required:            true,
						},
					},
				},
				Validators: []validator.List{
					systemMapping.ValidateProtocolList([]string{"RFC", "RFCS"}),
				},
			},
		},
	}
}

func (d *SystemMappingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SystemMappingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemMappingConfig
	var respObj apiobjects.SystemMapping
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()
	virtualHost := data.VirtualHost.ValueString()
	virtualPort := data.VirtualPort.ValueString()
	endpoint := endpoints.GetSystemMappingEndpoint(regionHost, subaccount, virtualHost, virtualPort)

	err := requestAndUnmarshal(d.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSystemMappingFailed, err.Error())
		return
	}

	responseModel, diags := SystemMappingValueFrom(ctx, data, respObj)
	if diags.HasError() {
		resp.Diagnostics.AddError(errMsgMapSystemMappingFailed, fmt.Sprintf("%s", diags))
		return
	}
	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
