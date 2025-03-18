package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceDomainMappings(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_domain_mappings")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + DataSourceDomainMappings("mappings", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_domain_mappings.mappings", "credentials.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_domain_mappings.mappings", "credentials.subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.cloudconnector_domain_mappings.mappings", "domain_mappings.#", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_domain_mappings.mappings", "domain_mappings.0.virtual_domain", "testterraformvirtualdomain"),
						resource.TestCheckResourceAttr("data.cloudconnector_domain_mappings.mappings", "domain_mappings.0.internal_domain", "testterraforminternaldomain"),
					),
				},
			},
		})

	})

}

func DataSourceDomainMappings(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "cloudconnector_domain_mappings" "%s" {
	credentials= {
	region_host= "%s"
    subaccount= "%s"
	}
	}
	`, datasourceName, regionHost, subaccountID)
}
