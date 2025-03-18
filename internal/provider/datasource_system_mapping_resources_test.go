package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMappingResources(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping_resources")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + DataSourceSystemMappingResources("test", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7", "testterraformvirtual", "900"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.virtual_host", "testterraformvirtual"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.virtual_port", "900"),

						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.#", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.id", "/google.com"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.enabled", "true"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.exact_match_only", "true"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.websocket_upgrade_allowed", "false"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.description", ""),
					),
				},
			},
		})

	})

}

func DataSourceSystemMappingResources(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "cloudconnector_system_mapping_resources" "%s" {
    credentials= {
        region_host= "%s"
        subaccount= "%s"
        virtual_host= "%s"
        virtual_port= "%s"
    }
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort)
}
