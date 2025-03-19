package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-cloudconnector/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.Resource = &SubaccountServiceChannelK8SResource{}

func NewSubaccountServiceChannelK8SResource() resource.Resource {
	return &SubaccountServiceChannelK8SResource{}
}

type SubaccountServiceChannelK8SResource struct {
	client *api.RestApiClient
}

func (r *SubaccountServiceChannelK8SResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_service_channel_k8s"
}

func (r *SubaccountServiceChannelK8SResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cloud Connector Subaccount Service Channel K8S Resource",
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
			"subaccount_service_channel_k8s": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"k8s_cluster": schema.StringAttribute{
						MarkdownDescription: "Host name to access the Kubernetes cluster.",
						Required:            true,
					},

					"k8s_service": schema.StringAttribute{
						MarkdownDescription: "Host name providiing the service inside of Kubernetes cluster.",
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
					"port": schema.Int64Attribute{
						MarkdownDescription: "Port of the subaccount service channel for the virtual machine.",
						Required:            true,
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
			},
		},
	}
}

func (r *SubaccountServiceChannelK8SResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubaccountServiceChannelK8SResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubaccountServiceChannelK8SConfig
	var respObj apiobjects.SubaccountServiceChannelsK8S
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := plan.SubaccountServiceChannelK8SCredentials.RegionHost.ValueString()
	subaccount := plan.SubaccountServiceChannelK8SCredentials.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountServiceChannelBaseEndpoint(region_host, subaccount, "K8S")

	planBody := map[string]string{
		"k8sCluster":  plan.SubaccountServiceChannelK8SData.K8SCluster.ValueString(),
		"k8sService":  plan.SubaccountServiceChannelK8SData.K8SService.ValueString(),
		"port":        fmt.Sprintf("%d", plan.SubaccountServiceChannelK8SData.Port.ValueInt64()),
		"connections": fmt.Sprintf("%d", plan.SubaccountServiceChannelK8SData.Connections.ValueInt64()),
		"comment":     plan.SubaccountServiceChannelK8SData.Comment.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj.SubaccountServiceChannelsK8S, "POST", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error creating the cloud connector subaccount service channel", err.Error())
		return
	}

	err = requestAndUnmarshal(r.client, &respObj.SubaccountServiceChannelsK8S, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector subaccount service channels", err.Error())
		return
	}

	serviceChannelRespObj, err := r.getSubaccountServiceChannel(respObj, plan.SubaccountServiceChannelK8SData.K8SCluster.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("error fetching the subaccount service channel", err.Error())
		return
	}

	id := serviceChannelRespObj.ID

	if !plan.SubaccountServiceChannelK8SData.Enabled.IsNull() {
		endpoint = endpoints.GetSubaccountServiceChannelEndpoint(region_host, subaccount, "K8S", id)
		r.enableSubaccountServiceChannel(plan, *serviceChannelRespObj, (*resource.UpdateResponse)(resp), endpoint+"/state")

		err = requestAndUnmarshal(r.client, &serviceChannelRespObj, "GET", endpoint, nil, true)
		if err != nil {
			resp.Diagnostics.AddError("error fetching the cloud connector subaccount service channel", err.Error())
			return
		}
	}

	responseModel, err := SubaccountServiceChannelK8SValueFrom(ctx, plan, *serviceChannelRespObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount service channel value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountServiceChannelK8SResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubaccountServiceChannelK8SConfig
	var respObj apiobjects.SubaccountServiceChannelK8S
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := state.SubaccountServiceChannelK8SCredentials.RegionHost.ValueString()
	subaccount := state.SubaccountServiceChannelK8SCredentials.Subaccount.ValueString()
	id := state.SubaccountServiceChannelK8SData.ID.ValueInt64()
	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(region_host, subaccount, "K8S", id)

	err := requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector subaccount service channel", err.Error())
		return
	}

	responseModel, err := SubaccountServiceChannelK8SValueFrom(ctx, state, respObj)
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

func (r *SubaccountServiceChannelK8SResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SubaccountServiceChannelK8SConfig
	var respObj apiobjects.SubaccountServiceChannelK8S

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

	region_host := plan.SubaccountServiceChannelK8SCredentials.RegionHost.ValueString()
	subaccount := plan.SubaccountServiceChannelK8SCredentials.Subaccount.ValueString()
	id := state.SubaccountServiceChannelK8SData.ID.ValueInt64()

	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(region_host, subaccount, "K8S", id)

	serviceChannelEnabled := plan.SubaccountServiceChannelK8SData.Enabled.ValueBool() != state.SubaccountServiceChannelK8SData.Enabled.ValueBool()

	configUpdated := (plan.SubaccountServiceChannelK8SData.K8SCluster.ValueString() != state.SubaccountServiceChannelK8SData.K8SCluster.ValueString()) ||
		(plan.SubaccountServiceChannelK8SData.K8SService.ValueString() != state.SubaccountServiceChannelK8SData.K8SService.ValueString()) ||
		(plan.SubaccountServiceChannelK8SData.Port.ValueInt64() != state.SubaccountServiceChannelK8SData.Port.ValueInt64()) ||
		(plan.SubaccountServiceChannelK8SData.Connections.ValueInt64() != state.SubaccountServiceChannelK8SData.Connections.ValueInt64()) ||
		(plan.SubaccountServiceChannelK8SData.Comment.ValueString() != state.SubaccountServiceChannelK8SData.Comment.ValueString())

	if configUpdated {
		r.updateSubaccountServiceChannel(plan, respObj, resp, endpoint)
		if plan.SubaccountServiceChannelK8SData.Enabled.ValueBool() {
			r.enableSubaccountServiceChannel(plan, respObj, resp, endpoint+"/state")
		}
	}

	if serviceChannelEnabled {
		r.enableSubaccountServiceChannel(plan, respObj, resp, endpoint+"/state")
	}

	endpoint = endpoints.GetSubaccountServiceChannelEndpoint(region_host, subaccount, "K8S", id)

	err := requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector subaccount service channel", err.Error())
		return
	}

	responseModel, err := SubaccountServiceChannelK8SValueFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount service channel value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountServiceChannelK8SResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubaccountServiceChannelK8SConfig
	var respObj apiobjects.SubaccountServiceChannelK8S
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := state.SubaccountServiceChannelK8SCredentials.RegionHost.ValueString()
	subaccount := state.SubaccountServiceChannelK8SCredentials.Subaccount.ValueString()
	id := state.SubaccountServiceChannelK8SData.ID.ValueInt64()

	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(region_host, subaccount, "K8S", id)

	err := requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError("error deleting the subaccount service channel", err.Error())
		return
	}

	responseModel, err := SubaccountServiceChannelK8SValueFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount service channel value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountServiceChannelK8SResource) getSubaccountServiceChannel(serviceChannels apiobjects.SubaccountServiceChannelsK8S, targetK8SCluster string) (*apiobjects.SubaccountServiceChannelK8S, error) {
	for _, channel := range serviceChannels.SubaccountServiceChannelsK8S {
		if channel.K8SCluster == targetK8SCluster {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("%s", "subaccount service channel doesn't exist")
}

func (r *SubaccountServiceChannelK8SResource) enableSubaccountServiceChannel(plan SubaccountServiceChannelK8SConfig, respObj apiobjects.SubaccountServiceChannelK8S, resp *resource.UpdateResponse, endpoint string) {
	planBody := map[string]string{
		"enabled": fmt.Sprintf("%t", plan.SubaccountServiceChannelK8SData.Enabled.ValueBool()),
	}

	err := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error enabling the cloud connector subaccount service channel", err.Error())
		return
	}
}

func (r *SubaccountServiceChannelK8SResource) updateSubaccountServiceChannel(plan SubaccountServiceChannelK8SConfig, respObj apiobjects.SubaccountServiceChannelK8S, resp *resource.UpdateResponse, endpoint string) {
	planBody := map[string]string{
		"k8sCluster":  plan.SubaccountServiceChannelK8SData.K8SCluster.ValueString(),
		"k8sService":  plan.SubaccountServiceChannelK8SData.K8SService.ValueString(),
		"port":        fmt.Sprintf("%d", plan.SubaccountServiceChannelK8SData.Port.ValueInt64()),
		"connections": fmt.Sprintf("%d", plan.SubaccountServiceChannelK8SData.Connections.ValueInt64()),
		"comment":     plan.SubaccountServiceChannelK8SData.Comment.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error updating the cloud connector subaccount service channel", err.Error())
		return
	}
}
