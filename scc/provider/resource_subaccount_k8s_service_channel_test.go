package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSubaccountK8SServiceChannel(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_k8s_service_channel")
		if len(user.K8SCluster) == 0 || len(user.K8SService) == 0 {
			t.Fatalf("Missing TF_VAR_k8s_cluster_host or TF_VAR_k8s_service_id for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountK8SServiceChannel("test", regionHost, subaccount, user.K8SCluster, user.K8SService, 3000, 1, true, "Created"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "region_host", regionHost),
						resource.TestMatchResourceAttr("scc_subaccount_k8s_service_channel.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "k8s_cluster_host", user.K8SCluster),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "k8s_service_id", user.K8SService),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "local_port", "3000"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "connections", "1"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "type", "K8S"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "enabled", "true"),
						resource.TestCheckResourceAttrSet("scc_subaccount_k8s_service_channel.test", "id"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "state.connected", "true"),
						resource.TestMatchResourceAttr("scc_subaccount_k8s_service_channel.test", "state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "state.opened_connections", "1"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_k8s_service_channel.test",
							map[string]knownvalue.Check{
								"id":          knownvalue.NotNull(),
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					ResourceName:    "scc_subaccount_k8s_service_channel.test",
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
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
					ImportStateId: regionHost + subaccount + "1", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*id.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount_k8s_service_channel.test",
					ImportState:   true,
					ImportStateId: regionHost + "," + subaccount + ",1, extra",
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*id.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount_k8s_service_channel.test",
					ImportState:   true,
					ImportStateId: regionHost + "," + subaccount + ",not-an-int",
					ExpectError:   regexp.MustCompile(`(?is)The 'id' part must be an integer.*Got:.*not-an-int`),
				},
			},
		})

	})

	t.Run("update path - description and connections update", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_k8s_service_channel_update")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountK8SServiceChannel("test", regionHost, subaccount, user.K8SCluster, user.K8SService, 3000, 1, true, "Created"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "description", "Created"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "connections", "1"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "enabled", "true"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_k8s_service_channel.test",
							map[string]knownvalue.Check{
								"id":          knownvalue.NotNull(),
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSubaccountK8SServiceChannel("test", "cf.us10.hana.ondemand.com", subaccount, user.K8SCluster, user.K8SService, 3000, 1, true, "Updated"),
					ExpectError: regexp.MustCompile(`(?is)error updating the cloud connector subaccount K8S service channel.*mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountK8SServiceChannel("test", regionHost, subaccount, user.K8SCluster, user.K8SService, 3000, 2, false, "Updated"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "description", "Updated"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "connections", "2"),
						resource.TestCheckResourceAttr("scc_subaccount_k8s_service_channel.test", "enabled", "false"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_k8s_service_channel.test",
							map[string]knownvalue.Check{
								"id":          knownvalue.NotNull(),
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
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
					Config:      ResourceSubaccountK8SServiceChannelWoRegionHost("test", regionHost, "testclusterhost", "testserviceid", 3000, 1, true),
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
					Config:      ResourceSubaccountK8SServiceChannelWoSubaccount("test", regionHost, "testclusterhost", "testserviceid", 3000, 1, true),
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
					Config:      ResourceSubaccountK8SServiceChannelWoCluster("test", regionHost, subaccount, "testserviceid", 3000, 1, true),
					ExpectError: regexp.MustCompile(`The argument "k8s_cluster_host" is required, but no definition was found.`),
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
					Config:      ResourceSubaccountK8SServiceChannelWoService("test", regionHost, subaccount, "testclusterhost", 3000, 1, true),
					ExpectError: regexp.MustCompile(`The argument "k8s_service_id" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - local port mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountK8SServiceChannelWoPort("test", regionHost, subaccount, "testclusterhost", "testserviceid", 1, true),
					ExpectError: regexp.MustCompile(`The argument "local_port" is required, but no definition was found.`),
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
					Config:      ResourceSubaccountK8SServiceChannelWoConnections("test", regionHost, subaccount, "testclusterhost", "testserviceid", 3000, true),
					ExpectError: regexp.MustCompile(`The argument "connections" is required, but no definition was found.`),
				},
			},
		})
	})

}

func ResourceSubaccountK8SServiceChannel(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, localPort int64, connections int64, enabled bool, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster_host =  "%s"
	k8s_service_id =  "%s"
	local_port = "%d"
	connections = "%d"
	enabled= "%t"
	description = "%s"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, k8sService, localPort, connections, enabled, description)
}

func ResourceSubaccountK8SServiceChannelWoRegionHost(datasourceName string, subaccount string, k8sCluster string, k8sService string, localPort int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	subaccount = "%s"
	k8s_cluster_host =  "%s"
	k8s_service_id =  "%s"
	local_port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, subaccount, k8sCluster, k8sService, localPort, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoSubaccount(datasourceName string, regionHost string, k8sCluster string, k8sService string, localPort int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	k8s_cluster_host =  "%s"
	k8s_service_id =  "%s"
	local_port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, k8sCluster, k8sService, localPort, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoCluster(datasourceName string, regionHost string, subaccount string, k8sService string, localPort int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_service_id =  "%s"
	local_port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sService, localPort, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoService(datasourceName string, regionHost string, subaccount string, k8sCluster string, localPort int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster_host =  "%s"
	local_port = "%d"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, localPort, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoPort(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster_host =  "%s"
	k8s_service_id =  "%s"
	connections = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, k8sService, connections, enabled)
}

func ResourceSubaccountK8SServiceChannelWoConnections(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, localPort int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_k8s_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	k8s_cluster_host =  "%s"
	k8s_service_id =  "%s"
	local_port = "%d"
	enabled= "%t"
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, k8sService, localPort, enabled)
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
