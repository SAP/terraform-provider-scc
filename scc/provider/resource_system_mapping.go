package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/systemMapping"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	actionCreate = "Create"
	actionUpdate = "Update"
)

var _ resource.Resource = &SystemMappingResource{}

func NewSystemMappingResource() resource.Resource {
	return &SystemMappingResource{}
}

type SystemMappingResource struct {
	client *api.RestApiClient
}

func (r *SystemMappingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping"
}

func (r *SystemMappingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector System Mapping Resource.
				
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
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
				Required: true,
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
				Required: true,
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
				Required: true,
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
				Required: true,
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
				Optional: true,
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
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("internal", "INTERNAL", "virtual", "VIRTUAL"),
					systemMapping.ValidateProtocolString([]string{"HTTP", "HTTPS"}),
				},
			},
			"sid": schema.StringAttribute{
				MarkdownDescription: "The ID of the system.",
				Computed:            true,
				Optional:            true,
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
				Optional:            true,
			},
			"sap_router": schema.StringAttribute{
				MarkdownDescription: `SAP router string (only applicable if an SAP router is used). Only applicable for RFC-based communication.
__Format rules:__
* Sequence of hops separated by */H/* and */S/*
* Each hop must contain a host and a port
* Host can be a hostname, FQDN, or IPv4
* Port must be numeric (0–65535)`,
				Computed: true,
				Optional: true,
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
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^p:.+`), "snc_partner_name must match syntax: 'p:<Distinguished_Name>'"),
					systemMapping.ValidateProtocolString([]string{"RFCS"}),
				},
			},
			"allowed_clients": schema.ListAttribute{
				MarkdownDescription: "List of allowed SAP clients (3 characters each). Only applicable for RFC-based communication.",
				ElementType:         types.StringType,
				Computed:            true,
				Optional:            true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(3, 3),
					),
					systemMapping.ValidateProtocolList([]string{"RFC", "RFCS"}),
				},
			},
			"blacklisted_users": schema.ListNestedAttribute{
				Optional:            true,
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

func (r *SystemMappingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *SystemMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SystemMappingConfig
	var respObj apiobjects.SystemMapping
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	virtualHost := plan.VirtualHost.ValueString()
	virtualPort := plan.VirtualPort.ValueString()
	endpoint := endpoints.GetSystemMappingBaseEndpoint(regionHost, subaccount)

	planBody := buildSystemMappingBody(ctx, actionCreate, plan)

	diags = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint = endpoints.GetSystemMappingEndpoint(regionHost, subaccount, virtualHost, virtualPort)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemMappingConfig
	var respObj apiobjects.SystemMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	virtualHost := state.VirtualHost.ValueString()
	virtualPort := state.VirtualPort.ValueString()
	endpoint := endpoints.GetSystemMappingEndpoint(regionHost, subaccount, virtualHost, virtualPort)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingValueFrom(ctx, state, respObj)
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

func (r *SystemMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SystemMappingConfig
	var respObj apiobjects.SystemMapping
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	virtualHost := plan.VirtualHost.ValueString()
	virtualPort := plan.VirtualPort.ValueString()

	if (regionHost != state.RegionHost.ValueString()) ||
		(subaccount != state.Subaccount.ValueString()) ||
		(virtualHost != state.VirtualHost.ValueString()) ||
		(virtualPort != state.VirtualPort.ValueString()) {
		resp.Diagnostics.AddError("Error updating the cloud connector system mapping", "Failed to update the cloud connector system mapping due to mismatched configuration values.")
		return
	}
	endpoint := endpoints.GetSystemMappingEndpoint(regionHost, subaccount, virtualHost, virtualPort)

	planBody := buildSystemMappingBody(ctx, actionUpdate, plan)

	diags = requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemMappingConfig
	var respObj apiobjects.SystemMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	virtualHost := state.VirtualHost.ValueString()
	virtualPort := state.VirtualPort.ValueString()
	endpoint := fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/systemMappings/%s:%s", regionHost, subaccount, virtualHost, virtualPort)

	diags = requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingValueFrom(ctx, state, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func buildSystemMappingBody(ctx context.Context, action string, plan SystemMappingConfig) map[string]any {
	planBody := map[string]any{
		"virtualHost":        plan.VirtualHost.ValueString(),
		"virtualPort":        plan.VirtualPort.ValueString(),
		"localHost":          plan.InternalHost.ValueString(),
		"localPort":          plan.InternalPort.ValueString(),
		"protocol":           plan.Protocol.ValueString(),
		"backendType":        plan.BackendType.ValueString(),
		"authenticationMode": plan.AuthenticationMode.ValueString(),
	}

	// Optional fields (only if user provided them)
	addIfSet := func(attr types.String, key string) {
		if !attr.IsNull() && !attr.IsUnknown() {
			planBody[key] = attr.ValueString()
		}
	}

	addIfSet(plan.Description, "description")
	addIfSet(plan.HostInHeader, "hostInHeader")
	addIfSet(plan.Sid, "sid")
	addIfSet(plan.SAPRouter, "sapRouter")
	addIfSet(plan.SNCPartnerName, "sncPartnerName")

	// Handle case sensitive host in header values
	if !plan.HostInHeader.IsNull() && !plan.HostInHeader.IsUnknown() {
		planBody["hostInHeader"] = strings.ToUpper(plan.HostInHeader.ValueString())
	}

	// Handle allowed clients and blacklisted users (Update only)
	if action == actionUpdate {
		if !plan.AllowedClients.IsNull() && !plan.AllowedClients.IsUnknown() {
			var allowedClients []string
			plan.AllowedClients.ElementsAs(ctx, &allowedClients, false)
			planBody["allowedClients"] = allowedClients
		}

		if !plan.BlacklistedUsers.IsNull() && !plan.BlacklistedUsers.IsUnknown() {
			var blacklistedUsers []SystemMappingBlacklistedUsersData
			plan.BlacklistedUsers.ElementsAs(ctx, &blacklistedUsers, false)

			users := []map[string]string{}
			for _, u := range blacklistedUsers {
				users = append(users, map[string]string{
					"client": u.Client.ValueString(),
					"user":   u.User.ValueString(),
				})
			}
			planBody["blacklistedUsers"] = users
		}
	}
	return planBody
}

func (rs *SystemMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: region_host, subaccount, virtual_host, virtual_port. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_host"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_port"), idParts[3])...)
}
