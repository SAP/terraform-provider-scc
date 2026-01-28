package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMappings(t *testing.T) {
	regionHost:= "cf.eu12.hana.ondemand.com"
	subaccount:= "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	virtualHost:= "testterraformvirtual"
	virtualPort:= "900"
	internalHost:= "testterraforminternal"
	internalPort:= "900"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mappings")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSystemMappings("scc_sms", regionHost, subaccount),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_system_mappings.scc_sms", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.#", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.virtual_host", virtualHost),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.virtual_port", virtualPort),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.internal_host", internalHost),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.internal_port", internalPort),
						resource.TestMatchResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.protocol", "HTTP"),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.backend_type", "abapSys"),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.authentication_mode", "KERBEROS"),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.sid", ""),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.total_resources_count", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.enabled_resources_count", "1"),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.description", ""),
						resource.TestCheckResourceAttr("data.scc_system_mappings.scc_sms", "system_mappings.0.sap_router", ""),
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
					Config:      DataSourceSystemMappingsWoRegionHost("scc_sms", subaccount),
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
					Config:      DataSourceSystemMappingsWoSubaccount("scc_sms", regionHost),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSystemMappings(datasourceName string, regionHost string, subaccount string) string {
	return fmt.Sprintf(`
	data "scc_system_mappings" "%s" {
	region_host= "%s"
	subaccount= "%s"
	}
	`, datasourceName, regionHost, subaccount)
}

func DataSourceSystemMappingsWoRegionHost(datasourceName string, subaccount string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping" "%s" {
	subaccount= "%s"
	}
	`, datasourceName, subaccount)
}

func DataSourceSystemMappingsWoSubaccount(datasourceName string, regionHost string) string {
	return fmt.Sprintf(`
	data "scc_system_mapping" "%s" {
	region_host= "%s"
	}
	`, datasourceName, regionHost)
}
