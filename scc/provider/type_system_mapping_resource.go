package provider

import (
	"context"
	"strings"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemMappingResourceData struct {
	URLPath                 types.String `tfsdk:"url_path"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	PathOnly                types.Bool   `tfsdk:"path_only"`
	WebsocketUpgradeAllowed types.Bool   `tfsdk:"websocket_upgrade_allowed"`
	CreationDate            types.String `tfsdk:"creation_date"`
	Description             types.String `tfsdk:"description"`
}

type SystemMappingResourceConfig struct {
	RegionHost              types.String `tfsdk:"region_host"`
	Subaccount              types.String `tfsdk:"subaccount"`
	VirtualHost             types.String `tfsdk:"virtual_host"`
	VirtualPort             types.String `tfsdk:"virtual_port"`
	URLPath                 types.String `tfsdk:"url_path"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	PathOnly                types.Bool   `tfsdk:"path_only"`
	WebsocketUpgradeAllowed types.Bool   `tfsdk:"websocket_upgrade_allowed"`
	CreationDate            types.String `tfsdk:"creation_date"`
	Description             types.String `tfsdk:"description"`
}

type SystemMappingResourcesConfig struct {
	RegionHost             types.String                `tfsdk:"region_host"`
	Subaccount             types.String                `tfsdk:"subaccount"`
	VirtualHost            types.String                `tfsdk:"virtual_host"`
	VirtualPort            types.String                `tfsdk:"virtual_port"`
	SystemMappingResources []SystemMappingResourceData `tfsdk:"system_mapping_resources"`
}

func SystemMappingResourceValueFrom(ctx context.Context, plan SystemMappingResourceConfig, value apiobjects.SystemMappingResource) (SystemMappingResourceConfig, diag.Diagnostics) {
	model := &SystemMappingResourceConfig{
		RegionHost:              plan.RegionHost,
		Subaccount:              plan.Subaccount,
		VirtualHost:             plan.VirtualHost,
		VirtualPort:             plan.VirtualPort,
		URLPath:                 types.StringValue(value.URLPath),
		Enabled:                 types.BoolValue(value.Enabled),
		PathOnly:                types.BoolValue(value.PathOnly),
		WebsocketUpgradeAllowed: types.BoolValue(value.WebsocketUpgradeAllowed),
		CreationDate:            ConvertMillisToTimes(value.CreationDate).UTC,
		Description:             types.StringValue(value.Description),
	}

	return *model, diag.Diagnostics{}
}

func SystemMappingResourcesValueFrom(ctx context.Context, plan SystemMappingResourcesConfig, value apiobjects.SystemMappingResources) (SystemMappingResourcesConfig, diag.Diagnostics) {
	system_mapping_resources := []SystemMappingResourceData{}
	for _, smr := range value.SystemMappingResources {
		r := SystemMappingResourceData{
			URLPath:                 types.StringValue(smr.URLPath),
			Enabled:                 types.BoolValue(smr.Enabled),
			PathOnly:                types.BoolValue(smr.PathOnly),
			WebsocketUpgradeAllowed: types.BoolValue(smr.WebsocketUpgradeAllowed),
			CreationDate:            ConvertMillisToTimes(smr.CreationDate).UTC,
			Description:             types.StringValue(smr.Description),
		}
		system_mapping_resources = append(system_mapping_resources, r)
	}

	model := &SystemMappingResourcesConfig{
		RegionHost:             plan.RegionHost,
		Subaccount:             plan.Subaccount,
		VirtualHost:            plan.VirtualHost,
		VirtualPort:            plan.VirtualPort,
		SystemMappingResources: system_mapping_resources,
	}

	return *model, diag.Diagnostics{}
}

/*
CreateEncodedResourceID encodes the given resource ID to make it safe for use in a URI path.

According to the encoding rules, it replaces specific characters to avoid collisions:
- '+' is replaced with '+2B'
- '-' is replaced with '+2D'
- '/' is replaced with '-'

This ensures the resource ID can be safely used in URI paths without misinterpretation.
*/
func CreateEncodedResourceID(input string) (encodedResourceID string) {
	input = strings.ReplaceAll(input, "+", "+2B")
	input = strings.ReplaceAll(input, "-", "+2D")
	input = strings.ReplaceAll(input, "/", "-")

	return input
}
