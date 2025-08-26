package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DomainMappingConfig struct {
	RegionHost     types.String `tfsdk:"region_host"`
	Subaccount     types.String `tfsdk:"subaccount"`
	VirtualDomain  types.String `tfsdk:"virtual_domain"`
	InternalDomain types.String `tfsdk:"internal_domain"`
}

type DomainMapping struct {
	VirtualDomain  types.String `tfsdk:"virtual_domain"`
	InternalDomain types.String `tfsdk:"internal_domain"`
}

type DomainMappingsConfig struct {
	RegionHost     types.String    `tfsdk:"region_host"`
	Subaccount     types.String    `tfsdk:"subaccount"`
	DomainMappings []DomainMapping `tfsdk:"domain_mappings"`
}

func DomainMappingsValueFrom(ctx context.Context, plan DomainMappingsConfig, value apiobjects.DomainMappings) (DomainMappingsConfig, diag.Diagnostics) {
	domain_mappings := []DomainMapping{}
	for _, mappings := range value.DomainMappings {
		c := DomainMapping{
			VirtualDomain:  types.StringValue(mappings.VirtualDomain),
			InternalDomain: types.StringValue(mappings.InternalDomain),
		}
		domain_mappings = append(domain_mappings, c)
	}

	model := &DomainMappingsConfig{
		RegionHost:     plan.RegionHost,
		Subaccount:     plan.Subaccount,
		DomainMappings: domain_mappings,
	}

	return *model, diag.Diagnostics{}
}

func DomainMappingValueFrom(ctx context.Context, plan DomainMappingConfig, value apiobjects.DomainMapping) (DomainMappingConfig, diag.Diagnostics) {
	model := &DomainMappingConfig{
		RegionHost:     plan.RegionHost,
		Subaccount:     plan.Subaccount,
		VirtualDomain:  types.StringValue(value.VirtualDomain),
		InternalDomain: types.StringValue(value.InternalDomain),
	}
	return *model, diag.Diagnostics{}
}

func GetDomainMapping(domainMappings apiobjects.DomainMappings, targetInternalDomain string) (*apiobjects.DomainMapping, diag.Diagnostics) {
	var diags diag.Diagnostics
	for _, mapping := range domainMappings.DomainMappings {
		if mapping.InternalDomain == targetInternalDomain {
			return &mapping, diags
		}
	}
	diags.AddError("Mapping doesn't exist", "The specified mapping doesn't exist.")
	return nil, diags
}
