package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMapping(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	virtualHost := "testterraformvirtual"
	virtualPort := "900"
	internalHost := "testterraforminternal"
	internalPort := "900"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSystemMapping("scc_sm", regionHost, subaccount, virtualHost, virtualPort),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_system_mapping.scc_sm", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "virtual_host", virtualHost),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "virtual_port", virtualPort),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "internal_host", internalHost),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "internal_port", internalPort),
						resource.TestMatchResourceAttr("data.scc_system_mapping.scc_sm", "creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "protocol", "HTTP"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "backend_type", "abapSys"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "authentication_mode", "KERBEROS"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "sid", ""),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "total_resources_count", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "enabled_resources_count", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "description", ""),
						resource.TestCheckResourceAttr("data.scc_system_mapping.scc_sm", "sap_router", ""),
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
					Config:      DataSourceSystemMappingWoRegionHost("scc_sm", subaccount, virtualHost, virtualPort),
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
					Config:      DataSourceSystemMappingWoSubaccount("scc_sm", regionHost, virtualHost, virtualPort),
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
					Config:      DataSourceSystemMappingWoVirtualHost("scc_sm", regionHost, subaccount, virtualPort),
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
					Config:      DataSourceSystemMappingWoVirtualPort("scc_sm", regionHost, subaccount, virtualHost),
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
