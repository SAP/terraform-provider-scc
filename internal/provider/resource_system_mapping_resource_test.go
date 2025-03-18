package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					Config: providerConfig("", user) + ResourceSystemMappingResource("test", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7", "testtfvirtual", "900", "/google.com", "create resource", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "credentials.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_system_mapping_resource.test", "credentials.subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "credentials.virtual_host", "testtfvirtual"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "credentials.virtual_port", "900"),

						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "system_mapping_resource.id", "/google.com"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "system_mapping_resource.description", "create resource"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "system_mapping_resource.enabled", "true"),
					),
				},
			},
		})

	})

}

func ResourceSystemMappingResource(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	id string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "cloudconnector_system_mapping_resource" "%s" {
    credentials= {
        region_host= "%s"
        subaccount= "%s"
        virtual_host= "%s"
        virtual_port= "%s"
    }
    system_mapping_resource= {
        id= "%s"
        description= "%s"
        enabled=%t
    }
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, id, description, enabled)
}
