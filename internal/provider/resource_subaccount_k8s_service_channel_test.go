package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSubaccountK8SServiceChannel(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_k8s_service_channel")
		rec.SetRealTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountK8SServiceChannel("test", "cf.eu12.hana.ondemand.com", "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8", "cp.ace9fb5.stage.kyma.ondemand.com:443", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f", 3000, 1, true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "k8s_cluster", "cp.ace9fb5.stage.kyma.ondemand.com:443"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "k8s_service", "29d4e6f6-8e7f-4882-b434-21a52bb75e0f"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "port", "3000"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "connections", "1"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "type", "K8S"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "enabled", "true"),
						resource.TestCheckResourceAttrSet("cloudconnector_subaccount_k8s_service_channel.test", "id"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "state.connected", "true"),
						resource.TestMatchResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("cloudconnector_subaccount_k8s_service_channel.test", "state.opened_connections", "1"),
					),
				},
				{
					ResourceName:      "cloudconnector_subaccount_k8s_service_channel.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: getImportStateForSubaccountK8SServiceChannel("cloudconnector_subaccount_k8s_service_channel.test"),
				},
			},
		})

	})

}

func ResourceSubaccountK8SServiceChannel(datasourceName string, regionHost string, subaccount string, k8sCluster string, k8sService string, port int64, connections int64, enabled bool) string {
	return fmt.Sprintf(`
	resource "cloudconnector_subaccount_k8s_service_channel" "%s" {
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
