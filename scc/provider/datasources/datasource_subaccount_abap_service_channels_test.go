package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccountABAPServiceChannels(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_subaccount_abap_service_channels")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceSubaccountABAPServiceChannels("scc_abap_scs", regionHost, subaccount, false),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "subaccount", tfutils.RegexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_subaccount_abap_service_channels.scc_abap_scs", "snc_encrypted", "false"),

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
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountABAPServiceChannelsWoRegionHost("scc_sc", subaccount, false),
					ExpectError: regexp.MustCompile(`The argument "region_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - subaccount id mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountABAPServiceChannelsWoSubaccount("scc_sc", regionHost, false),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - SNC Encrypted mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountABAPServiceChannelsWoSNCEncrypted("scc_sc", regionHost, subaccount),
					ExpectError: regexp.MustCompile(`The argument "snc_encrypted" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSubaccountABAPServiceChannels(datasourceName string, regionHost string, subaccountID string, sncEncrypted bool) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channels" "%s" {
	region_host = "%s"
	subaccount = "%s"
	snc_encrypted = "%t"
	}
	`, datasourceName, regionHost, subaccountID, sncEncrypted)
}

func DataSourceSubaccountABAPServiceChannelsWoSubaccount(datasourceName string, regionHost string, sncEncrypted bool) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channels" "%s" {
	region_host = "%s"
	snc_encrypted = "%t"
	}
	`, datasourceName, regionHost, sncEncrypted)
}

func DataSourceSubaccountABAPServiceChannelsWoRegionHost(datasourceName string, subaccountID string, sncEncrypted bool) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channels" "%s" {
	subaccount = "%s"
	snc_encrypted = "%t"
	}
	`, datasourceName, subaccountID, sncEncrypted)
}

func DataSourceSubaccountABAPServiceChannelsWoSNCEncrypted(datasourceName string, regionHost string, subaccount string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_abap_service_channels" "%s" {
	region_host = "%s"
	subaccount = "%s"
	}
	`, datasourceName, regionHost, subaccount)
}
