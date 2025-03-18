package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DomainMappingData struct {
	VirtualDomain  types.String `tfsdk:"virtual_domain"`
	InternalDomain types.String `tfsdk:"internal_domain"`
}

type DomainMappingCredentials struct {
	RegionHost types.String `tfsdk:"region_host"`
	Subaccount types.String `tfsdk:"subaccount"`
}

type DomainMappingResourceData struct {
	Credentials   DomainMappingCredentials `tfsdk:"credentials"`
	DomainMapping DomainMappingData        `tfsdk:"domain_mapping"`
}

type DomainMappingsData struct {
	DomainMappingCredentials DomainMappingCredentials `tfsdk:"credentials"`
	DomainMappings           []DomainMappingData      `tfsdk:"domain_mappings"`
}

func DomainMappingsValueFrom(ctx context.Context, plan DomainMappingsData, value apiobjects.DomainMappings) (DomainMappingsData, error) {
	domain_mappings := []DomainMappingData{}
	for _, mappings := range value.DomainMappings {
		c := DomainMappingData{
			VirtualDomain:  types.StringValue(mappings.VirtualDomain),
			InternalDomain: types.StringValue(mappings.InternalDomain),
		}
		domain_mappings = append(domain_mappings, c)
	}

	domain_mapping_credentials := DomainMappingCredentials{
		RegionHost: plan.DomainMappingCredentials.RegionHost,
		Subaccount: plan.DomainMappingCredentials.Subaccount,
	}

	model := &DomainMappingsData{
		DomainMappingCredentials: domain_mapping_credentials,
		DomainMappings:           domain_mappings,
	}

	return *model, nil
}

func DomainMappingValueFrom(ctx context.Context, plan DomainMappingResourceData, value apiobjects.DomainMapping) (DomainMappingResourceData, error) {
	domain_mapping := DomainMappingData{
		VirtualDomain:  types.StringValue(value.VirtualDomain),
		InternalDomain: types.StringValue(value.InternalDomain),
	}

	domain_mapping_credentials := DomainMappingCredentials{
		RegionHost: plan.Credentials.RegionHost,
		Subaccount: plan.Credentials.Subaccount,
	}

	model := &DomainMappingResourceData{
		Credentials:   domain_mapping_credentials,
		DomainMapping: domain_mapping,
	}

	return *model, nil
}
