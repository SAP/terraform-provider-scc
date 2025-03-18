package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemMappingCredentials struct {
	RegionHost types.String `tfsdk:"region_host"`
	Subaccount types.String `tfsdk:"subaccount"`
}

type SystemMappingsData struct {
	Credentials    SystemMappingCredentials `tfsdk:"credentials"`
	SystemMappings []SystemMapping          `tfsdk:"system_mappings"`
}

type SystemMappingData struct {
	Credentials   SystemMappingCredentials `tfsdk:"credentials"`
	SystemMapping SystemMapping            `tfsdk:"system_mapping"`
}

type SystemMapping struct {
	VirtualHost           types.String `tfsdk:"virtual_host"`
	VirtualPort           types.String `tfsdk:"virtual_port"`
	LocalHost             types.String `tfsdk:"local_host"`
	LocalPort             types.String `tfsdk:"local_port"`
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
}

func SystemMappingsValueFrom(ctx context.Context, plan SystemMappingsData, value apiobjects.SystemMappings) (SystemMappingsData, error) {
	system_mappings := []SystemMapping{}
	for _, mapping := range value.SystemMappings {
		c := SystemMapping{
			VirtualHost:           types.StringValue(mapping.VirtualHost),
			VirtualPort:           types.StringValue(mapping.VirtualPort),
			LocalHost:             types.StringValue(mapping.LocalHost),
			LocalPort:             types.StringValue(mapping.LocalPort),
			CreationDate:          types.StringValue(mapping.CreationDate),
			Protocol:              types.StringValue(mapping.Protocol),
			BackendType:           types.StringValue(mapping.BackendType),
			AuthenticationMode:    types.StringValue(mapping.AuthenticationMode),
			HostInHeader:          types.StringValue(mapping.HostInHeader),
			Sid:                   types.StringValue(mapping.Sid),
			TotalResourcesCount:   types.Int64Value(mapping.TotalResourcesCount),
			EnabledResourcesCount: types.Int64Value(mapping.TotalResourcesCount),
			Description:           types.StringValue(mapping.Description),
			SAPRouter:             types.StringValue(mapping.SAPRouter),
		}
		system_mappings = append(system_mappings, c)
	}

	systemMappingCredentials := SystemMappingCredentials{
		RegionHost: plan.Credentials.RegionHost,
		Subaccount: plan.Credentials.Subaccount,
	}

	model := &SystemMappingsData{
		Credentials:    systemMappingCredentials,
		SystemMappings: system_mappings,
	}
	return *model, nil
}

func SystemMappingValueFrom(ctx context.Context, plan SystemMappingData, value apiobjects.SystemMapping) (SystemMappingData, error) {
	systemMappingCredentials := SystemMappingCredentials{
		RegionHost: plan.Credentials.RegionHost,
		Subaccount: plan.Credentials.Subaccount,
	}

	systemMapping := SystemMapping{
		VirtualHost:           types.StringValue(value.VirtualHost),
		VirtualPort:           types.StringValue(value.VirtualPort),
		LocalHost:             types.StringValue(value.LocalHost),
		LocalPort:             types.StringValue(value.LocalPort),
		CreationDate:          types.StringValue(value.CreationDate),
		Protocol:              types.StringValue(value.Protocol),
		BackendType:           types.StringValue(value.BackendType),
		AuthenticationMode:    types.StringValue(value.AuthenticationMode),
		HostInHeader:          types.StringValue(value.HostInHeader),
		Sid:                   types.StringValue(value.Sid),
		TotalResourcesCount:   types.Int64Value(value.TotalResourcesCount),
		EnabledResourcesCount: types.Int64Value(value.EnabledResourcesCount),
		Description:           types.StringValue(value.Description),
		SAPRouter:             types.StringValue(value.SAPRouter),
	}

	model := &SystemMappingData{
		Credentials:   systemMappingCredentials,
		SystemMapping: systemMapping,
	}
	return *model, nil
}
