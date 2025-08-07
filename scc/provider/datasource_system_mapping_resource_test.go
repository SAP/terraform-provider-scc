package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMappingResource(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping_resource")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSystemMappingResource("test", "cf.eu12.hana.ondemand.com", "304492be-5f0f-4bb0-8f59-c982107bc878", "testterraformvirtual", "900", "/"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.scc_system_mapping_resource.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "virtual_host", "testterraformvirtual"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "virtual_port", "900"),

						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "url_path", "/"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "enabled", "true"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "path_only", "true"),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "websocket_upgrade_allowed", "false"),
						resource.TestMatchResourceAttr("data.scc_system_mapping_resource.test", "creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_system_mapping_resource.test", "description", ""),
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
					Config:      DataSourceSystemMappingResourceWoRegionHost("test", "304492be-5f0f-4bb0-8f59-c982107bc878", "testterraformvirtual", "900", "/"),
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
					Config:      DataSourceSystemMappingResourceWoSubaccount("test", "cf.eu12.hana.ondemand.com", "testterraformvirtual", "900", "/"),
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
					Config:      DataSourceSystemMappingResourceWoResourceID("test", "cf.eu12.hana.ondemand.com", "304492be-5f0f-4bb0-8f59-c982107bc878", "testterraformvirtual", "900"),
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
					Config:      DataSourceSystemMappingResourceWoVirtualHost("test", "cf.eu12.hana.ondemand.com", "304492be-5f0f-4bb0-8f59-c982107bc878", "900", "/"),
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
					Config:      DataSourceSystemMappingResourceWoVirtualPort("test", "cf.eu12.hana.ondemand.com", "304492be-5f0f-4bb0-8f59-c982107bc878", "testterraformvirtual", "/"),
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
