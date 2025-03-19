package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubaccountServiceChannelK8SData struct {
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

type SubaccountServiceChannelK8SStateData struct {
	Connected               types.Bool  `tfsdk:"connected"`
	OpenedConnections       types.Int64 `tfsdk:"opened_connections"`
	ConnectedSinceTimeStamp types.Int64 `tfsdk:"connected_since_time_stamp"`
}

var SubaccountServiceChannelK8SStateType = map[string]attr.Type{
	"connected":                  types.BoolType,
	"opened_connections":         types.Int64Type,
	"connected_since_time_stamp": types.Int64Type,
}

type SubaccountServiceChannelK8SCredentials struct {
	RegionHost types.String `tfsdk:"region_host"`
	Subaccount types.String `tfsdk:"subaccount"`
}

type SubaccountServiceChannelsK8SConfig struct {
	SubaccountServiceChannelK8SCredentials SubaccountServiceChannelK8SCredentials `tfsdk:"credentials"`
	SubaccountServiceChannelsK8SData       []SubaccountServiceChannelK8SData      `tfsdk:"subaccount_service_channels_k8s"`
}

type SubaccountServiceChannelK8SConfig struct {
	SubaccountServiceChannelK8SCredentials SubaccountServiceChannelK8SCredentials `tfsdk:"credentials"`
	SubaccountServiceChannelK8SData        SubaccountServiceChannelK8SData        `tfsdk:"subaccount_service_channel_k8s"`
}

func SubaccountServiceChannelK8SValueFrom(ctx context.Context, plan SubaccountServiceChannelK8SConfig, value apiobjects.SubaccountServiceChannelK8S) (SubaccountServiceChannelK8SConfig, error) {
	stateObj := SubaccountServiceChannelK8SStateData{
		Connected:               types.BoolValue(value.State.Connected),
		OpenedConnections:       types.Int64Value(value.State.OpenedConnections),
		ConnectedSinceTimeStamp: types.Int64Value(value.State.ConnectedSinceTimeStamp),
	}

	state, _ := types.ObjectValueFrom(ctx, SubaccountServiceChannelK8SStateType, stateObj)

	credentials := SubaccountServiceChannelK8SCredentials{
		RegionHost: plan.SubaccountServiceChannelK8SCredentials.RegionHost,
		Subaccount: plan.SubaccountServiceChannelK8SCredentials.Subaccount,
	}

	serviceChannel := SubaccountServiceChannelK8SData{
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

	model := &SubaccountServiceChannelK8SConfig{
		SubaccountServiceChannelK8SCredentials: credentials,
		SubaccountServiceChannelK8SData:        serviceChannel,
	}

	return *model, nil
}

func ServiceChannelsK8SValueFrom(ctx context.Context, plan SubaccountServiceChannelsK8SConfig, value apiobjects.SubaccountServiceChannelsK8S) (SubaccountServiceChannelsK8SConfig, error) {
	serviceChannels := []SubaccountServiceChannelK8SData{}
	for _, channel := range value.SubaccountServiceChannelsK8S {
		stateObj := SubaccountServiceChannelK8SStateData{
			Connected:               types.BoolValue(channel.State.Connected),
			OpenedConnections:       types.Int64Value(channel.State.OpenedConnections),
			ConnectedSinceTimeStamp: types.Int64Value(channel.State.ConnectedSinceTimeStamp),
		}

		state, _ := types.ObjectValueFrom(ctx, SubaccountServiceChannelK8SStateType, stateObj)

		c := SubaccountServiceChannelK8SData{
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

	credentials := SubaccountServiceChannelK8SCredentials{
		RegionHost: plan.SubaccountServiceChannelK8SCredentials.RegionHost,
		Subaccount: plan.SubaccountServiceChannelK8SCredentials.Subaccount,
	}

	model := &SubaccountServiceChannelsK8SConfig{
		SubaccountServiceChannelK8SCredentials: credentials,
		SubaccountServiceChannelsK8SData:       serviceChannels,
	}

	return *model, nil
}
