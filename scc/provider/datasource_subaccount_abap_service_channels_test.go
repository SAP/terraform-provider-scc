package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccountABAPServiceChannels(t *testing.T) {
	regionHost:= "cf.eu12.hana.ondemand.com"
	subaccount:= "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccount_abap_service_channels")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSubaccountABAPServiceChannels("scc_abap_scs", regionHost, subaccount),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.#", "1"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.abap_cloud_tenant_host"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.instance_number", "50"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.port", "3350"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.connections", "1"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.type", "ABAPCloud"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.enabled", "false"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.id"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.state.connected", "false"),
						resource.TestMatchResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount_abap_service_channels.0.state.opened_connections", "0"),
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
					Config:      DataSourceSubaccountABAPServiceChannelsWoRegionHost("scc_sc", subaccount),
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
					Config:      DataSourceSubaccountABAPServiceChannelsWoSubaccount("scc_sc", regionHost),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSubaccountABAPServiceChannels(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channels" "%s" {
	region_host = "%s"
	subaccount = "%s"
	}
	`, datasourceName, regionHost, subaccountID)
}

func DataSourceSubaccountABAPServiceChannelsWoSubaccount(datasourceName string, regionHost string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channels" "%s" {
	region_host = "%s"
	}
	`, datasourceName, regionHost)
}

func DataSourceSubaccountABAPServiceChannelsWoRegionHost(datasourceName string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channels" "%s" {
	subaccount = "%s"
	}
	`, datasourceName, subaccountID)
}
