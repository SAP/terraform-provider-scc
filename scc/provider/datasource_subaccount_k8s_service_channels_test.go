package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccountK8SServiceChannels(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount:= "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccount_k8s_service_channels")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSubaccountK8SServiceChannels("scc_k8s_scs", regionHost, subaccount),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.#", "1"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.k8s_cluster_host"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.k8s_service_id"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.local_port", "3000"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.connections", "1"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.type", "K8S"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.enabled", "false"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.id"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.state.connected", "false"),
						resource.TestMatchResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channels.scc_k8s_scs", "subaccount_k8s_service_channels.0.state.opened_connections", "0"),
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
					Config:      DataSourceSubaccountK8SServiceChannelsWoRegionHost("scc_sc", subaccount),
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
					Config:      DataSourceSubaccountK8SServiceChannelsWoSubaccount("scc_sc", regionHost),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSubaccountK8SServiceChannels(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_k8s_service_channels" "%s" {
	region_host = "%s"
	subaccount = "%s"
	}
	`, datasourceName, regionHost, subaccountID)
}

func DataSourceSubaccountK8SServiceChannelsWoSubaccount(datasourceName string, regionHost string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_k8s_service_channels" "%s" {
	region_host = "%s"
	}
	`, datasourceName, regionHost)
}

func DataSourceSubaccountK8SServiceChannelsWoRegionHost(datasourceName string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_k8s_service_channels" "%s" {
	subaccount = "%s"
	}
	`, datasourceName, subaccountID)
}
