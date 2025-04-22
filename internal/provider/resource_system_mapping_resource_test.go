package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSystemMappingResource(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_system_mapping_resource")
		rec.SetRealTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSystemMappingResource("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualtesting", "90", "/google.com", "create resource", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_system_mapping_resource.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "virtual_host", "testtfvirtualtesting"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "virtual_port", "90"),

						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "id", "/google.com"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "description", "create resource"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping_resource.test", "enabled", "true"),
					),
				},
				{
					ResourceName:      "cloudconnector_system_mapping_resource.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: getImportStateForSystemMappingResource("cloudconnector_system_mapping_resource.test"),
				},
			},
		})

	})

}

func ResourceSystemMappingResource(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	id string, description string, enabled bool) string {
	return fmt.Sprintf(`
	resource "cloudconnector_system_mapping_resource" "%s" {
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
