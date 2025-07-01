package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
						resource.TestCheckResourceAttr("scc_system_mapping.test", "local_host", "testtfinternal"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "local_port", "900"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "protocol", "HTTP"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "backend_type", "abapSys"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "authentication_mode", "KERBEROS"),
					),
				},

				// UPDATE
				{
					Config: providerConfig(user) + ResourceSystemMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "updatedlocal", "905", "HTTPS", "hana", "INTERNAL", "X509_GENERAL"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_mapping.test", "local_host", "updatedlocal"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "local_port", "905"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "protocol", "HTTPS"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "backend_type", "hana"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "host_in_header", "INTERNAL"),
						resource.TestCheckResourceAttr("scc_system_mapping.test", "authentication_mode", "X509_GENERAL"),
					),
				},

				// IMPORT
				{
					ResourceName:                         "scc_system_mapping.test",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSystemMapping("scc_system_mapping.test"),
					ImportStateVerifyIdentifierAttribute: "virtual_host",
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

	t.Run("error path - local host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoLocalHost("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "local_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - local port mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoLocalPort("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "local_port" is required, but no definition was found.`),
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

	t.Run("error path - host in header mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoHostInHeader("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "KERBEROS"),
					ExpectError: regexp.MustCompile(`The argument "host_in_header" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - authentication mode mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemMappingWoAuthMode("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL"),
					ExpectError: regexp.MustCompile(`The argument "authentication_mode" is required, but no definition was found.`),
				},
			},
		})
	})

}

func ResourceSystemMapping(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, localPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoRegionHost(datasourceName string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, subaccount, virtualHost, virtualPort, localHost, localPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoSubaccount(datasourceName string, regionHost string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, virtualHost, virtualPort, localHost, localPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoVirtualHost(datasourceName string, regionHost string, subaccount string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualPort, localHost, localPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoVirtualPort(datasourceName string, regionHost string, subaccount string, virtualHost string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, localHost, localPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoLocalHost(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localPort, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoLocalPort(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, protocol, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoProtocol(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, localPort, backendType, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoBackendType(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, localPort, protocol, hostInHeader, authenticationMode)
}

func ResourceSystemMappingWoHostInHeader(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, localPort, protocol, backendType, authenticationMode)
}

func ResourceSystemMappingWoAuthMode(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, localPort, protocol, backendType, hostInHeader)
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
