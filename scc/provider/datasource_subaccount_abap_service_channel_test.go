package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccountABAPServiceChannel(t *testing.T) {

	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "304492be-5f0f-4bb0-8f59-c982107bc878"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccount_abap_service_channel")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSubaccountABAPServiceChannel("scc_sc", regionHost, subaccount, 1),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_abap_service_channel.scc_sc", "abap_cloud_tenant_host"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "instance_number", "50"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "port", "3350"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "connections", "1"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "type", "ABAPCloud"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "enabled", "false"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_abap_service_channel.scc_sc", "id"),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "state.connected", "false"),
						resource.TestMatchResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channel.scc_sc", "state.opened_connections", "0"),
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
					Config:      DataSourceSubaccountABAPServiceChannelWoRegionHost("scc_sc", subaccount, 49),
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
					Config:      DataSourceSubaccountABAPServiceChannelWoSubaccount("scc_sc", regionHost, 49),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - channel id mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountABAPServiceChannelWoID("scc_sc", regionHost, subaccount),
					ExpectError: regexp.MustCompile(`The argument "id" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSubaccountABAPServiceChannel(datasourceName string, regionHost string, subaccountID string, id int64) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	id = "%d"
	}
	`, datasourceName, regionHost, subaccountID, id)
}

func DataSourceSubaccountABAPServiceChannelWoSubaccount(datasourceName string, regionHost string, id int64) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	id = "%d"
	}
	`, datasourceName, regionHost, id)
}

func DataSourceSubaccountABAPServiceChannelWoRegionHost(datasourceName string, subaccountID string, id int64) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channel" "%s" {
	subaccount = "%s"
	id = "%d"
	}
	`, datasourceName, subaccountID, id)
}

func DataSourceSubaccountABAPServiceChannelWoID(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	}
	`, datasourceName, regionHost, subaccountID)
}
