package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemMappingConfig struct {
	RegionHost            types.String `tfsdk:"region_host"`
	Subaccount            types.String `tfsdk:"subaccount"`
	VirtualHost           types.String `tfsdk:"virtual_host"`
	VirtualPort           types.String `tfsdk:"virtual_port"`
	InternalHost             types.String `tfsdk:"internal_host"`
	InternalPort             types.String `tfsdk:"internal_port"`
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

type SystemMappingsConfig struct {
	RegionHost     types.String    `tfsdk:"region_host"`
	Subaccount     types.String    `tfsdk:"subaccount"`
	SystemMappings []SystemMapping `tfsdk:"system_mappings"`
}

type SystemMapping struct {
	VirtualHost           types.String `tfsdk:"virtual_host"`
	VirtualPort           types.String `tfsdk:"virtual_port"`
	InternalHost             types.String `tfsdk:"internal_host"`
	InternalPort             types.String `tfsdk:"internal_port"`
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

func SystemMappingsValueFrom(ctx context.Context, plan SystemMappingsConfig, value apiobjects.SystemMappings) (SystemMappingsConfig, error) {
	system_mappings := []SystemMapping{}
	for _, mapping := range value.SystemMappings {
		c := SystemMapping{
			VirtualHost:           types.StringValue(mapping.VirtualHost),
			VirtualPort:           types.StringValue(mapping.VirtualPort),
			InternalHost:             types.StringValue(mapping.InternalHost),
			InternalPort:             types.StringValue(mapping.InternalPort),
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

	model := &SystemMappingsConfig{
		RegionHost:     plan.RegionHost,
		Subaccount:     plan.Subaccount,
		SystemMappings: system_mappings,
	}
	return *model, nil
}

func SystemMappingValueFrom(ctx context.Context, plan SystemMappingConfig, value apiobjects.SystemMapping) (SystemMappingConfig, error) {
	model := &SystemMappingConfig{
		RegionHost:            plan.RegionHost,
		Subaccount:            plan.Subaccount,
		VirtualHost:           types.StringValue(value.VirtualHost),
		VirtualPort:           types.StringValue(value.VirtualPort),
		InternalHost:             types.StringValue(value.InternalHost),
		InternalPort:             types.StringValue(value.InternalPort),
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

	return *model, nil
}
