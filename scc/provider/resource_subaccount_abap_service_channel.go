package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.Resource = &SubaccountABAPServiceChannelResource{}

func NewSubaccountABAPServiceChannelResource() resource.Resource {
	return &SubaccountABAPServiceChannelResource{}
}

type SubaccountABAPServiceChannelResource struct {
	client *api.RestApiClient
}

func (r *SubaccountABAPServiceChannelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_abap_service_channel"
}

func (r *SubaccountABAPServiceChannelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount ABAP Service Channel Resource.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

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
			"abap_cloud_tenant_host": schema.StringAttribute{
				MarkdownDescription: "Host name to access the Host of ABAP Cloud Tenant.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"instance_number": schema.Int64Attribute{
				MarkdownDescription: "Local Instance number under which the ABAP Cloud system is reachable for the client systems.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(00, 99),
				},
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Unique identifier for the subaccount service channel (a positive integer number, starting with 1). This identifier is unique across all types of service channels.",
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
				Optional:            true,
				Computed:            true,
			},
			"connections": schema.Int64Attribute{
				MarkdownDescription: "Maximal number of open connections.",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment or short description. This property is not supplied if no comment was provided.",
				Optional:            true,
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
	}
}

func (r *SubaccountABAPServiceChannelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *SubaccountABAPServiceChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannels
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, "ABAPCloud")

	planBody := map[string]any{
		"abapCloudTenantHost": plan.ABAPCloudTenantHost.ValueString(),
		"instanceNumber":      fmt.Sprintf("%d", plan.InstanceNumber.ValueInt64()),
		"connections":         fmt.Sprintf("%d", plan.Connections.ValueInt64()),
		"comment":             plan.Comment.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj.SubaccountABAPServiceChannels, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = requestAndUnmarshal(r.client, &respObj.SubaccountABAPServiceChannels, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceChannelRespObj, diags := r.getSubaccountABAPServiceChannel(respObj, plan.ABAPCloudTenantHost.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := serviceChannelRespObj.ID

	if !plan.Enabled.IsNull() {
		endpoint = endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "ABAPCloud", id)
		diags = r.enableSubaccountABAPServiceChannel(plan, *serviceChannelRespObj, endpoint+"/state")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		diags = requestAndUnmarshal(r.client, &serviceChannelRespObj, "GET", endpoint, nil, true)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	responseModel, diags := SubaccountABAPServiceChannelValueFrom(ctx, plan, *serviceChannelRespObj)
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

func (r *SubaccountABAPServiceChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	id := state.ID.ValueInt64()
	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "ABAPCloud", id)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountABAPServiceChannelValueFrom(ctx, state, respObj)
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

func (r *SubaccountABAPServiceChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannel

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
	id := state.ID.ValueInt64()

	if (state.RegionHost.ValueString() != regionHost) ||
		(state.Subaccount.ValueString() != subaccount) {
		resp.Diagnostics.AddError("Error updating the cloud connector subaccount ABAP service channel", "Failed to update the cloud connector ABAP service channel due to mismatched configuration values.")
		return
	}
	// Update Service Channel
	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "ABAPCloud", id)
	diags = r.updateSubaccountABAPServiceChannel(plan, respObj, endpoint)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Enable/Disable Service Channel
	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		diags = r.enableSubaccountABAPServiceChannel(plan, respObj, endpoint+"/state")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	endpoint = endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "ABAPCloud", id)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountABAPServiceChannelValueFrom(ctx, plan, respObj)
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

func (r *SubaccountABAPServiceChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	id := state.ID.ValueInt64()

	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "ABAPCloud", id)

	diags = requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountABAPServiceChannelValueFrom(ctx, state, respObj)
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

func (r *SubaccountABAPServiceChannelResource) getSubaccountABAPServiceChannel(serviceChannels apiobjects.SubaccountABAPServiceChannels, targetABAPCloudTenantHost string) (*apiobjects.SubaccountABAPServiceChannel, diag.Diagnostics) {
	var diags diag.Diagnostics
	for _, channel := range serviceChannels.SubaccountABAPServiceChannels {
		if channel.ABAPCloudTenantHost == targetABAPCloudTenantHost {
			return &channel, nil
		}
	}
	diags.AddError("Subaccount ABAP Service Channel Not Found", "The specified subaccount ABAP service channel with the given ABAP Cloud Tenant Host was not found.")
	return nil, diags
}

func (r *SubaccountABAPServiceChannelResource) enableSubaccountABAPServiceChannel(plan SubaccountABAPServiceChannelConfig, respObj apiobjects.SubaccountABAPServiceChannel, endpoint string) diag.Diagnostics {
	planBody := map[string]any{
		"enabled": fmt.Sprintf("%t", plan.Enabled.ValueBool()),
	}

	diags := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (r *SubaccountABAPServiceChannelResource) updateSubaccountABAPServiceChannel(plan SubaccountABAPServiceChannelConfig, respObj apiobjects.SubaccountABAPServiceChannel, endpoint string) diag.Diagnostics {
	planBody := map[string]any{
		"abapCloudTenantHost": plan.ABAPCloudTenantHost.ValueString(),
		"instanceNumber":      fmt.Sprintf("%d", plan.InstanceNumber.ValueInt64()),
		"connections":         fmt.Sprintf("%d", plan.Connections.ValueInt64()),
		"comment":             plan.Comment.ValueString(),
	}

	diags := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (rs *SubaccountABAPServiceChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: region_host, subaccount, id. Got: %q", req.ID),
		)
		return
	}

	intID, diags := strconv.Atoi(idParts[2])
	if diags != nil {
		resp.Diagnostics.AddError(
			"Invalid ID Format",
			fmt.Sprintf("The 'id' part must be an integer. Got: %q", idParts[2]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), intID)...)

}
