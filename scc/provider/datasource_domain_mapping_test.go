package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceDomainMapping(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_domain_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceDomainMapping("mapping", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8", "testterraforminternaldomain"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_domain_mapping.mapping", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.scc_domain_mapping.mapping", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_domain_mapping.mapping", "virtual_domain", "testterraformvirtualdomain"),
						resource.TestCheckResourceAttr("data.scc_domain_mapping.mapping", "internal_domain", "testterraforminternaldomain"),
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
					Config:      DataSourceDomainMappingWoRegionHost("mapping", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8", "testterraforminternaldomain"),
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
					Config:      DataSourceDomainMappingWoSubaccount("mapping", "cf.eu12.hana.ondemand.com", "testterraforminternaldomain"),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - internal domain mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceDomainMappingWoInternalDomain("mapping", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8"),
					ExpectError: regexp.MustCompile(`The argument "internal_domain" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceDomainMapping(datasourceName string, regionHost string, subaccountID string, internalDomain string) string {
	return fmt.Sprintf(`
	data "scc_domain_mapping" "%s" {
	region_host= "%s"
    subaccount= "%s"
	internal_domain= "%s"
	}
	`, datasourceName, regionHost, subaccountID, internalDomain)
}

func DataSourceDomainMappingWoRegionHost(datasourceName string, subaccountID string, internalDomain string) string {
	return fmt.Sprintf(`
	data "scc_domain_mapping" "%s" {
    subaccount= "%s"
	internal_domain= "%s"
	}
	`, datasourceName, subaccountID, internalDomain)
}

func DataSourceDomainMappingWoSubaccount(datasourceName string, regionHost string, internalDomain string) string {
	return fmt.Sprintf(`
	data "scc_domain_mapping" "%s" {
	region_host= "%s"
	internal_domain= "%s"
	}
	`, datasourceName, regionHost, internalDomain)
}

func DataSourceDomainMappingWoInternalDomain(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_domain_mapping" "%s" {
	region_host= "%s"
    subaccount= "%s"
	}
	`, datasourceName, regionHost, subaccountID)
}
