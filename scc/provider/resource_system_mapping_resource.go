package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.Resource = &SystemMappingResourceResource{}

func NewSystemMappingResourceResource() resource.Resource {
	return &SystemMappingResourceResource{}
}

type SystemMappingResourceResource struct {
	client *api.RestApiClient
}

func (r *SystemMappingResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping_resource"
}

func (r *SystemMappingResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector System Mapping Resource Resource.
				
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

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
				Optional:            true,
			},
			"exact_match_only": schema.BoolAttribute{
				MarkdownDescription: "Boolean flag determining whether access is granted only if the requested resource is an exact match.",
				Computed:            true,
				Optional:            true,
			},
			"websocket_upgrade_allowed": schema.BoolAttribute{
				MarkdownDescription: "Boolean flag indicating whether websocket upgrade is allowed. This property is of relevance only if the owning system mapping employs protocol HTTP or HTTPS.",
				Computed:            true,
				Optional:            true,
			},
			"creation_date": schema.StringAttribute{
				MarkdownDescription: "Date of creation of system mapping resource.",
				Computed:            true,
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the system mapping resource.",
				Computed:            true,
				Optional:            true,
			},
		},
	}
}

func (r *SystemMappingResourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SystemMappingResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	virtualHost := plan.VirtualHost.ValueString()
	virtualPort := plan.VirtualPort.ValueString()
	resource_id := CreateEncodedResourceID(plan.ID.ValueString())
	endpoint := endpoints.GetSystemMappingResourceBaseEndpoint(regionHost, subaccount, virtualHost, virtualPort)

	planBody := map[string]string{
		"id":                      plan.ID.ValueString(),
		"enabled":                 fmt.Sprintf("%t", plan.Enabled.ValueBool()),
		"exactMatchOnly":          fmt.Sprintf("%t", plan.ExactMatchOnly.ValueBool()),
		"websocketUpgradeAllowed": fmt.Sprintf("%t", plan.WebsocketUpgradeAllowed.ValueBool()),
		"description":             plan.Description.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgAddSystemMappingResourceFailed, err.Error())
		return
	}

	endpoint = endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resource_id)

	err = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSystemMappingResourceFailed, err.Error())
		return
	}

	responseModel, err := SystemMappingResourceValueFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapSystemMappingResourceFailed, fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	virtualHost := state.VirtualHost.ValueString()
	virtualPort := state.VirtualPort.ValueString()
	resource_id := CreateEncodedResourceID(state.ID.ValueString())
	endpoint := endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resource_id)

	err := requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSystemMappingResourceFailed, err.Error())
		return
	}

	responseModel, err := SystemMappingResourceValueFrom(ctx, state, respObj)
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

func (r *SystemMappingResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
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
	resource_id := CreateEncodedResourceID(plan.ID.ValueString())

	if (plan.RegionHost.ValueString() != regionHost) ||
		(plan.Subaccount.ValueString() != subaccount) ||
		(plan.VirtualHost.ValueString() != virtualHost) ||
		(plan.VirtualPort.ValueString() != virtualPort) {
		resp.Diagnostics.AddError(errMsgUpdateSystemMappingResourceFailed, "Failed to update the cloud connector system mapping resource due to mismatched configuration values.")
		return
	}
	endpoint := fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/systemMappings/%s:%s/resources/%s", regionHost, subaccount, virtualHost, virtualPort, resource_id)

	planBody := map[string]string{
		"enabled":                 fmt.Sprintf("%t", plan.Enabled.ValueBool()),
		"exactMatchOnly":          fmt.Sprintf("%t", plan.ExactMatchOnly.ValueBool()),
		"websocketUpgradeAllowed": fmt.Sprintf("%t", plan.WebsocketUpgradeAllowed.ValueBool()),
		"description":             plan.Description.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgUpdateSystemMappingResourceFailed, err.Error())
		return
	}

	endpoint = endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resource_id)

	err = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSystemMappingResourceFailed, err.Error())
		return
	}

	responseModel, err := SystemMappingResourceValueFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapSystemMappingResourceFailed, fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	virtualHost := state.VirtualHost.ValueString()
	virtualPort := state.VirtualPort.ValueString()
	resource_id := CreateEncodedResourceID(state.ID.ValueString())
	endpoint := endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resource_id)

	err := requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgDeleteSystemMappingResourceFailed, err.Error())
		return
	}

	responseModel, err := SystemMappingResourceValueFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapSystemMappingResourceFailed, fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (rs *SystemMappingResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 5 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: region_host, subaccount, virtual_host, virtual_port, id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_host"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_port"), idParts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[4])...)
}
