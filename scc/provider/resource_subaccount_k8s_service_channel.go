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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.Resource = &SubaccountK8SServiceChannelResource{}

func NewSubaccountK8SServiceChannelResource() resource.Resource {
	return &SubaccountK8SServiceChannelResource{}
}

type SubaccountK8SServiceChannelResource struct {
	client *api.RestApiClient
}

func (r *SubaccountK8SServiceChannelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_k8s_service_channel"
}

func (r *SubaccountK8SServiceChannelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount K8S Service Channel Resource.

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
			"k8s_cluster_host": schema.StringAttribute{
				MarkdownDescription: "Host name to access the Kubernetes cluster.",
				Required:            true,
			},

			"k8s_service_id": schema.StringAttribute{
				MarkdownDescription: "Host name providing the service inside of Kubernetes cluster.",
				Required:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Unique identifier for the subaccount service channel (a positive integer number, starting with 1). This identifier is unique across all types of service channels.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of Subaccount Service Channel.",
				Computed:            true,
			},
			"local_port": schema.Int64Attribute{
				MarkdownDescription: "Port of the subaccount service channel for the Kubernetes Cluster.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
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
			"description": schema.StringAttribute{
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

func (r *SubaccountK8SServiceChannelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *SubaccountK8SServiceChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubaccountK8SServiceChannelConfig
	var respObj apiobjects.SubaccountK8SServiceChannels
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, "K8S")

	planBody := map[string]any{
		"k8sCluster":  plan.K8SClusterHost.ValueString(),
		"k8sService":  plan.K8SServiceID.ValueString(),
		"port":        fmt.Sprintf("%d", plan.LocalPort.ValueInt64()),
		"connections": fmt.Sprintf("%d", plan.Connections.ValueInt64()),
		"comment":     plan.Description.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj.SubaccountK8SServiceChannels, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = requestAndUnmarshal(r.client, &respObj.SubaccountK8SServiceChannels, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceChannelRespObj, diags := r.getSubaccountK8SServiceChannel(respObj, plan.K8SClusterHost.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := serviceChannelRespObj.ID

	if !plan.Enabled.IsNull() {
		endpoint = endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "K8S", id)
		diags = r.enableSubaccountK8SServiceChannel(plan, *serviceChannelRespObj, endpoint+"/state")
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

	responseModel, diags := SubaccountK8SServiceChannelValueFrom(ctx, plan, *serviceChannelRespObj)
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

func (r *SubaccountK8SServiceChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubaccountK8SServiceChannelConfig
	var respObj apiobjects.SubaccountK8SServiceChannel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	id := state.ID.ValueInt64()
	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "K8S", id)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountK8SServiceChannelValueFrom(ctx, state, respObj)
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

func (r *SubaccountK8SServiceChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SubaccountK8SServiceChannelConfig
	var respObj apiobjects.SubaccountK8SServiceChannel

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
		resp.Diagnostics.AddError("Error updating the cloud connector subaccount K8S service channel", "Failed to update the cloud connector k8s service channel due to mismatched configuration values.")
		return
	}
	// Update Service Channel
	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "K8S", id)
	diags = r.updateSubaccountK8SServiceChannel(plan, respObj, endpoint)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Enable/Disable Service Channel
	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		diags = r.enableSubaccountK8SServiceChannel(plan, respObj, endpoint+"/state")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

	}

	endpoint = endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "K8S", id)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountK8SServiceChannelValueFrom(ctx, plan, respObj)
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

func (r *SubaccountK8SServiceChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubaccountK8SServiceChannelConfig
	var respObj apiobjects.SubaccountK8SServiceChannel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	id := state.ID.ValueInt64()

	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, "K8S", id)

	diags = requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SubaccountK8SServiceChannelValueFrom(ctx, state, respObj)
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

func (r *SubaccountK8SServiceChannelResource) getSubaccountK8SServiceChannel(serviceChannels apiobjects.SubaccountK8SServiceChannels, targetK8SCluster string) (*apiobjects.SubaccountK8SServiceChannel, diag.Diagnostics) {
	var diags diag.Diagnostics
	for _, channel := range serviceChannels.SubaccountK8SServiceChannels {
		if channel.K8SClusterHost == targetK8SCluster {
			return &channel, nil
		}
	}
	diags.AddError("Subaccount K8S Service Channel Not Found", "The specified subaccount k8s service channel was not found.")
	return nil, diags
}

func (r *SubaccountK8SServiceChannelResource) enableSubaccountK8SServiceChannel(plan SubaccountK8SServiceChannelConfig, respObj apiobjects.SubaccountK8SServiceChannel, endpoint string) diag.Diagnostics {
	planBody := map[string]any{
		"enabled": fmt.Sprintf("%t", plan.Enabled.ValueBool()),
	}

	diags := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (r *SubaccountK8SServiceChannelResource) updateSubaccountK8SServiceChannel(plan SubaccountK8SServiceChannelConfig, respObj apiobjects.SubaccountK8SServiceChannel, endpoint string) diag.Diagnostics {
	planBody := map[string]any{
		"k8sCluster":  plan.K8SClusterHost.ValueString(),
		"k8sService":  plan.K8SServiceID.ValueString(),
		"port":        fmt.Sprintf("%d", plan.LocalPort.ValueInt64()),
		"connections": fmt.Sprintf("%d", plan.Connections.ValueInt64()),
		"comment":     plan.Description.ValueString(),
	}

	diags := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (rs *SubaccountK8SServiceChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
