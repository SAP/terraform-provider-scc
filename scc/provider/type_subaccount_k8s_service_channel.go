package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubaccountK8SServiceChannel struct {
	K8SCluster  types.String `tfsdk:"k8s_cluster"`
	K8SService  types.String `tfsdk:"k8s_service"`
	ID          types.Int64  `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Port        types.Int64  `tfsdk:"port"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Connections types.Int64  `tfsdk:"connections"`
	Comment     types.String `tfsdk:"comment"`
	State       types.Object `tfsdk:"state"`
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
	RegionHost  types.String `tfsdk:"region_host"`
	Subaccount  types.String `tfsdk:"subaccount"`
	K8SCluster  types.String `tfsdk:"k8s_cluster"`
	K8SService  types.String `tfsdk:"k8s_service"`
	ID          types.Int64  `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Port        types.Int64  `tfsdk:"port"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Connections types.Int64  `tfsdk:"connections"`
	Comment     types.String `tfsdk:"comment"`
	State       types.Object `tfsdk:"state"`
}

type SubaccountK8SServiceChannelsConfig struct {
	RegionHost                   types.String                  `tfsdk:"region_host"`
	Subaccount                   types.String                  `tfsdk:"subaccount"`
	SubaccountK8SServiceChannels []SubaccountK8SServiceChannel `tfsdk:"subaccount_k8s_service_channels"`
}

func SubaccountK8SServiceChannelValueFrom(ctx context.Context, plan SubaccountK8SServiceChannelConfig, value apiobjects.SubaccountK8SServiceChannel) (SubaccountK8SServiceChannelConfig, error) {
	stateObj := SubaccountK8SServiceChannelStateData{
		Connected:               types.BoolValue(value.State.Connected),
		OpenedConnections:       types.Int64Value(value.State.OpenedConnections),
		ConnectedSinceTimeStamp: types.Int64Value(value.State.ConnectedSinceTimeStamp),
	}

	state, _ := types.ObjectValueFrom(ctx, SubaccountK8SServiceChannelStateType, stateObj)

	model := &SubaccountK8SServiceChannelConfig{
		RegionHost:  plan.RegionHost,
		Subaccount:  plan.Subaccount,
		K8SCluster:  types.StringValue(value.K8SCluster),
		K8SService:  types.StringValue(value.K8SService),
		ID:          types.Int64Value(value.ID),
		Type:        types.StringValue(value.Type),
		Port:        types.Int64Value(value.Port),
		Enabled:     types.BoolValue(value.Enabled),
		Connections: types.Int64Value(value.Connections),
		Comment:     types.StringValue(value.Comment),
		State:       state,
	}

	return *model, nil
}

func SubaccountK8SServiceChannelsValueFrom(ctx context.Context, plan SubaccountK8SServiceChannelsConfig, value apiobjects.SubaccountK8SServiceChannels) (SubaccountK8SServiceChannelsConfig, error) {
	serviceChannels := []SubaccountK8SServiceChannel{}
	for _, channel := range value.SubaccountK8SServiceChannels {
		stateObj := SubaccountK8SServiceChannelStateData{
			Connected:               types.BoolValue(channel.State.Connected),
			OpenedConnections:       types.Int64Value(channel.State.OpenedConnections),
			ConnectedSinceTimeStamp: types.Int64Value(channel.State.ConnectedSinceTimeStamp),
		}

		state, _ := types.ObjectValueFrom(ctx, SubaccountK8SServiceChannelStateType, stateObj)

		c := SubaccountK8SServiceChannel{
			K8SCluster:  types.StringValue(channel.K8SCluster),
			K8SService:  types.StringValue(channel.K8SService),
			ID:          types.Int64Value(channel.ID),
			Type:        types.StringValue(channel.Type),
			Port:        types.Int64Value(channel.Port),
			Enabled:     types.BoolValue(channel.Enabled),
			Connections: types.Int64Value(channel.Connections),
			Comment:     types.StringValue(channel.Comment),
			State:       state,
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
