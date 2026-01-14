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

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_domain_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceDomainMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualdomain", "testtfinternaldomain"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_domain_mapping.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("scc_domain_mapping.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_domain_mapping.test", "virtual_domain", "testtfvirtualdomain"),
						resource.TestCheckResourceAttr("scc_domain_mapping.test", "internal_domain", "testtfinternaldomain"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_domain_mapping.test",
							map[string]knownvalue.Check{
								"internal_domain": knownvalue.StringExact("testtfinternaldomain"),
								"region_host":     knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":      knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceDomainMapping("test", "cf.us10.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "updatedtfvirtualdomain", "testtfinternaldomain"),
					ExpectError: regexp.MustCompile(`(?is)error updating the cloud connector domain mapping.*mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceDomainMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "updatedtfvirtualdomain", "testtfinternaldomain"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_domain_mapping.test", "virtual_domain", "updatedtfvirtualdomain"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_domain_mapping.test",
							map[string]knownvalue.Check{
								"internal_domain": knownvalue.StringExact("testtfinternaldomain"),
								"region_host":     knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":      knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					ResourceName:                         "scc_domain_mapping.test",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubaccountEntitlement("scc_domain_mapping.test"),
					ImportStateVerifyIdentifierAttribute: "internal_domain",
				},
				{
					ResourceName:    "scc_domain_mapping.test",
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
				{
					ResourceName:  "scc_domain_mapping.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.comd3bbbcd7-d5e0-483b-a524-6dee7205f8e8testtfinternaldomain", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*internal_domain.*Got:`),
				},
				{
					ResourceName:  "scc_domain_mapping.test",
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
					Config:      ResourceDomainMappingWoRegionHost("test", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualdomain", "testtfinternaldomain"),
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
					Config:      ResourceDomainMappingWoSubaccount("test", "cf.eu12.hana.ondemand.com", "testtfvirtualdomain", "testtfinternaldomain"),
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
					Config:      ResourceDomainMappingWoInternalDomain("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualdomain"),
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
					Config:      ResourceDomainMappingWoVirtualDomain("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfinternaldomain"),
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
