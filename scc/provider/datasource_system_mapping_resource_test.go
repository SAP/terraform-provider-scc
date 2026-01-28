package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMappingResource(t *testing.T) {
	regionHost:= "cf.eu12.hana.ondemand.com"
	subaccount:= "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	virtualHost:= "testterraformvirtual"
	virtualPort:= "900"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping_resource")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSystemMappingResource("scc_smr", regionHost, subaccount, virtualHost, virtualPort, "/"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_system_mapping_resource.scc_smr", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "virtual_host", virtualHost),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "virtual_port", virtualPort),

						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "url_path", "/"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "enabled", "true"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "path_only", "true"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "websocket_upgrade_allowed", "false"),
						resource.TestMatchResourceAttr("data.scc_system_mapping_resource.scc_smr", "creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.scc_smr", "description", ""),
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
					Config:      DataSourceSystemMappingResourceWoRegionHost("scc_smr", subaccount, virtualHost, virtualPort, "/"),
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
					Config:      DataSourceSystemMappingResourceWoSubaccount("scc_smr", regionHost, virtualHost, virtualPort, "/"),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
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
					Config:      DataSourceSystemMappingResourceWoResourceID("scc_smr", regionHost, subaccount, virtualHost, virtualPort),
					ExpectError: regexp.MustCompile(`The argument "url_path" is required, but no definition was found.`),
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
					Config:      DataSourceSystemMappingResourceWoVirtualHost("scc_smr", regionHost, subaccount, virtualPort, "/"),
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
					Config:      DataSourceSystemMappingResourceWoVirtualPort("scc_smr", regionHost, subaccount, virtualHost, "/"),
					ExpectError: regexp.MustCompile(`The argument "virtual_port" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSystemMappingResource(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string, urlPath string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resource" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	url_path= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, urlPath)
}

func DataSourceSystemMappingResourceWoRegionHost(datasourceName string, subaccount string, virtualHost string, virtualPort string, urlPath string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resource" "%s" {
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	url_path= "%s"
	}
	`, datasourceName, subaccount, virtualHost, virtualPort, urlPath)
}

func DataSourceSystemMappingResourceWoSubaccount(datasourceName string, regionHost string, virtualHost string, virtualPort string, urlPath string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resource" "%s" {
	region_host= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	url_path= "%s"
	}
	`, datasourceName, regionHost, virtualHost, virtualPort, urlPath)
}

func DataSourceSystemMappingResourceWoVirtualHost(datasourceName string, regionHost string, subaccount string, virtualPort string, urlPath string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resource" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_port= "%s"
	url_path= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualPort, urlPath)
}

func DataSourceSystemMappingResourceWoVirtualPort(datasourceName string, regionHost string, subaccount string, virtualHost string, urlPath string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resource" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	url_path= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, urlPath)
}

func DataSourceSystemMappingResourceWoResourceID(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping_resource" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort)
}
