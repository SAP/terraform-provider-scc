package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceDomainMappings(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	virtualDomain := "testterraformvirtualdomain"
	internalDomain := "testterraforminternaldomain"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_domain_mappings")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceDomainMappings("scc_dms", regionHost, subaccount),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_domain_mappings.scc_dms", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_domain_mappings.scc_dms", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.scc_domain_mappings.scc_dms", "domain_mappings.#", "1"),
						resource.TestCheckResourceAttr("data.scc_domain_mappings.scc_dms", "domain_mappings.0.virtual_domain", virtualDomain),
						resource.TestCheckResourceAttr("data.scc_domain_mappings.scc_dms", "domain_mappings.0.internal_domain", internalDomain),
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
					Config:      DataSourceDomainMappingsWoRegionHost("scc_dms", subaccount),
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
					Config:      DataSourceDomainMappingsWoSubaccount("scc_dms", regionHost),
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
