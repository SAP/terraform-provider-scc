package provider

import (
	"context"
	"strings"

	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemMappingResourceData struct {
	ID                      types.String `tfsdk:"id"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	ExactMatchOnly          types.Bool   `tfsdk:"exact_match_only"`
	WebsocketUpgradeAllowed types.Bool   `tfsdk:"websocket_upgrade_allowed"`
	CreationDate            types.String `tfsdk:"creation_date"`
	Description             types.String `tfsdk:"description"`
}

type SystemMappingResourceCredentials struct {
	RegionHost  types.String `tfsdk:"region_host"`
	Subaccount  types.String `tfsdk:"subaccount"`
	VirtualHost types.String `tfsdk:"virtual_host"`
	VirtualPort types.String `tfsdk:"virtual_port"`
}

type SystemMappingResourceDataSourceData struct {
	Credentials           SystemMappingResourceCredentials `tfsdk:"credentials"`
	SystemMappingResource SystemMappingResourceData        `tfsdk:"system_mapping_resource"`
}

type SystemMappingResourcesData struct {
	Credentials            SystemMappingResourceCredentials `tfsdk:"credentials"`
	SystemMappingResources []SystemMappingResourceData      `tfsdk:"system_mapping_resources"`
}

func SystemMappingResourceFrom(ctx context.Context, plan SystemMappingResourceDataSourceData, value apiobjects.SystemMappingResourceDataSource) (SystemMappingResourceDataSourceData, error) {
	system_mapping_resource := SystemMappingResourceData{
		ID:                      types.StringValue(value.SystemMappingResource.ID),
		Enabled:                 types.BoolValue(value.SystemMappingResource.Enabled),
		ExactMatchOnly:          types.BoolValue(value.SystemMappingResource.ExactMatchOnly),
		WebsocketUpgradeAllowed: types.BoolValue(value.SystemMappingResource.WebsocketUpgradeAllowed),
		CreationDate:            types.StringValue(value.SystemMappingResource.CreationDate),
		Description:             types.StringValue(value.SystemMappingResource.Description),
	}

	systemMappingResourceCredentials := SystemMappingResourceCredentials{
		RegionHost:  plan.Credentials.RegionHost,
		Subaccount:  plan.Credentials.Subaccount,
		VirtualHost: plan.Credentials.VirtualHost,
		VirtualPort: plan.Credentials.VirtualPort,
	}

	model := &SystemMappingResourceDataSourceData{
		Credentials:           systemMappingResourceCredentials,
		SystemMappingResource: system_mapping_resource,
	}

	return *model, nil
}

func SystemMappingResourcesFrom(ctx context.Context, plan SystemMappingResourcesData, value apiobjects.SystemMappingResources) (SystemMappingResourcesData, error) {
	system_mapping_resources := []SystemMappingResourceData{}
	for _, resource := range value.SystemMappingResources {
		r := SystemMappingResourceData{
			ID:                      types.StringValue(resource.ID),
			Enabled:                 types.BoolValue(resource.Enabled),
			ExactMatchOnly:          types.BoolValue(resource.ExactMatchOnly),
			WebsocketUpgradeAllowed: types.BoolValue(resource.WebsocketUpgradeAllowed),
			CreationDate:            types.StringValue(resource.CreationDate),
			Description:             types.StringValue(resource.Description),
		}
		system_mapping_resources = append(system_mapping_resources, r)
	}

	systemMappingResourceCredentials := SystemMappingResourceCredentials{
		RegionHost:  plan.Credentials.RegionHost,
		Subaccount:  plan.Credentials.Subaccount,
		VirtualHost: plan.Credentials.VirtualHost,
		VirtualPort: plan.Credentials.VirtualPort,
	}

	model := &SystemMappingResourcesData{
		Credentials:            systemMappingResourceCredentials,
		SystemMappingResources: system_mapping_resources,
	}

	return *model, nil
}

func CreateEncodedResourceID(input string) (encodedResourceID string) {
	input = strings.ReplaceAll(input, "+", "+2B")
	input = strings.ReplaceAll(input, "-", "+2D")
	input = strings.ReplaceAll(input, "/", "-")

	return input
}
