package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMappingResources(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	virtualHost := "testterraformvirtual"
	virtualPort := "900"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping_resources")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSystemMappingResources("scc_smrs", regionHost, subaccount, virtualHost, virtualPort),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_system_mapping_resources.scc_smrs", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "virtual_host", virtualHost),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "virtual_port", virtualPort),

						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "system_mapping_resources.#", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "system_mapping_resources.0.url_path", "/"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "system_mapping_resources.0.enabled", "true"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "system_mapping_resources.0.path_only", "true"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "system_mapping_resources.0.websocket_upgrade_allowed", "false"),
						resource.TestMatchResourceAttr("data.scc_system_mapping_resources.scc_smrs", "system_mapping_resources.0.creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resources.scc_smrs", "system_mapping_resources.0.description", ""),
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
					Config:      DataSourceSystemMappingResourcesWoRegionHost("scc_smrs", subaccount, virtualHost, virtualPort),
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
					Config:      DataSourceSystemMappingResourcesWoSubaccount("scc_smrs", regionHost, virtualHost, virtualPort),
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
					Config:      DataSourceSystemMappingResourcesWoVirtualHost("scc_smrs", regionHost, subaccount, virtualPort),
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
					Config:      DataSourceSystemMappingResourcesWoVirtualPort("scc_smrs", regionHost, subaccount, virtualHost),
					ExpectError: regexp.MustCompile(`The argument "virtual_port" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSystemMappingResources(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resources" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort)
}

func DataSourceSystemMappingResourcesWoRegionHost(datasourceName string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resources" "%s" {
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	}
	`, datasourceName, subaccount, virtualHost, virtualPort)
}

func DataSourceSystemMappingResourcesWoSubaccount(datasourceName string, regionHost string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resources" "%s" {
	region_host= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	}
	`, datasourceName, regionHost, virtualHost, virtualPort)
}

func DataSourceSystemMappingResourcesWoVirtualHost(datasourceName string, regionHost string, subaccount string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resources" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_port= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualPort)
}

func DataSourceSystemMappingResourcesWoVirtualPort(datasourceName string, regionHost string, subaccount string, virtualHost string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resources" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost)
}
