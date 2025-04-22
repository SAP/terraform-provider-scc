package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSystemMapping(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_system_mapping")
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
					Config: providerConfig(user) + ResourceSystemMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_system_mapping.test", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "virtual_host", "testtfvirtual"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "virtual_port", "900"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "local_host", "testtfinternal"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "local_port", "900"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "protocol", "HTTP"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "backend_type", "abapSys"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "authentication_mode", "KERBEROS"),
					),
				},
				{
					ResourceName:                         "cloudconnector_system_mapping.test",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSystemMapping("cloudconnector_system_mapping.test"),
					ImportStateVerifyIdentifierAttribute: "virtual_host",
				},
			},
		})

	})

}

func ResourceSystemMapping(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string,
	localHost string, localPort string, protocol string, backendType string, hostInHeader string, authenticationMode string) string {
	return fmt.Sprintf(`
	resource "cloudconnector_system_mapping" "%s" {
	region_host= "%s"
	subaccount= "%s"
	virtual_host= "%s"
	virtual_port= "%s"
	local_host= "%s"
	local_port= "%s"
	protocol= "%s"
	backend_type= "%s"
	host_in_header= "%s"
	authentication_mode= "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort, localHost, localPort, protocol, backendType, hostInHeader, authenticationMode)
}

func getImportStateForSystemMapping(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
			rs.Primary.Attributes["virtual_host"],
			rs.Primary.Attributes["virtual_port"],
		), nil
	}
}
