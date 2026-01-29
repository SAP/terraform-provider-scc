package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubaccountK8SServiceChannel struct {
	K8SClusterHost types.String `tfsdk:"k8s_cluster_host"`
	K8SServiceID   types.String `tfsdk:"k8s_service_id"`
	ID             types.Int64  `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	LocalPort      types.Int64  `tfsdk:"local_port"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Connections    types.Int64  `tfsdk:"connections"`
	Description    types.String `tfsdk:"description"`
	State          types.Object `tfsdk:"state"`
}

type SubaccountK8SServiceChannelStateData struct {
	Connected               types.Bool  `tfsdk:"connected"`
	OpenedConnections       types.Int64 `tfsdk:"opened_connections"`
	ConnectedSinceTimeStamp types.Int64 `tfsdk:"connected_since_time_stamp"`
}

var SubaccountK8SServiceChannelStateType = map[string]attr.Type{
	"connected":                  types.BoolType,
	"opened_connections":         types.Int64Type,
	"connected_since_time_stamp": types.Int64Type,
}

type SubaccountK8SServiceChannelConfig struct {
	RegionHost     types.String `tfsdk:"region_host"`
	Subaccount     types.String `tfsdk:"subaccount"`
	K8SClusterHost types.String `tfsdk:"k8s_cluster_host"`
	K8SServiceID   types.String `tfsdk:"k8s_service_id"`
	ID             types.Int64  `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	LocalPort      types.Int64  `tfsdk:"local_port"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Connections    types.Int64  `tfsdk:"connections"`
	Description    types.String `tfsdk:"description"`
	State          types.Object `tfsdk:"state"`
}

type SubaccountK8SServiceChannelsConfig struct {
	RegionHost                   types.String                  `tfsdk:"region_host"`
	Subaccount                   types.String                  `tfsdk:"subaccount"`
	SubaccountK8SServiceChannels []SubaccountK8SServiceChannel `tfsdk:"subaccount_k8s_service_channels"`
}

func SubaccountK8SServiceChannelValueFrom(ctx context.Context, plan SubaccountK8SServiceChannelConfig, value apiobjects.SubaccountK8SServiceChannel) (SubaccountK8SServiceChannelConfig, diag.Diagnostics) {
	stateObj := SubaccountK8SServiceChannelStateData{
		Connected:               types.BoolValue(value.State.Connected),
		OpenedConnections:       types.Int64Value(value.State.OpenedConnections),
		ConnectedSinceTimeStamp: types.Int64Value(value.State.ConnectedSinceTimeStamp),
	}

	state, diags := types.ObjectValueFrom(ctx, SubaccountK8SServiceChannelStateType, stateObj)
	if diags.HasError() {
		return SubaccountK8SServiceChannelConfig{}, diags
	}

	model := &SubaccountK8SServiceChannelConfig{
		RegionHost:     plan.RegionHost,
		Subaccount:     plan.Subaccount,
		K8SClusterHost: types.StringValue(value.K8SClusterHost),
		K8SServiceID:   types.StringValue(value.K8SServiceID),
		ID:             types.Int64Value(value.ID),
		Type:           types.StringValue(value.Type),
		LocalPort:      types.Int64Value(value.LocalPort),
		Enabled:        types.BoolValue(value.Enabled),
		Connections:    types.Int64Value(value.Connections),
		Description:    types.StringValue(value.Description),
		State:          state,
	}

	return *model, nil
}

func SubaccountK8SServiceChannelsValueFrom(ctx context.Context, plan SubaccountK8SServiceChannelsConfig, value apiobjects.SubaccountK8SServiceChannels) (SubaccountK8SServiceChannelsConfig, diag.Diagnostics) {
	serviceChannels := []SubaccountK8SServiceChannel{}
	for _, channel := range value.SubaccountK8SServiceChannels {
		stateObj := SubaccountK8SServiceChannelStateData{
			Connected:               types.BoolValue(channel.State.Connected),
			OpenedConnections:       types.Int64Value(channel.State.OpenedConnections),
			ConnectedSinceTimeStamp: types.Int64Value(channel.State.ConnectedSinceTimeStamp),
		}

		state, diags := types.ObjectValueFrom(ctx, SubaccountK8SServiceChannelStateType, stateObj)
		if diags.HasError() {
			return SubaccountK8SServiceChannelsConfig{}, diags
		}

		c := SubaccountK8SServiceChannel{
			K8SClusterHost: types.StringValue(channel.K8SClusterHost),
			K8SServiceID:   types.StringValue(channel.K8SServiceID),
			ID:             types.Int64Value(channel.ID),
			Type:           types.StringValue(channel.Type),
			LocalPort:      types.Int64Value(channel.LocalPort),
			Enabled:        types.BoolValue(channel.Enabled),
			Connections:    types.Int64Value(channel.Connections),
			Description:    types.StringValue(channel.Description),
			State:          state,
		}
		serviceChannels = append(serviceChannels, c)
	}

	model := &SubaccountK8SServiceChannelsConfig{
		RegionHost:                   plan.RegionHost,
		Subaccount:                   plan.Subaccount,
		SubaccountK8SServiceChannels: serviceChannels,
	}

	return *model, nil
}

func SubaccountK8SServiceChannelListValueFrom(ctx context.Context, filter subaccountK8SServiceChannelListResourceFilterModel, value apiobjects.SubaccountK8SServiceChannel) (*SubaccountK8SServiceChannelConfig, diag.Diagnostics) {
	stateObj := SubaccountK8SServiceChannelStateData{
		Connected:               types.BoolValue(value.State.Connected),
		OpenedConnections:       types.Int64Value(value.State.OpenedConnections),
		ConnectedSinceTimeStamp: types.Int64Value(value.State.ConnectedSinceTimeStamp),
	}

	state, diags := types.ObjectValueFrom(ctx, SubaccountK8SServiceChannelStateType, stateObj)
	if diags.HasError() {
		return &SubaccountK8SServiceChannelConfig{}, diags
	}

	listRes := &SubaccountK8SServiceChannelConfig{
		RegionHost:     filter.RegionHost,
		Subaccount:     filter.Subaccount,
		K8SClusterHost: types.StringValue(value.K8SClusterHost),
		K8SServiceID:   types.StringValue(value.K8SServiceID),
		ID:             types.Int64Value(value.ID),
		Type:           types.StringValue(value.Type),
		LocalPort:      types.Int64Value(value.LocalPort),
		Enabled:        types.BoolValue(value.Enabled),
		Connections:    types.Int64Value(value.Connections),
		Description:    types.StringValue(value.Description),
		State:          state,
	}

	return listRes, nil
}
