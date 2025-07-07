package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSubaccountK8SServiceChannel(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_k8s_service_channel")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountK8SServiceChannel("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "cp.ace9fb5.stage.kyma.ondemand.com:443", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f", 3000, 1, true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("scc_subaccount_k8s_service_channel.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "k8s_cluster", "cp.ace9fb5.stage.kyma.ondemand.com:443"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "k8s_service", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "port", "3000"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "connections", "1"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "type", "K8S"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "enabled", "true"),
						resource.TestCheckResourceAttrSet("scc_subaccount_k8s_service_channel.test", "id"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "state.connected", "true"),
						resource.TestMatchResourceAttr("scc_subaccount_k8s_service_channel.test", "state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "state.opened_connections", "1"),
					),
				},
				{
					ResourceName:      "scc_subaccount_k8s_service_channel.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: getImportStateForSubaccountK8SServiceChannel("scc_subaccount_k8s_service_channel.test"),
				},
				{
					ResourceName:  "scc_subaccount_k8s_service_channel.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.comd3bbbcd7-d5e0-483b-a524-6dee7205f8e81", // malformed ID
					ExpectError:   regexp.MustCompile(`(?s)Expected import identifier with format:.*id.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount_k8s_service_channel.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,d3bbbcd7-d5e0-483b-a524-6dee7205f8e8,1,extra",
					ExpectError:   regexp.MustCompile(`(?s)Expected import identifier with format:.*id.*Got:`),
				},
			},
		})

	})

	t.Run("update path - comment and connections update", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_k8s_service_channel_update")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + `
					resource "scc_subaccount_k8s_service_channel" "test" {
					  region_host = "cf.eu12.hana.ondemand.com"
					  subaccount = "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8"
					  k8s_cluster = "cp.ace9fb5.stage.kyma.ondemand.com:443"
					  k8s_service = "29d4e6f6-8e7f-4882-b434-21a52bb75e0f"
					  port = 3000
					  connections = 1
					  comment = "initial"
					  enabled = true
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "comment", "initial"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "connections", "1"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "enabled", "true"),
					),
				},
				{
					Config: providerConfig(user) + `
					resource "scc_subaccount_k8s_service_channel" "test" {
					  region_host = "cf.eu12.hana.ondemand.com"
					  subaccount = "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8"
					  k8s_cluster = "cp.ace9fb5.stage.kyma.ondemand.com:443"
					  k8s_service = "29d4e6f6-8e7f-4882-b434-21a52bb75e0f"
					  port = 3000
					  connections = 2
					  comment = "updated"
					  enabled = false
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "comment", "updated"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "connections", "2"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "enabled", "false"),
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
					Config:      ResourceSubaccountK8SServiceChannelWoRegionHost("test", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "cp.ace9fb5.stage.kyma.ondemand.com:443", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f", 3000, 1, true),
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
					Config:      ResourceSubaccountK8SServiceChannelWoSubaccount("test", "cf.eu12.hana.ondemand.com", "cp.ace9fb5.stage.kyma.ondemand.com:443", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f", 3000, 1, true),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - k8s cluster mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountK8SServiceChannelWoCluster("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f", 3000, 1, true),
					ExpectError: regexp.MustCompile(`The argument "k8s_cluster" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - k8s service mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountK8SServiceChannelWoService("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "cp.ace9fb5.stage.kyma.ondemand.com:443", 3000, 1, true),
					ExpectError: regexp.MustCompile(`The argument "k8s_service" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - port mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountK8SServiceChannelWoPort("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "cp.ace9fb5.stage.kyma.ondemand.com:443", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f", 1, true),
					ExpectError: regexp.MustCompile(`The argument "port" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - connections mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountK8SServiceChannelWoConnections("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "cp.ace9fb5.stage.kyma.ondemand.com:443", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f", 3000, true),
					ExpectError: regexp.MustCompile(`The argument "connections" is required, but no definition was found.`),
				},
			},
		})
	})

}

func ResourceSubaccountK8SServiceChannel(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, port int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster =  "%s"
	k8s_service =  "%s"
	port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, k8sService, port, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoRegionHost(datasourceName string, subaccount string, k8sCluster string, k8sService string, port int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	subaccount = "%s"
	k8s_cluster =  "%s"
	k8s_service =  "%s"
	port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, subaccount, k8sCluster, k8sService, port, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoSubaccount(datasourceName string, regionHost string, k8sCluster string, k8sService string, port int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	k8s_cluster =  "%s"
	k8s_service =  "%s"
	port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, k8sCluster, k8sService, port, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoCluster(datasourceName string, regionHost string, subaccount string, k8sService string, port int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_service =  "%s"
	port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sService, port, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoService(datasourceName string, regionHost string, subaccount string, k8sCluster string, port int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster =  "%s"
	port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, port, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoPort(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster =  "%s"
	k8s_service =  "%s"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, k8sService, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoConnections(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, port int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster =  "%s"
	k8s_service =  "%s"
	port = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, k8sService, port, enabled)
}

func getImportStateForSubaccountK8SServiceChannel(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
			rs.Primary.Attributes["id"],
		), nil
	}
}
