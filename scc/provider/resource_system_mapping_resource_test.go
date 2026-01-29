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

func TestResourceSystemMappingResource(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "9f7390c8-f201-4b2d-b751-04c0a63c2671"
	virtualHost := "testtfvirtualtesting"
	virtualPort := "90"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_system_mapping_resource")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSystemMappingResource("scc_smr", regionHost, subaccount, virtualHost, virtualPort, "/", "create resource", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_mapping_resource.scc_smr", "description", "create resource"),
						resource.TestCheckResourceAttr("scc_system_mapping_resource.scc_smr", "enabled", "true"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_system_mapping_resource.scc_smr",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact(virtualHost),
								"virtual_port": knownvalue.StringExact(virtualPort),
								"url_path":     knownvalue.StringExact("/"),
							},
						),
					},
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSystemMappingResource("scc_smr", "cf.us10.hana.ondemand.com", subaccount, virtualHost, virtualPort, "/", "create resource", true),
					ExpectError: regexp.MustCompile(`(?is)error updating the cloud connector system mapping resource.*mismatched\s+configuration values`),
				},
				{
					// 🚀 This is the update step
					Config: providerConfig(user) + ResourceSystemMappingResource("scc_smr", regionHost, subaccount, virtualHost, virtualPort, "/", "updated resource", false),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_mapping_resource.scc_smr", "description", "updated resource"),
						resource.TestCheckResourceAttr("scc_system_mapping_resource.scc_smr", "enabled", "false"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_system_mapping_resource.scc_smr",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact(virtualHost),
								"virtual_port": knownvalue.StringExact(virtualPort),
								"url_path":     knownvalue.StringExact("/"),
							},
						),
					},
				},
				{
					ResourceName:                         "scc_system_mapping_resource.scc_smr",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateVerifyIdentifierAttribute: "region_host",
					ImportStateIdFunc:                    getImportStateForSystemMappingResource("scc_system_mapping_resource.scc_smr"),
				},
				{
					ResourceName:    "scc_system_mapping_resource.scc_smr",
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
				{
					ResourceName:  "scc_system_mapping_resource.scc_smr",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com9f7390c8-f201-4b2d-b751-04c0a63c2671testtfvirtualtesting90/", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*url_path.*Got:`),
				},
				{
					ResourceName:  "scc_system_mapping_resource.scc_smr",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,9f7390c8-f201-4b2d-b751-04c0a63c2671,testtfvirtualtesting,90,/,extra",
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*url_path.*Got:`),
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
					Config:      ResourceSystemMappingResourceWoRegionHost("scc_smr", subaccount, virtualHost, virtualPort, "/", "create resource", true),
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
					Config:      ResourceSystemMappingResourceWoSubaccount("scc_smr", regionHost, virtualHost, virtualPort, "/", "create resource", true),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - virtual host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingResourceWoVirtualHost("scc_smr", regionHost, subaccount, virtualPort, "/", "create resource", true),
					ExpectError: regexp.MustCompile(`The argument "virtual_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - virtual port mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingResourceWoVirtualPort("scc_smr", regionHost, subaccount, virtualHost, "/", "create resource", true),
					ExpectError: regexp.MustCompile(`The argument "virtual_port" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - resource id mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingResourceWoURLPath("scc_smr", regionHost, subaccount, virtualHost, virtualPort, "create resource", true),
					ExpectError: regexp.MustCompile(`The argument "url_path" is required, but no definition was found.`),
				},
			},
		})
	})

}

func ResourceSystemMappingResource(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	urlPath string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	subaccount = "%s"
	virtual_host = "%s"
	virtual_port = "%s"
	url_path = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, urlPath, description, enabled)
}

func ResourceSystemMappingResourceWoRegionHost(datasourceName string, subaccount string, virtualHost string, virtualPort string,
	urlPath string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	subaccount = "%s"
	virtual_host = "%s"
	virtual_port = "%s"
	url_path = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, subaccount, virtualHost, virtualPort, urlPath, description, enabled)
}

func ResourceSystemMappingResourceWoSubaccount(datasourceName string, regionHost string, virtualHost string, virtualPort string,
	urlPath string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	virtual_host = "%s"
	virtual_port = "%s"
	url_path = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, virtualHost, virtualPort, urlPath, description, enabled)
}

func ResourceSystemMappingResourceWoVirtualHost(datasourceName string, regionHost string, subaccount string, virtualPort string,
	urlPath string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	subaccount = "%s"
	virtual_port = "%s"
	url_path = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, subaccount, virtualPort, urlPath, description, enabled)
}

func ResourceSystemMappingResourceWoVirtualPort(datasourceName string, regionHost string, subaccount string, virtualHost string,
	urlPath string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	subaccount = "%s"
	virtual_host = "%s"
	url_path = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, urlPath, description, enabled)
}

func ResourceSystemMappingResourceWoURLPath(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	subaccount = "%s"
	virtual_host = "%s"
	virtual_port = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, description, enabled)
}
func getImportStateForSystemMappingResource(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s,%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
			rs.Primary.Attributes["virtual_host"],
			rs.Primary.Attributes["virtual_port"],
			rs.Primary.Attributes["url_path"],
		), nil
	}
}
