package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubaccountData struct {
	RegionHost  types.String `tfsdk:"region_host"`
	Subaccount  types.String `tfsdk:"subaccount"`
	LocationID  types.String `tfsdk:"location_id"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Tunnel      types.Object `tfsdk:"tunnel"`
}

type SubaccountTunnelData struct {
	State                   types.String `tfsdk:"state"`
	ConnectedSinceTimeStamp types.Int64  `tfsdk:"connected_since_time_stamp"`
	Connections             types.Int64  `tfsdk:"connections"`
	SubaccountCertificate   types.Object `tfsdk:"subaccount_certificate"`
	User                    types.String `tfsdk:"user"`
	ApplicationConnections  types.List   `tfsdk:"application_connections"`
	ServiceChannels         types.List   `tfsdk:"service_channels"`
}

var SubaccountTunnelType = map[string]attr.Type{
	"state":                      types.StringType,
	"connected_since_time_stamp": types.Int64Type,
	"connections":                types.Int64Type,
	"user":                       types.StringType,
	"subaccount_certificate": types.ObjectType{
		AttrTypes: SubaccountCertificateType,
	},
	"application_connections": types.ListType{
		ElemType: SubaccountApplicationConnectionsType,
	},
	"service_channels": types.ListType{
		ElemType: SubaccountServiceChannelsType,
	},
}

type SubaccountCertificateData struct {
	NotAfterTimeStamp  types.Int64  `tfsdk:"not_after_time_stamp"`
	NotBeforeTimeStamp types.Int64  `tfsdk:"not_before_time_stamp"`
	SubjectDN          types.String `tfsdk:"subject_dn"`
	Issuer             types.String `tfsdk:"issuer"`
	SerialNumber       types.String `tfsdk:"serial_number"`
}

var SubaccountCertificateType = map[string]attr.Type{
	"not_after_time_stamp":  types.Int64Type,
	"not_before_time_stamp": types.Int64Type,
	"subject_dn":            types.StringType,
	"issuer":                types.StringType,
	"serial_number":         types.StringType,
}

type SubaccountApplicationConnectionsData struct {
	ConnectionCount types.Int64  `tfsdk:"connection_count"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
}

var SubaccountApplicationConnectionsType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"connection_count": types.Int64Type,
		"name":             types.StringType,
		"type":             types.StringType,
	},
}

type SubaccountServiceChannelsData struct {
	Type    types.String `tfsdk:"type"`
	State   types.String `tfsdk:"state"`
	Details types.String `tfsdk:"details"`
	Comment types.String `tfsdk:"comment"`
}

var SubaccountServiceChannelsType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type":    types.StringType,
		"state":   types.StringType,
		"details": types.StringType,
		"comment": types.StringType,
	},
}

type SubaccountsData struct {
	RegionHost types.String `tfsdk:"region_host"`
	Subaccount types.String `tfsdk:"subaccount"`
	LocationID types.String `tfsdk:"location_id"`
}

type SubaccountsConfig struct {
	Subaccounts []SubaccountsData `tfsdk:"subaccounts"`
}

type SubaccountConfig struct {
	RegionHost    types.String `tfsdk:"region_host"`
	Subaccount    types.String `tfsdk:"subaccount"`
	CloudUser     types.String `tfsdk:"cloud_user"`
	CloudPassword types.String `tfsdk:"cloud_password"`
	LocationID    types.String `tfsdk:"location_id"`
	DisplayName   types.String `tfsdk:"display_name"`
	Description   types.String `tfsdk:"description"`
	Tunnel        types.Object `tfsdk:"tunnel"`
	Connected     types.Bool   `tfsdk:"connected"`
}

type SubaccountUsingAuthConfig struct {
	RegionHost         types.String `tfsdk:"region_host"`
	Subaccount         types.String `tfsdk:"subaccount"`
	AuthenticationData types.String `tfsdk:"authentication_data"`
	LocationID         types.String `tfsdk:"location_id"`
	DisplayName        types.String `tfsdk:"display_name"`
	Description        types.String `tfsdk:"description"`
	Tunnel             types.Object `tfsdk:"tunnel"`
	Connected          types.Bool   `tfsdk:"connected"`
}

func SubaccountsDataSourceValueFrom(value apiobjects.SubaccountsDataSource) (SubaccountsConfig, diag.Diagnostics) {
	subaccounts := []SubaccountsData{}
	for _, subaccount := range value.Subaccounts {
		c := SubaccountsData{
			RegionHost: types.StringValue(subaccount.RegionHost),
			Subaccount: types.StringValue(subaccount.Subaccount),
			LocationID: types.StringValue(subaccount.LocationID),
		}
		subaccounts = append(subaccounts, c)
	}
	model := &SubaccountsConfig{
		Subaccounts: subaccounts,
	}
	return *model, diag.Diagnostics{}
}

func SubaccountDataSourceValueFrom(ctx context.Context, value apiobjects.Subaccount) (SubaccountData, diag.Diagnostics) {
	certificateObj := SubaccountCertificateData{
		NotAfterTimeStamp:  types.Int64Value(value.Tunnel.SubaccountCertificate.NotAfterTimeStamp),
		NotBeforeTimeStamp: types.Int64Value(value.Tunnel.SubaccountCertificate.NotBeforeTimeStamp),
		SubjectDN:          types.StringValue(value.Tunnel.SubaccountCertificate.SubjectDN),
		Issuer:             types.StringValue(value.Tunnel.SubaccountCertificate.Issuer),
		SerialNumber:       types.StringValue(value.Tunnel.SubaccountCertificate.SerialNumber),
	}

	certificate, err := types.ObjectValueFrom(ctx, SubaccountCertificateType, certificateObj)
	if err.HasError() {
		return SubaccountData{}, err
	}

	applicationConnectionsValues := []SubaccountApplicationConnectionsData{}
	for _, connection := range value.Tunnel.ApplicationConnections {
		ac := SubaccountApplicationConnectionsData{
			ConnectionCount: types.Int64Value(connection.ConnectionCount),
			Name:            types.StringValue(connection.Name),
			Type:            types.StringValue(connection.Type),
		}

		applicationConnectionsValues = append(applicationConnectionsValues, ac)
	}

	applicationConnections, err := types.ListValueFrom(ctx, SubaccountApplicationConnectionsType, applicationConnectionsValues)
	if err.HasError() {
		return SubaccountData{}, err
	}

	serviceChannelsValues := []SubaccountServiceChannelsData{}
	for _, channel := range value.Tunnel.ServiceChannels {
		sc := SubaccountServiceChannelsData{
			Type:    types.StringValue(channel.Type),
			State:   types.StringValue(channel.State),
			Details: types.StringValue(channel.Details),
			Comment: types.StringValue(channel.Comment),
		}

		serviceChannelsValues = append(serviceChannelsValues, sc)
	}

	serviceChannels, err := types.ListValueFrom(ctx, SubaccountServiceChannelsType, serviceChannelsValues)
	if err.HasError() {
		return SubaccountData{}, err
	}

	tunnelObj := SubaccountTunnelData{
		State:                   types.StringValue(value.Tunnel.State),
		ConnectedSinceTimeStamp: types.Int64Value(value.Tunnel.ConnectedSinceTimeStamp),
		Connections:             types.Int64Value(value.Tunnel.Connections),
		User:                    types.StringValue(value.Tunnel.User),
		SubaccountCertificate:   certificate,
		ApplicationConnections:  applicationConnections,
		ServiceChannels:         serviceChannels,
	}

	tunnel, err := types.ObjectValueFrom(ctx, SubaccountTunnelType, tunnelObj)
	if err.HasError() {
		return SubaccountData{}, err
	}

	model := &SubaccountData{
		RegionHost:  types.StringValue(value.RegionHost),
		Subaccount:  types.StringValue(value.Subaccount),
		LocationID:  types.StringValue(value.LocationID),
		DisplayName: types.StringValue(value.DisplayName),
		Description: types.StringValue(value.Description),
		Tunnel:      tunnel,
	}
	return *model, diag.Diagnostics{}
}

func SubaccountResourceValueFrom(ctx context.Context, plan SubaccountConfig, value apiobjects.SubaccountResource) (SubaccountConfig, diag.Diagnostics) {
	certificateObj := SubaccountCertificateData{
		NotAfterTimeStamp:  types.Int64Value(value.Tunnel.SubaccountCertificate.NotAfterTimeStamp),
		NotBeforeTimeStamp: types.Int64Value(value.Tunnel.SubaccountCertificate.NotBeforeTimeStamp),
		SubjectDN:          types.StringValue(value.Tunnel.SubaccountCertificate.SubjectDN),
		Issuer:             types.StringValue(value.Tunnel.SubaccountCertificate.Issuer),
		SerialNumber:       types.StringValue(value.Tunnel.SubaccountCertificate.SerialNumber),
	}

	certificate, err := types.ObjectValueFrom(ctx, SubaccountCertificateType, certificateObj)
	if err.HasError() {
		return SubaccountConfig{}, err
	}

	applicationConnectionsValues := []SubaccountApplicationConnectionsData{}
	for _, connection := range value.Tunnel.ApplicationConnections {
		ac := SubaccountApplicationConnectionsData{
			ConnectionCount: types.Int64Value(connection.ConnectionCount),
			Name:            types.StringValue(connection.Name),
			Type:            types.StringValue(connection.Type),
		}

		applicationConnectionsValues = append(applicationConnectionsValues, ac)
	}

	applicationConnections, err := types.ListValueFrom(ctx, SubaccountApplicationConnectionsType, applicationConnectionsValues)
	if err.HasError() {
		return SubaccountConfig{}, err
	}

	serviceChannelsValues := []SubaccountServiceChannelsData{}
	for _, channel := range value.Tunnel.ServiceChannels {
		sc := SubaccountServiceChannelsData{
			Type:    types.StringValue(channel.Type),
			State:   types.StringValue(channel.State),
			Details: types.StringValue(channel.Details),
			Comment: types.StringValue(channel.Comment),
		}

		serviceChannelsValues = append(serviceChannelsValues, sc)
	}

	serviceChannels, err := types.ListValueFrom(ctx, SubaccountServiceChannelsType, serviceChannelsValues)
	if err.HasError() {
		return SubaccountConfig{}, err
	}

	tunnelObj := SubaccountTunnelData{
		State:                   types.StringValue(value.Tunnel.State),
		ConnectedSinceTimeStamp: types.Int64Value(value.Tunnel.ConnectedSinceTimeStamp),
		Connections:             types.Int64Value(value.Tunnel.Connections),
		User:                    types.StringValue(value.Tunnel.User),
		SubaccountCertificate:   certificate,
		ApplicationConnections:  applicationConnections,
		ServiceChannels:         serviceChannels,
	}

	tunnel, err := types.ObjectValueFrom(ctx, SubaccountTunnelType, tunnelObj)
	if err.HasError() {
		return SubaccountConfig{}, err
	}

	model := &SubaccountConfig{
		RegionHost:    types.StringValue(value.RegionHost),
		Subaccount:    types.StringValue(value.Subaccount),
		LocationID:    types.StringValue(value.LocationID),
		DisplayName:   types.StringValue(value.DisplayName),
		Description:   types.StringValue(value.Description),
		CloudUser:     plan.CloudUser,
		CloudPassword: plan.CloudPassword,
		Tunnel:        tunnel,
		Connected:     plan.Connected,
	}
	return *model, diag.Diagnostics{}
}

func SubaccountUsingAuthResourceValueFrom(ctx context.Context, plan SubaccountUsingAuthConfig, value apiobjects.SubaccountUsingAuthResource) (SubaccountUsingAuthConfig, diag.Diagnostics) {
	certificateObj := SubaccountCertificateData{
		NotAfterTimeStamp:  types.Int64Value(value.Tunnel.SubaccountCertificate.NotAfterTimeStamp),
		NotBeforeTimeStamp: types.Int64Value(value.Tunnel.SubaccountCertificate.NotBeforeTimeStamp),
		SubjectDN:          types.StringValue(value.Tunnel.SubaccountCertificate.SubjectDN),
		Issuer:             types.StringValue(value.Tunnel.SubaccountCertificate.Issuer),
		SerialNumber:       types.StringValue(value.Tunnel.SubaccountCertificate.SerialNumber),
	}

	certificate, err := types.ObjectValueFrom(ctx, SubaccountCertificateType, certificateObj)
	if err.HasError() {
		return SubaccountUsingAuthConfig{}, err
	}

	applicationConnectionsValues := []SubaccountApplicationConnectionsData{}
	for _, connection := range value.Tunnel.ApplicationConnections {
		ac := SubaccountApplicationConnectionsData{
			ConnectionCount: types.Int64Value(connection.ConnectionCount),
			Name:            types.StringValue(connection.Name),
			Type:            types.StringValue(connection.Type),
		}

		applicationConnectionsValues = append(applicationConnectionsValues, ac)
	}

	applicationConnections, err := types.ListValueFrom(ctx, SubaccountApplicationConnectionsType, applicationConnectionsValues)
	if err.HasError() {
		return SubaccountUsingAuthConfig{}, err
	}

	serviceChannelsValues := []SubaccountServiceChannelsData{}
	for _, channel := range value.Tunnel.ServiceChannels {
		sc := SubaccountServiceChannelsData{
			Type:    types.StringValue(channel.Type),
			State:   types.StringValue(channel.State),
			Details: types.StringValue(channel.Details),
			Comment: types.StringValue(channel.Comment),
		}

		serviceChannelsValues = append(serviceChannelsValues, sc)
	}

	serviceChannels, err := types.ListValueFrom(ctx, SubaccountServiceChannelsType, serviceChannelsValues)
	if err.HasError() {
		return SubaccountUsingAuthConfig{}, err
	}

	tunnelObj := SubaccountTunnelData{
		State:                   types.StringValue(value.Tunnel.State),
		ConnectedSinceTimeStamp: types.Int64Value(value.Tunnel.ConnectedSinceTimeStamp),
		Connections:             types.Int64Value(value.Tunnel.Connections),
		User:                    types.StringValue(value.Tunnel.User),
		SubaccountCertificate:   certificate,
		ApplicationConnections:  applicationConnections,
		ServiceChannels:         serviceChannels,
	}

	tunnel, err := types.ObjectValueFrom(ctx, SubaccountTunnelType, tunnelObj)
	if err.HasError() {
		return SubaccountUsingAuthConfig{}, err
	}

	model := &SubaccountUsingAuthConfig{
		RegionHost:         types.StringValue(value.RegionHost),
		Subaccount:         types.StringValue(value.Subaccount),
		AuthenticationData: plan.AuthenticationData,
		LocationID:         types.StringValue(value.LocationID),
		DisplayName:        types.StringValue(value.DisplayName),
		Description:        types.StringValue(value.Description),
		Tunnel:             tunnel,
		Connected:          plan.Connected,
	}
	return *model, diag.Diagnostics{}
}
