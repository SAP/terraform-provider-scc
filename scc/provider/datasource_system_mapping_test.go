package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMapping(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSystemMapping("test", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8", "testterraformvirtual", "900"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.scc_system_mapping.test", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "virtual_host", "testterraformvirtual"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "virtual_port", "900"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "local_host", "testterraforminternal"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "local_port", "900"),
						resource.TestMatchResourceAttr("data.scc_system_mapping.test", "creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "protocol", "HTTP"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "backend_type", "abapSys"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "authentication_mode", "KERBEROS"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "sid", ""),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "total_resources_count", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "enabled_resources_count", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "description", ""),
						resource.TestCheckResourceAttr("data.scc_system_mapping.test", "sap_router", ""),
					),
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
					Config:      DataSourceSystemMappingWoRegionHost("test", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8", "testterraformvirtual", "900"),
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
					Config:      DataSourceSystemMappingWoSubaccount("test", "cf.eu12.hana.ondemand.com", "testterraformvirtual", "900"),
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
					Config:      DataSourceSystemMappingWoVirtualHost("test", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8", "900"),
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
					Config:      DataSourceSystemMappingWoVirtualPort("test", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8", "testterraformvirtual"),
					ExpectError: regexp.MustCompile(`The argument "virtual_port" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSystemMapping(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"	
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort)
}

func DataSourceSystemMappingWoRegionHost(datasourceName string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping" "%s" {
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"	
	}
	`, datasourceName, subaccount, virtualHost, virtualPort)
}

func DataSourceSystemMappingWoSubaccount(datasourceName string, regionHost string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping" "%s" {
	region_host= "%s"
	virtual_host= "%s"
	virtual_port= "%s"	
	}
	`, datasourceName, regionHost, virtualHost, virtualPort)
}

func DataSourceSystemMappingWoVirtualHost(datasourceName string, regionHost string, subaccount string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_port= "%s"	
	}
	`, datasourceName, regionHost, subaccount, virtualPort)
}

func DataSourceSystemMappingWoVirtualPort(datasourceName string, regionHost string, subaccount string, virtualHost string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost)
}
