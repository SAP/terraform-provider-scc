package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceDomainMapping(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_domain_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + ResourceDomainMapping("test", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7", "testtfvirtualdomain", "testtfinternaldomain"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_domain_mapping.test", "credentials.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_domain_mapping.test", "credentials.subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_domain_mapping.test", "domain_mapping.virtual_domain", "testtfvirtualdomain"),
						resource.TestCheckResourceAttr("cloudconnector_domain_mapping.test", "domain_mapping.internal_domain", "testtfinternaldomain"),
					),
				},
			},
		})

	})

}

func ResourceDomainMapping(datasourceName string, regionHost string, subaccount string, virtualDomain string, internalDomain string) string {
	return fmt.Sprintf(`
	resource "cloudconnector_domain_mapping" "%s" {
	credentials= {
		region_host= "%s"
		subaccount= "%s"
	}
	domain_mapping={
		virtual_domain="%s"
		internal_domain="%s"
	}
	}
	`, datasourceName, regionHost, subaccount, virtualDomain, internalDomain)
}
