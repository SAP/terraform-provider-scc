package provider

import (
	"fmt"
	"regexp"
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
					Config: providerConfig(user) + DataSourceDomainMappings("mappings", "cf.eu12.hana.ondemand.com", "304492be-5f0f-4bb0-8f59-c982107bc878"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_domain_mappings.mappings", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.scc_domain_mappings.mappings", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.scc_domain_mappings.mappings", "domain_mappings.#", "1"),
						resource.TestCheckResourceAttr("data.scc_domain_mappings.mappings", "domain_mappings.0.virtual_domain", "testterraformvirtualdomain"),
						resource.TestCheckResourceAttr("data.scc_domain_mappings.mappings", "domain_mappings.0.internal_domain", "testterraforminternaldomain"),
					),
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
					Config:      DataSourceDomainMappingsWoRegionHost("mappings", "304492be-5f0f-4bb0-8f59-c982107bc878"),
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
					Config:      DataSourceDomainMappingsWoSubaccount("mappings", "cf.eu12.hana.ondemand.com"),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceDomainMappings(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_domain_mappings" "%s" {
	region_host= "%s"
    subaccount= "%s"
	}
	`, datasourceName, regionHost, subaccountID)
}

func DataSourceDomainMappingsWoRegionHost(datasourceName string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_domain_mappings" "%s" {
    subaccount= "%s"
	}
	`, datasourceName, subaccountID)
}

func DataSourceDomainMappingsWoSubaccount(datasourceName string, regionHost string) string {
	return fmt.Sprintf(`
	data "scc_domain_mappings" "%s" {
	region_host= "%s"
	}
	`, datasourceName, regionHost)
}
