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

func TestResourceSystemMapping(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_system_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				// CREATE
				{
					Config: providerConfig(user) + ResourceSystemMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_mapping.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "subaccount", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "virtual_host", "testtfvirtual"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "virtual_port", "900"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "internal_host", "testtfinternal"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "internal_port", "900"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "protocol", "HTTP"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "backend_type", "abapSys"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "authentication_mode", "KERBEROS"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_system_mapping.test",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact("testtfvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
							},
						),
					},
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSystemMapping("test", "cf.us10.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`(?is)error updating the cloud connector system mapping.*mismatched\s+configuration values`),
				},
				// UPDATE
				{
					Config: providerConfig(user) + ResourceSystemMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "updatedlocal", "905", "HTTPS", "hana", "INTERNAL", "X509_GENERAL"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_mapping.test", "internal_host", "updatedlocal"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "internal_port", "905"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "protocol", "HTTPS"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "backend_type", "hana"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "host_in_header", "INTERNAL"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "authentication_mode", "X509_GENERAL"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_system_mapping.test",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact("testtfvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
							},
						),
					},
				},

				// IMPORT
				{
					ResourceName:                         "scc_system_mapping.test",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSystemMapping("scc_system_mapping.test"),
					ImportStateVerifyIdentifierAttribute: "virtual_host",
				},
				{
					ResourceName:  "scc_system_mapping.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.comd3bbbcd7-d5e0-483b-a524-6dee7205f8e8testtfvirtual900", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*virtual_port.*Got:`),
				},
				{
					ResourceName:  "scc_system_mapping.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,d3bbbcd7-d5e0-483b-a524-6dee7205f8e8,testtfvirtual,900,extra",
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*virtual_port.*Got:`),
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
					Config:      ResourceSystemMappingWoRegionHost("test", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
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
					Config:      ResourceSystemMappingWoSubaccount("test", "cf.eu12.hana.ondemand.com", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
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
					Config:      ResourceSystemMappingWoVirtualHost("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
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
					Config:      ResourceSystemMappingWoVirtualPort("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "virtual_port" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - internal host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoInternalHost("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "internal_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - internal port mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoInternalPort("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "internal_port" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - protocol mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoProtocol("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "abapSys", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "protocol" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - backend type mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoBackendType("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "backend_type" is required, but no definition was found.`),
				},
			},
		})
	})

}

func ResourceSystemMapping(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	internalHost string, internalPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	internal_host= "%s"
	internal_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, internalHost, internalPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoRegionHost(datasourceName string, subaccount string, virtualHost string, virtualPort string,
	internalHost string, internalPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	internal_host= "%s"
	internal_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, subaccount, virtualHost, virtualPort, internalHost, internalPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoSubaccount(datasourceName string, regionHost string, virtualHost string, virtualPort string,
	internalHost string, internalPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	internal_host= "%s"
	internal_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, virtualHost, virtualPort, internalHost, internalPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoVirtualHost(datasourceName string, regionHost string, subaccount string, virtualPort string,
	internalHost string, internalPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_port= "%s"
	internal_host= "%s"
	internal_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualPort, internalHost, internalPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoVirtualPort(datasourceName string, regionHost string, subaccount string, virtualHost string,
	internalHost string, internalPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	internal_host= "%s"
	internal_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, internalHost, internalPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoInternalHost(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string, internalPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	internal_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, internalPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoInternalPort(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	internalHost string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	internal_host= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, internalHost, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoProtocol(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	internalHost string, internalPort string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	internal_host= "%s"
	internal_port= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, internalHost, internalPort, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoBackendType(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	internalHost string, internalPort string, protocol string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	internal_host= "%s"
	internal_port= "%s"
	protocol= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, internalHost, internalPort, protocol, hostInHeader, authenticationMode)
}

func getImportStateForSystemMapping(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
			rs.Primary.Attributes["virtual_host"],
			rs.Primary.Attributes["virtual_port"],
		), nil
	}
}
