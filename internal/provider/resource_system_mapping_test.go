package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				{
					Config: providerConfig("", user) + ResourceSystemMapping("test", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7", "testtfvirtualtesting", "90", "testtfinternaltesting", "90", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "credentials.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_system_mapping.test", "credentials.subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.virtual_host", "testtfvirtualtesting"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.virtual_port", "90"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.local_host", "testtfinternaltesting"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.local_port", "90"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.protocol", "HTTP"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.backend_type", "abapSys"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.authentication_mode", "KERBEROS"),
					),
				},
			},
		})

	})

}

func ResourceSystemMapping(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "cloudconnector_system_mapping" "%s" {
    credentials= {
        region_host= "%s"
        subaccount= "%s"
    }
    system_mapping= {
      virtual_host= "%s"
      virtual_port= "%s"
      local_host= "%s"
      local_port= "%s"
      protocol= "%s"
      backend_type= "%s"
      host_in_header= "%s"
      authentication_mode= "%s"
    }
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, localPort, protocol, backendType, hostInHeader, authenticationMode)
}
