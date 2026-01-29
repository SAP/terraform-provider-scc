package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceDomainMapping(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "9f7390c8-f201-4b2d-b751-04c0a63c2671"
	virtualDomain := "testtfvirtualdomain"
	internalDomain := "testtfinternaldomain"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_domain_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceDomainMapping("scc_dm", regionHost, subaccount, virtualDomain, internalDomain),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_domain_mapping.scc_dm", "region_host", regionHost),
						resource.TestMatchResourceAttr("scc_domain_mapping.scc_dm", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_domain_mapping.scc_dm", "virtual_domain", virtualDomain),
						resource.TestCheckResourceAttr("scc_domain_mapping.scc_dm", "internal_domain", internalDomain),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_domain_mapping.scc_dm",
							map[string]knownvalue.Check{
								"internal_domain": knownvalue.StringExact(internalDomain),
								"region_host":     knownvalue.StringExact(regionHost),
								"subaccount":      knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceDomainMapping("scc_dm", "cf.us10.hana.ondemand.com", subaccount, "updatedtfvirtualdomain", internalDomain),
					ExpectError: regexp.MustCompile(`(?is)error updating the cloud connector domain mapping.*mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceDomainMapping("scc_dm", regionHost, subaccount, "updatedtfvirtualdomain", internalDomain),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_domain_mapping.scc_dm", "virtual_domain", "updatedtfvirtualdomain"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_domain_mapping.scc_dm",
							map[string]knownvalue.Check{
								"internal_domain": knownvalue.StringExact(internalDomain),
								"region_host":     knownvalue.StringExact(regionHost),
								"subaccount":      knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					ResourceName:                         "scc_domain_mapping.scc_dm",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubaccountEntitlement("scc_domain_mapping.scc_dm"),
					ImportStateVerifyIdentifierAttribute: "internal_domain",
				},
				{
					ResourceName:    "scc_domain_mapping.scc_dm",
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
				{
					ResourceName:  "scc_domain_mapping.scc_dm",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.comd3bbbcd7-d5e0-483b-a524-6dee7205f8e8testtfinternaldomain", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*internal_domain.*Got:`),
				},
				{
					ResourceName:  "scc_domain_mapping.scc_dm",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,d3bbbcd7-d5e0-483b-a524-6dee7205f8e8,testtfinternaldomain,extra",
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*internal_domain.*Got:`),
				},
			},
		})
	})

	t.Run("error path - region host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceDomainMappingWoRegionHost("scc_dm", subaccount, virtualDomain, internalDomain),
					ExpectError: regexp.MustCompile(`The argument "region_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - subaccount id mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceDomainMappingWoSubaccount("scc_dm", regionHost, virtualDomain, internalDomain),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - internal domain mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceDomainMappingWoInternalDomain("scc_dm", regionHost, subaccount, virtualDomain),
					ExpectError: regexp.MustCompile(`The argument "internal_domain" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - virtual domain mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceDomainMappingWoVirtualDomain("scc_dm", regionHost, subaccount, internalDomain),
					ExpectError: regexp.MustCompile(`The argument "virtual_domain" is required, but no definition was found.`),
				},
			},
		})
	})

}

func ResourceDomainMapping(datasourceName string, regionHost string, subaccount string, virtualDomain string, internalDomain string) string {
	return fmt.Sprintf(`
	resource "scc_domain_mapping" "%s" {
    region_host = "%s"
    subaccount = "%s"
    virtual_domain = "%s"
    internal_domain = "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualDomain, internalDomain)
}

func getImportStateForSubaccountEntitlement(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
			rs.Primary.Attributes["internal_domain"],
		), nil
	}
}

func ResourceDomainMappingWoRegionHost(datasourceName string, subaccount string, virtualDomain string, internalDomain string) string {
	return fmt.Sprintf(`
	resource "scc_domain_mapping" "%s" {
    subaccount = "%s"
    virtual_domain = "%s"
    internal_domain = "%s"
	}
	`, datasourceName, subaccount, virtualDomain, internalDomain)
}

func ResourceDomainMappingWoSubaccount(datasourceName string, regionHost string, virtualDomain string, internalDomain string) string {
	return fmt.Sprintf(`
	resource "scc_domain_mapping" "%s" {
    region_host = "%s"
    virtual_domain = "%s"
    internal_domain = "%s"
	}
	`, datasourceName, regionHost, virtualDomain, internalDomain)
}

func ResourceDomainMappingWoInternalDomain(datasourceName string, regionHost string, subaccount string, virtualDomain string) string {
	return fmt.Sprintf(`
	resource "scc_domain_mapping" "%s" {
    region_host = "%s"
    subaccount = "%s"
    virtual_domain = "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualDomain)
}

func ResourceDomainMappingWoVirtualDomain(datasourceName string, regionHost string, subaccount string, internalDomain string) string {
	return fmt.Sprintf(`
	resource "scc_domain_mapping" "%s" {
    region_host = "%s"
    subaccount = "%s"
    internal_domain = "%s"
	}
	`, datasourceName, regionHost, subaccount, internalDomain)
}
