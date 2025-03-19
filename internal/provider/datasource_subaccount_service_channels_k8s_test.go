package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccountServiceChannelsK8S(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccount_service_channels_k8s")
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
					Config: providerConfig("", user) + DataSourceSubaccountServiceChannels("cc_scs", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "credentials.region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "credentials.subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.#", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.k8s_cluster", "cp.c-15e4ac4.kyma.ondemand.com:443"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.k8s_service", "676969a3-096e-4ab1-adde-c9400fccc7ab"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.port", "9999"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.connections", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.type", "K8S"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.enabled", "false"),
						resource.TestCheckResourceAttrSet("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.id"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.state.connected", "false"),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.state.connected_since_time_stamp", regexp.MustCompile("^(0|\\d{13})$")),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount_service_channels_k8s.cc_scs", "subaccount_service_channels_k8s.0.state.opened_connections", "0"),
					),
				},
			},
		})

	})

}

func DataSourceSubaccountServiceChannels(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "cloudconnector_subaccount_service_channels_k8s" "%s" {
	credentials = {
		region_host = "%s"
		subaccount = "%s"
	}
	}
	`, datasourceName, regionHost, subaccountID)
}
