package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceDomainMapping(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	virtualDomain := "testterraformvirtualdomain"
	internalDomain := "testterraforminternaldomain"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_domain_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceDomainMapping("scc_dm", regionHost, subaccount, internalDomain),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_domain_mapping.scc_dm", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_domain_mapping.scc_dm", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_domain_mapping.scc_dm", "virtual_domain", virtualDomain),
						resource.TestCheckResourceAttr("data.scc_domain_mapping.scc_dm", "internal_domain", internalDomain),
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
					Config:      DataSourceDomainMappingWoRegionHost("scc_dm", subaccount, internalDomain),
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
					Config:      DataSourceDomainMappingWoSubaccount("scc_dm", regionHost, internalDomain),
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
					Config:      DataSourceDomainMappingWoInternalDomain("scc_dm", regionHost, subaccount),
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
