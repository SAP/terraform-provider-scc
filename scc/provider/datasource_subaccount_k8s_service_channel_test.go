package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccountK8SServiceChannel(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount:= "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccount_k8s_service_channel")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSubaccountK8SServiceChannel("scc_k8s_sc", regionHost, subaccount, 4),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "k8s_cluster_host"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "k8s_service_id"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "local_port", "3000"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "connections", "1"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "type", "K8S"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "enabled", "false"),
						resource.TestCheckResourceAttrSet("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "id"),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "state.connected", "false"),
						resource.TestMatchResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("data.scc_subaccount_k8s_service_channel.scc_k8s_sc", "state.opened_connections", "0"),
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
					Config:      DataSourceSubaccountK8SServiceChannelWoRegionHost("scc_k8s_sc", subaccount, 4),
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
					Config:      DataSourceSubaccountK8SServiceChannelWoSubaccount("scc_k8s_sc", regionHost, 4),
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
					Config:      DataSourceSubaccountK8SServiceChannelWoID("scc_k8s_sc", regionHost, subaccount),
					ExpectError: regexp.MustCompile(`The argument "id" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSubaccountK8SServiceChannel(datasourceName string, regionHost string, subaccountID string, id int64) string {
	return fmt.Sprintf(`
	data "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	id = "%d"
	}
	`, datasourceName, regionHost, subaccountID, id)
}

func DataSourceSubaccountK8SServiceChannelWoSubaccount(datasourceName string, regionHost string, id int64) string {
	return fmt.Sprintf(`
	data "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	id = "%d"
	}
	`, datasourceName, regionHost, id)
}

func DataSourceSubaccountK8SServiceChannelWoRegionHost(datasourceName string, subaccountID string, id int64) string {
	return fmt.Sprintf(`
	data "scc_subaccount_k8s_service_channel" "%s" {
	subaccount = "%s"
	id = "%d"
	}
	`, datasourceName, subaccountID, id)
}

func DataSourceSubaccountK8SServiceChannelWoID(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	}
	`, datasourceName, regionHost, subaccountID)
}
