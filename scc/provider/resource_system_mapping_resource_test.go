package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSystemMappingResource(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_system_mapping_resource")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSystemMappingResource("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualtesting", "90", "/", "create resource", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_mapping_resource.test", "description", "create resource"),
						resource.TestCheckResourceAttr("scc_system_mapping_resource.test", "enabled", "true"),
					),
				},
				{
					// 🚀 This is the update step
					Config: providerConfig(user) + ResourceSystemMappingResource("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualtesting", "90", "/", "updated resource", false),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_mapping_resource.test", "description", "updated resource"),
						resource.TestCheckResourceAttr("scc_system_mapping_resource.test", "enabled", "false"),
					),
				},
				{
					ResourceName:      "scc_system_mapping_resource.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: getImportStateForSystemMappingResource("scc_system_mapping_resource.test"),
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
					Config:      ResourceSystemMappingResourceWoRegionHost("test", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualtesting", "90", "/", "create resource", true),
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
					Config:      ResourceSystemMappingResourceWoSubaccount("test", "cf.eu12.hana.ondemand.com", "testtfvirtualtesting", "90", "/", "create resource", true),
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
					Config:      ResourceSystemMappingResourceWoVirtualHost("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "90", "/", "create resource", true),
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
					Config:      ResourceSystemMappingResourceWoVirtualPort("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualtesting", "/", "create resource", true),
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
					Config:      ResourceSystemMappingResourceWoID("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualtesting", "90", "create resource", true),
					ExpectError: regexp.MustCompile(`The argument "id" is required, but no definition was found.`),
				},
			},
		})
	})

}

func ResourceSystemMappingResource(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	id string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	subaccount = "%s"
	virtual_host = "%s"
	virtual_port = "%s"
	id = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, id, description, enabled)
}

func ResourceSystemMappingResourceWoRegionHost(datasourceName string, subaccount string, virtualHost string, virtualPort string,
	id string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	subaccount = "%s"
	virtual_host = "%s"
	virtual_port = "%s"
	id = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, subaccount, virtualHost, virtualPort, id, description, enabled)
}

func ResourceSystemMappingResourceWoSubaccount(datasourceName string, regionHost string, virtualHost string, virtualPort string,
	id string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	virtual_host = "%s"
	virtual_port = "%s"
	id = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, virtualHost, virtualPort, id, description, enabled)
}

func ResourceSystemMappingResourceWoVirtualHost(datasourceName string, regionHost string, subaccount string, virtualPort string,
	id string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	subaccount = "%s"
	virtual_port = "%s"
	id = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, subaccount, virtualPort, id, description, enabled)
}

func ResourceSystemMappingResourceWoVirtualPort(datasourceName string, regionHost string, subaccount string, virtualHost string,
	id string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_system_mapping_resource" "%s" {
	region_host = "%s"
	subaccount = "%s"
	virtual_host = "%s"
	id = "%s"
	description = "%s"
	enabled = "%t"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, id, description, enabled)
}

func ResourceSystemMappingResourceWoID(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
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
			rs.Primary.Attributes["id"],
		), nil
	}
}
