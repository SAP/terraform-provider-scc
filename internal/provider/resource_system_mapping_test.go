package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					Config: providerConfig("", user) + ResourceSystemMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtual", "900", "testtfinternal", "900", "HTTP", "abapSys", "VIRTUAL", "KERBEROS"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "credentials.region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_system_mapping.test", "credentials.subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.virtual_host", "testtfvirtual"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.virtual_port", "900"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.local_host", "testtfinternal"),
						resource.TestCheckResourceAttr("cloudconnector_system_mapping.test", "system_mapping.local_port", "900"),
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
