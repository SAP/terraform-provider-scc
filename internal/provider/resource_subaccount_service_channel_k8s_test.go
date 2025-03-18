package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceSubaccountServiceChannelK8S(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_service_channel_k8s")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + ResourceSubaccountServiceChannelK8S("test", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7", "cp.c-15e4ac4.kyma.ondemand.com:443", "676969a3-096e-4ab1-adde-c9400fccc7ab", 9999, 1),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "credentials.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "credentials.subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.k8s_cluster", "cp.c-15e4ac4.kyma.ondemand.com:443"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.k8s_service", "676969a3-096e-4ab1-adde-c9400fccc7ab"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.port", "9999"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.connections", "1"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.type", "K8S"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.enabled", "false"),
						resource.TestCheckResourceAttrSet("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.id"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.state.connected", "false"),
						resource.TestMatchResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.state.connected_since_time_stamp", regexp.MustCompile("^(0|\\d{13})$")),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_service_channel_k8s.test", "subaccount_service_channel_k8s.state.opened_connections", "0"),
					),
				},
			},
		})

	})

}

func ResourceSubaccountServiceChannelK8S(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, port int64, connections int64) string {
	return fmt.Sprintf(`
	resource "cloudconnector_subaccount_service_channel_k8s" "%s" {
	credentials = {
		region_host = "%s"
		subaccount = "%s"
	}
	subaccount_service_channel_k8s = {
		k8s_cluster =  "%s",
		k8s_service =  "%s",
		port = "%d",
		connections = "%d"
	}
	}
	`, datasourceName, regionHost, subaccount, k8sCluster, k8sService, port, connections)
}
