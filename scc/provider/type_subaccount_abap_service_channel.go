package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubaccountABAPServiceChannel struct {
	ABAPCloudTenantHost types.String `tfsdk:"abap_cloud_tenant_host"`
	InstanceNumber      types.Int64  `tfsdk:"instance_number"`
	ID                  types.Int64  `tfsdk:"id"`
	Type                types.String `tfsdk:"type"`
	Port                types.Int64  `tfsdk:"port"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Connections         types.Int64  `tfsdk:"connections"`
	Comment             types.String `tfsdk:"comment"`
	State               types.Object `tfsdk:"state"`
}

type SubaccountABAPServiceChannelStateData struct {
	Connected               types.Bool  `tfsdk:"connected"`
	OpenedConnections       types.Int64 `tfsdk:"opened_connections"`
	ConnectedSinceTimeStamp types.Int64 `tfsdk:"connected_since_time_stamp"`
}

var SubaccountABAPServiceChannelStateType = map[string]attr.Type{
	"connected":                  types.BoolType,
	"opened_connections":         types.Int64Type,
	"connected_since_time_stamp": types.Int64Type,
}

type SubaccountABAPServiceChannelConfig struct {
	RegionHost          types.String `tfsdk:"region_host"`
	Subaccount          types.String `tfsdk:"subaccount"`
	ABAPCloudTenantHost types.String `tfsdk:"abap_cloud_tenant_host"`
	InstanceNumber      types.Int64  `tfsdk:"instance_number"`
	ID                  types.Int64  `tfsdk:"id"`
	Type                types.String `tfsdk:"type"`
	Port                types.Int64  `tfsdk:"port"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Connections         types.Int64  `tfsdk:"connections"`
	Comment             types.String `tfsdk:"comment"`
	State               types.Object `tfsdk:"state"`
}

type SubaccountABAPServiceChannelsConfig struct {
	RegionHost                    types.String                   `tfsdk:"region_host"`
	Subaccount                    types.String                   `tfsdk:"subaccount"`
	SubaccountABAPServiceChannels []SubaccountABAPServiceChannel `tfsdk:"subaccount_abap_service_channels"`
}

func SubaccountABAPServiceChannelValueFrom(ctx context.Context, plan SubaccountABAPServiceChannelConfig, value apiobjects.SubaccountABAPServiceChannel) (SubaccountABAPServiceChannelConfig, diag.Diagnostics) {
	stateObj := SubaccountABAPServiceChannelStateData{
		Connected:               types.BoolValue(value.State.Connected),
		OpenedConnections:       types.Int64Value(value.State.OpenedConnections),
		ConnectedSinceTimeStamp: types.Int64Value(value.State.ConnectedSinceTimeStamp),
	}

	state, diags := types.ObjectValueFrom(ctx, SubaccountABAPServiceChannelStateType, stateObj)
	if diags.HasError() {
		return SubaccountABAPServiceChannelConfig{}, diags
	}

	model := &SubaccountABAPServiceChannelConfig{
		RegionHost:          plan.RegionHost,
		Subaccount:          plan.Subaccount,
		ABAPCloudTenantHost: types.StringValue(value.ABAPCloudTenantHost),
		InstanceNumber:      types.Int64Value(value.InstanceNumber),
		ID:                  types.Int64Value(value.ID),
		Type:                types.StringValue(value.Type),
		Port:                types.Int64Value(value.Port),
		Enabled:             types.BoolValue(value.Enabled),
		Connections:         types.Int64Value(value.Connections),
		Comment:             types.StringValue(value.Comment),
		State:               state,
	}

	return *model, nil
}

func SubaccountABAPServiceChannelsValueFrom(ctx context.Context, plan SubaccountABAPServiceChannelsConfig, value apiobjects.SubaccountABAPServiceChannels) (SubaccountABAPServiceChannelsConfig, diag.Diagnostics) {
	serviceChannels := []SubaccountABAPServiceChannel{}
	for _, channel := range value.SubaccountABAPServiceChannels {
		stateObj := SubaccountABAPServiceChannelStateData{
			Connected:               types.BoolValue(channel.State.Connected),
			OpenedConnections:       types.Int64Value(channel.State.OpenedConnections),
			ConnectedSinceTimeStamp: types.Int64Value(channel.State.ConnectedSinceTimeStamp),
		}

		state, diags := types.ObjectValueFrom(ctx, SubaccountABAPServiceChannelStateType, stateObj)
		if diags.HasError() {
			return SubaccountABAPServiceChannelsConfig{}, diags
		}

		c := SubaccountABAPServiceChannel{
			ABAPCloudTenantHost: types.StringValue(channel.ABAPCloudTenantHost),
			InstanceNumber:      types.Int64Value(channel.InstanceNumber),
			ID:                  types.Int64Value(channel.ID),
			Type:                types.StringValue(channel.Type),
			Port:                types.Int64Value(channel.Port),
			Enabled:             types.BoolValue(channel.Enabled),
			Connections:         types.Int64Value(channel.Connections),
			Comment:             types.StringValue(channel.Comment),
			State:               state,
		}
		serviceChannels = append(serviceChannels, c)
	}

	model := &SubaccountABAPServiceChannelsConfig{
		RegionHost:                    plan.RegionHost,
		Subaccount:                    plan.Subaccount,
		SubaccountABAPServiceChannels: serviceChannels,
	}

	return *model, nil
}

func SubaccountABAPServiceChannelListValueFrom(ctx context.Context, filter subaccountABAPServiceChannelListResourceFilterModel, value apiobjects.SubaccountABAPServiceChannel) (*SubaccountABAPServiceChannelConfig, diag.Diagnostics) {
	stateObj := SubaccountABAPServiceChannelStateData{
		Connected:               types.BoolValue(value.State.Connected),
		OpenedConnections:       types.Int64Value(value.State.OpenedConnections),
		ConnectedSinceTimeStamp: types.Int64Value(value.State.ConnectedSinceTimeStamp),
	}

	state, diags := types.ObjectValueFrom(ctx, SubaccountABAPServiceChannelStateType, stateObj)
	if diags.HasError() {
		return &SubaccountABAPServiceChannelConfig{}, diags
	}

	listRes := &SubaccountABAPServiceChannelConfig{
		RegionHost:          filter.RegionHost,
		Subaccount:          filter.Subaccount,
		ABAPCloudTenantHost: types.StringValue(value.ABAPCloudTenantHost),
		InstanceNumber:      types.Int64Value(value.InstanceNumber),
		ID:                  types.Int64Value(value.ID),
		Type:                types.StringValue(value.Type),
		Port:                types.Int64Value(value.Port),
		Enabled:             types.BoolValue(value.Enabled),
		Connections:         types.Int64Value(value.Connections),
		Comment:             types.StringValue(value.Comment),
		State:               state,
	}

	return listRes, nil
}
