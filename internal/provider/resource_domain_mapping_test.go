package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceDomainMapping(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_domain_mapping")
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
					Config: providerConfig(user) + ResourceDomainMapping("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "testtfvirtualdomain", "testtfinternaldomain"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_domain_mapping.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_domain_mapping.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_domain_mapping.test", "virtual_domain", "testtfvirtualdomain"),
						resource.TestCheckResourceAttr("cloudconnector_domain_mapping.test", "internal_domain", "testtfinternaldomain"),
					),
				},
				{
					ResourceName:                         "cloudconnector_domain_mapping.test",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubaccountEntitlement("cloudconnector_domain_mapping.test"),
					ImportStateVerifyIdentifierAttribute: "internal_domain",
				},
			},
		})

	})

}

func ResourceDomainMapping(datasourceName string, regionHost string, subaccount string, virtualDomain string, internalDomain string) string {
	return fmt.Sprintf(`
	resource "cloudconnector_domain_mapping" "%s" {
    region_host = "%s"
    subaccount = "%s"
    virtual_domain = "%s"
    internal_domain = "%s"
	}
	`, datasourceName, regionHost, subaccount, virtualDomain, internalDomain)
}

func getImportStateForSubaccountEntitlement(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
			rs.Primary.Attributes["internal_domain"],
		), nil
	}
}
