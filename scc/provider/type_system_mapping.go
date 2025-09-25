package provider

import (
	"context"
	"strings"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemMappingConfig struct {
	RegionHost            types.String `tfsdk:"region_host"`
	Subaccount            types.String `tfsdk:"subaccount"`
	VirtualHost           types.String `tfsdk:"virtual_host"`
	VirtualPort           types.String `tfsdk:"virtual_port"`
	InternalHost          types.String `tfsdk:"internal_host"`
	InternalPort          types.String `tfsdk:"internal_port"`
	CreationDate          types.String `tfsdk:"creation_date"`
	Protocol              types.String `tfsdk:"protocol"`
	BackendType           types.String `tfsdk:"backend_type"`
	AuthenticationMode    types.String `tfsdk:"authentication_mode"`
	HostInHeader          types.String `tfsdk:"host_in_header"`
	Sid                   types.String `tfsdk:"sid"`
	TotalResourcesCount   types.Int64  `tfsdk:"total_resources_count"`
	EnabledResourcesCount types.Int64  `tfsdk:"enabled_resources_count"`
	Description           types.String `tfsdk:"description"`
	SAPRouter             types.String `tfsdk:"sap_router"`
	SNCPartnerName        types.String `tfsdk:"snc_partner_name"`
	AllowedClients        types.List   `tfsdk:"allowed_clients"`
	BlacklistedUsers      types.List   `tfsdk:"blacklisted_users"`
}

type SystemMappingsConfig struct {
	RegionHost     types.String    `tfsdk:"region_host"`
	Subaccount     types.String    `tfsdk:"subaccount"`
	SystemMappings []SystemMapping `tfsdk:"system_mappings"`
}

type SystemMapping struct {
	VirtualHost           types.String `tfsdk:"virtual_host"`
	VirtualPort           types.String `tfsdk:"virtual_port"`
	InternalHost          types.String `tfsdk:"internal_host"`
	InternalPort          types.String `tfsdk:"internal_port"`
	CreationDate          types.String `tfsdk:"creation_date"`
	Protocol              types.String `tfsdk:"protocol"`
	BackendType           types.String `tfsdk:"backend_type"`
	AuthenticationMode    types.String `tfsdk:"authentication_mode"`
	HostInHeader          types.String `tfsdk:"host_in_header"`
	Sid                   types.String `tfsdk:"sid"`
	TotalResourcesCount   types.Int64  `tfsdk:"total_resources_count"`
	EnabledResourcesCount types.Int64  `tfsdk:"enabled_resources_count"`
	Description           types.String `tfsdk:"description"`
	SAPRouter             types.String `tfsdk:"sap_router"`
	SNCPartnerName        types.String `tfsdk:"snc_partner_name"`
	AllowedClients        types.List   `tfsdk:"allowed_clients"`
	BlacklistedUsers      types.List   `tfsdk:"blacklisted_users"`
}

type SystemMappingBlacklistedUsersData struct {
	Client types.String `tfsdk:"client"`
	User   types.String `tfsdk:"user"`
}

var SystemMappingBlacklistedUsersType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"client": types.StringType,
		"user":   types.StringType,
	},
}

func SystemMappingsValueFrom(ctx context.Context, plan SystemMappingsConfig, value apiobjects.SystemMappings) (SystemMappingsConfig, diag.Diagnostics) {
	system_mappings := []SystemMapping{}
	for _, mapping := range value.SystemMappings {
		blacklistedUsersValue := []SystemMappingBlacklistedUsersData{}
		for _, user := range mapping.BlacklistedUsers {
			bl := SystemMappingBlacklistedUsersData{
				Client: types.StringValue(user.Client),
				User:   types.StringValue(user.User),
			}
			blacklistedUsersValue = append(blacklistedUsersValue, bl)
		}
		blacklistedUsers, diags := types.ListValueFrom(ctx, SystemMappingBlacklistedUsersType, blacklistedUsersValue)
		if diags.HasError() {
			return SystemMappingsConfig{}, diags
		}

		allowedClients, diags := types.ListValueFrom(ctx, types.StringType, mapping.AllowedClients)
		if diags.HasError() {
			return SystemMappingsConfig{}, diags
		}

		c := SystemMapping{
			VirtualHost:           types.StringValue(mapping.VirtualHost),
			VirtualPort:           types.StringValue(mapping.VirtualPort),
			InternalHost:          types.StringValue(mapping.InternalHost),
			InternalPort:          types.StringValue(mapping.InternalPort),
			CreationDate:          ConvertMillisToTimes(mapping.CreationDate).UTC,
			Protocol:              types.StringValue(mapping.Protocol),
			BackendType:           types.StringValue(mapping.BackendType),
			AuthenticationMode:    types.StringValue(mapping.AuthenticationMode),
			HostInHeader:          types.StringValue(mapping.HostInHeader),
			Sid:                   types.StringValue(mapping.Sid),
			TotalResourcesCount:   types.Int64Value(mapping.TotalResourcesCount),
			EnabledResourcesCount: types.Int64Value(mapping.TotalResourcesCount),
			Description:           types.StringValue(mapping.Description),
			SAPRouter:             types.StringValue(mapping.SAPRouter),
			SNCPartnerName:        types.StringValue(mapping.SNCPartnerName),
			AllowedClients:        allowedClients,
			BlacklistedUsers:      blacklistedUsers,
		}
		system_mappings = append(system_mappings, c)
	}

	model := &SystemMappingsConfig{
		RegionHost:     plan.RegionHost,
		Subaccount:     plan.Subaccount,
		SystemMappings: system_mappings,
	}
	return *model, diag.Diagnostics{}
}

func SystemMappingValueFrom(ctx context.Context, plan SystemMappingConfig, value apiobjects.SystemMapping) (SystemMappingConfig, diag.Diagnostics) {
	blacklistedUsersValue := []SystemMappingBlacklistedUsersData{}
	for _, user := range value.BlacklistedUsers {
		bl := SystemMappingBlacklistedUsersData{
			Client: types.StringValue(user.Client),
			User:   types.StringValue(user.User),
		}
		blacklistedUsersValue = append(blacklistedUsersValue, bl)
	}
	blacklistedUsers, diags := types.ListValueFrom(ctx, SystemMappingBlacklistedUsersType, blacklistedUsersValue)
	if diags.HasError() {
		return SystemMappingConfig{}, diags
	}

	allowedClients, diags := types.ListValueFrom(ctx, types.StringType, value.AllowedClients)
	if diags.HasError() {
		return SystemMappingConfig{}, diags
	}

	hostInHeader := types.StringValue(value.HostInHeader)
	if !plan.HostInHeader.IsNull() && !plan.HostInHeader.IsUnknown() {
		if hostInHeader.ValueString() != strings.ToLower(plan.HostInHeader.ValueString()) {
			hostInHeader = plan.HostInHeader
		}
	}

	model := &SystemMappingConfig{
		RegionHost:            plan.RegionHost,
		Subaccount:            plan.Subaccount,
		VirtualHost:           types.StringValue(value.VirtualHost),
		VirtualPort:           types.StringValue(value.VirtualPort),
		InternalHost:          types.StringValue(value.InternalHost),
		InternalPort:          types.StringValue(value.InternalPort),
		CreationDate:          ConvertMillisToTimes(value.CreationDate).UTC,
		Protocol:              types.StringValue(value.Protocol),
		BackendType:           types.StringValue(value.BackendType),
		AuthenticationMode:    types.StringValue(value.AuthenticationMode),
		HostInHeader:          hostInHeader,
		Sid:                   types.StringValue(value.Sid),
		TotalResourcesCount:   types.Int64Value(value.TotalResourcesCount),
		EnabledResourcesCount: types.Int64Value(value.EnabledResourcesCount),
		Description:           types.StringValue(value.Description),
		SAPRouter:             types.StringValue(value.SAPRouter),
		SNCPartnerName:        types.StringValue(value.SNCPartnerName),
		AllowedClients:        allowedClients,
		BlacklistedUsers:      blacklistedUsers,
	}

	return *model, diag.Diagnostics{}
}
