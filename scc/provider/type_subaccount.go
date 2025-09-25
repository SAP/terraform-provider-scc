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
	State                  types.String `tfsdk:"state"`
	ConnectedSince         types.String `tfsdk:"connected_since"`
	Connections            types.Int64  `tfsdk:"connections"`
	SubaccountCertificate  types.Object `tfsdk:"subaccount_certificate"`
	User                   types.String `tfsdk:"user"`
	ApplicationConnections types.List   `tfsdk:"application_connections"`
	ServiceChannels        types.List   `tfsdk:"service_channels"`
}

var SubaccountTunnelType = map[string]attr.Type{
	"state":           types.StringType,
	"connected_since": types.StringType,
	"connections":     types.Int64Type,
	"user":            types.StringType,
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
	ValidTo      types.String `tfsdk:"valid_to"`
	ValidFrom    types.String `tfsdk:"valid_from"`
	SubjectDN    types.String `tfsdk:"subject_dn"`
	Issuer       types.String `tfsdk:"issuer"`
	SerialNumber types.String `tfsdk:"serial_number"`
}

var SubaccountCertificateType = map[string]attr.Type{
	"valid_to":      types.StringType,
	"valid_from":    types.StringType,
	"subject_dn":    types.StringType,
	"issuer":        types.StringType,
	"serial_number": types.StringType,
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
	RegionHost          types.String `tfsdk:"region_host"`
	Subaccount          types.String `tfsdk:"subaccount"`
	CloudUser           types.String `tfsdk:"cloud_user"`
	CloudPassword       types.String `tfsdk:"cloud_password"`
	LocationID          types.String `tfsdk:"location_id"`
	DisplayName         types.String `tfsdk:"display_name"`
	Description         types.String `tfsdk:"description"`
	Tunnel              types.Object `tfsdk:"tunnel"`
	Connected           types.Bool   `tfsdk:"connected"`
	AutoRenewBeforeDays types.Int64  `tfsdk:"auto_renew_before_days"`
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
		ValidTo:      ConvertMillisToTimes(value.Tunnel.SubaccountCertificate.NotAfterTimeStamp).WithTimezone,
		ValidFrom:    ConvertMillisToTimes(value.Tunnel.SubaccountCertificate.NotBeforeTimeStamp).WithTimezone,
		SubjectDN:    types.StringValue(value.Tunnel.SubaccountCertificate.SubjectDN),
		Issuer:       types.StringValue(value.Tunnel.SubaccountCertificate.Issuer),
		SerialNumber: types.StringValue(value.Tunnel.SubaccountCertificate.SerialNumber),
	}

	certificate, diags := types.ObjectValueFrom(ctx, SubaccountCertificateType, certificateObj)
	if diags.HasError() {
		return SubaccountData{}, diags
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

	applicationConnections, diags := types.ListValueFrom(ctx, SubaccountApplicationConnectionsType, applicationConnectionsValues)
	if diags.HasError() {
		return SubaccountData{}, diags
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

	serviceChannels, diags := types.ListValueFrom(ctx, SubaccountServiceChannelsType, serviceChannelsValues)
	if diags.HasError() {
		return SubaccountData{}, diags
	}

	tunnelObj := SubaccountTunnelData{
		State:                  types.StringValue(value.Tunnel.State),
		ConnectedSince:         ConvertMillisToTimes(value.Tunnel.ConnectedSinceTimeStamp).WithTimezone,
		Connections:            types.Int64Value(value.Tunnel.Connections),
		User:                   types.StringValue(value.Tunnel.User),
		SubaccountCertificate:  certificate,
		ApplicationConnections: applicationConnections,
		ServiceChannels:        serviceChannels,
	}

	tunnel, diags := types.ObjectValueFrom(ctx, SubaccountTunnelType, tunnelObj)
	if diags.HasError() {
		return SubaccountData{}, diags
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
		ValidTo:      ConvertMillisToTimes(value.Tunnel.SubaccountCertificate.NotAfterTimeStamp).WithTimezone,
		ValidFrom:    ConvertMillisToTimes(value.Tunnel.SubaccountCertificate.NotBeforeTimeStamp).WithTimezone,
		SubjectDN:    types.StringValue(value.Tunnel.SubaccountCertificate.SubjectDN),
		Issuer:       types.StringValue(value.Tunnel.SubaccountCertificate.Issuer),
		SerialNumber: types.StringValue(value.Tunnel.SubaccountCertificate.SerialNumber),
	}

	certificate, diags := types.ObjectValueFrom(ctx, SubaccountCertificateType, certificateObj)
	if diags.HasError() {
		return SubaccountConfig{}, diags
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

	applicationConnections, diags := types.ListValueFrom(ctx, SubaccountApplicationConnectionsType, applicationConnectionsValues)
	if diags.HasError() {
		return SubaccountConfig{}, diags
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

	serviceChannels, diags := types.ListValueFrom(ctx, SubaccountServiceChannelsType, serviceChannelsValues)
	if diags.HasError() {
		return SubaccountConfig{}, diags
	}

	tunnelObj := SubaccountTunnelData{
		State:                  types.StringValue(value.Tunnel.State),
		ConnectedSince:         ConvertMillisToTimes(value.Tunnel.ConnectedSinceTimeStamp).WithTimezone,
		Connections:            types.Int64Value(value.Tunnel.Connections),
		User:                   types.StringValue(value.Tunnel.User),
		SubaccountCertificate:  certificate,
		ApplicationConnections: applicationConnections,
		ServiceChannels:        serviceChannels,
	}

	tunnel, diags := types.ObjectValueFrom(ctx, SubaccountTunnelType, tunnelObj)
	if diags.HasError() {
		return SubaccountConfig{}, diags
	}

	model := &SubaccountConfig{
		RegionHost:          types.StringValue(value.RegionHost),
		Subaccount:          types.StringValue(value.Subaccount),
		LocationID:          types.StringValue(value.LocationID),
		DisplayName:         types.StringValue(value.DisplayName),
		Description:         types.StringValue(value.Description),
		CloudUser:           plan.CloudUser,
		CloudPassword:       plan.CloudPassword,
		Tunnel:              tunnel,
		Connected:           plan.Connected,
		AutoRenewBeforeDays: plan.AutoRenewBeforeDays,
	}
	return *model, diag.Diagnostics{}
}

func SubaccountUsingAuthResourceValueFrom(ctx context.Context, plan SubaccountUsingAuthConfig, value apiobjects.SubaccountUsingAuthResource) (SubaccountUsingAuthConfig, diag.Diagnostics) {
	certificateObj := SubaccountCertificateData{
		ValidTo:      ConvertMillisToTimes(value.Tunnel.SubaccountCertificate.NotAfterTimeStamp).WithTimezone,
		ValidFrom:    ConvertMillisToTimes(value.Tunnel.SubaccountCertificate.NotBeforeTimeStamp).WithTimezone,
		SubjectDN:    types.StringValue(value.Tunnel.SubaccountCertificate.SubjectDN),
		Issuer:       types.StringValue(value.Tunnel.SubaccountCertificate.Issuer),
		SerialNumber: types.StringValue(value.Tunnel.SubaccountCertificate.SerialNumber),
	}

	certificate, diags := types.ObjectValueFrom(ctx, SubaccountCertificateType, certificateObj)
	if diags.HasError() {
		return SubaccountUsingAuthConfig{}, diags
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

	applicationConnections, diags := types.ListValueFrom(ctx, SubaccountApplicationConnectionsType, applicationConnectionsValues)
	if diags.HasError() {
		return SubaccountUsingAuthConfig{}, diags
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

	serviceChannels, diags := types.ListValueFrom(ctx, SubaccountServiceChannelsType, serviceChannelsValues)
	if diags.HasError() {
		return SubaccountUsingAuthConfig{}, diags
	}

	tunnelObj := SubaccountTunnelData{
		State:                  types.StringValue(value.Tunnel.State),
		ConnectedSince:         ConvertMillisToTimes(value.Tunnel.ConnectedSinceTimeStamp).WithTimezone,
		Connections:            types.Int64Value(value.Tunnel.Connections),
		User:                   types.StringValue(value.Tunnel.User),
		SubaccountCertificate:  certificate,
		ApplicationConnections: applicationConnections,
		ServiceChannels:        serviceChannels,
	}

	tunnel, diags := types.ObjectValueFrom(ctx, SubaccountTunnelType, tunnelObj)
	if diags.HasError() {
		return SubaccountUsingAuthConfig{}, diags
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
