package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-cloudconnector/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var _ resource.Resource = &SystemMappingResource{}
var _ resource.ResourceWithImportState = &SystemMappingResource{}

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
		MarkdownDescription: "Cloud Connector System Mapping Resource",
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
					},
				},
			},
			"system_mapping": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"virtual_host": schema.StringAttribute{
						MarkdownDescription: "Virtual host used on the cloud side. Cannot be updated after creation.",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"virtual_port": schema.StringAttribute{
						MarkdownDescription: "Virtual port used on the cloud side.",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"local_host": schema.StringAttribute{
						MarkdownDescription: "Host on the on-premise side.",
						Required:            true,
					},
					"local_port": schema.StringAttribute{
						MarkdownDescription: "Port on the on-premise side.",
						Required:            true,
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
						Required: true,
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
						Required: true,
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
						Required: true,
					},
					"host_in_header": schema.StringAttribute{
						MarkdownDescription: "Policy for setting the host in the response header. This property is applicable to HTTP(S) protocols only.",
						Required:            true,
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
						MarkdownDescription: "SAP router route, required only if an SAP router is used.",
						Computed:            true,
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *SystemMappingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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
	var plan SystemMappingData
	var respObj apiobjects.SystemMapping
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := plan.Credentials.RegionHost.ValueString()
	subaccount := plan.Credentials.Subaccount.ValueString()
	virtual_host := plan.SystemMapping.VirtualHost.ValueString()
	virtual_port := plan.SystemMapping.VirtualPort.ValueString()
	endpoint := endpoints.GetSystemMappingBaseEndpoint(region_host, subaccount)

	planBody := map[string]string{
		"virtualHost":        plan.SystemMapping.VirtualHost.ValueString(),
		"virtualPort":        plan.SystemMapping.VirtualPort.ValueString(),
		"localHost":          plan.SystemMapping.LocalHost.ValueString(),
		"localPort":          plan.SystemMapping.LocalPort.ValueString(),
		"protocol":           plan.SystemMapping.Protocol.ValueString(),
		"backendType":        plan.SystemMapping.BackendType.ValueString(),
		"authenticationMode": plan.SystemMapping.AuthenticationMode.ValueString(),
		"hostInHeader":       plan.SystemMapping.HostInHeader.ValueString(),
		"sid":                plan.SystemMapping.Sid.ValueString(),
		"description":        plan.SystemMapping.Description.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error creating the cloud connector system mapping", err.Error())
		return
	}

	endpoint = endpoints.GetSystemMappingEndpoint(region_host, subaccount, virtual_host, virtual_port)

	err = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector system mapping", err.Error())
		return
	}

	responseModel, err := SystemMappingValueFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *SystemMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemMappingData
	var respObj apiobjects.SystemMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := state.Credentials.RegionHost.ValueString()
	subaccount := state.Credentials.Subaccount.ValueString()
	virtual_host := state.SystemMapping.VirtualHost.ValueString()
	virtual_port := state.SystemMapping.VirtualPort.ValueString()
	endpoint := endpoints.GetSystemMappingEndpoint(region_host, subaccount, virtual_host, virtual_port)

	err := requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector system mapping", err.Error())
		return
	}

	responseModel, err := SystemMappingValueFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SystemMappingData
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

	region_host := plan.Credentials.RegionHost.ValueString()
	subaccount := plan.Credentials.Subaccount.ValueString()
	virtual_host := state.SystemMapping.VirtualHost.ValueString()
	virtual_port := state.SystemMapping.VirtualPort.ValueString()
	endpoint := endpoints.GetSystemMappingEndpoint(region_host, subaccount, virtual_host, virtual_port)

	planBody := map[string]string{
		"virtualHost":        plan.SystemMapping.VirtualHost.ValueString(),
		"virtualPort":        plan.SystemMapping.VirtualPort.ValueString(),
		"localHost":          plan.SystemMapping.LocalHost.ValueString(),
		"localPort":          plan.SystemMapping.LocalPort.ValueString(),
		"protocol":           plan.SystemMapping.Protocol.ValueString(),
		"backendType":        plan.SystemMapping.BackendType.ValueString(),
		"authenticationMode": plan.SystemMapping.AuthenticationMode.ValueString(),
		"hostInHeader":       plan.SystemMapping.HostInHeader.ValueString(),
		"sid":                plan.SystemMapping.Sid.ValueString(),
		"description":        plan.SystemMapping.Description.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error updating the cloud connector system mapping", err.Error())
		return
	}

	err = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector system mapping", err.Error())
		return
	}

	responseModel, err := SystemMappingValueFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemMappingData
	var respObj apiobjects.SystemMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := state.Credentials.RegionHost.ValueString()
	subaccount := state.Credentials.Subaccount.ValueString()
	virtual_host := state.SystemMapping.VirtualHost.ValueString()
	virtual_port := state.SystemMapping.VirtualPort.ValueString()
	endpoint := fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/systemMappings/%s:%s", region_host, subaccount, virtual_host, virtual_port)

	err := requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError("error deleting the system mapping", err.Error())
		return
	}

	responseModel, err := SystemMappingValueFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
