package datasources_test

import (
	"fmt"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceProxySettings(t *testing.T) {
	// To run this test, you need to have proxy settings configured in your Cloud Connector Instance.
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_proxy_settings")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceProxySettings("scc_proxy_settings"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scc_proxy_settings.scc_proxy_settings", "host"),
						resource.TestCheckResourceAttrSet("data.scc_proxy_settings.scc_proxy_settings", "port"),
						resource.TestCheckResourceAttrSet("data.scc_proxy_settings.scc_proxy_settings", "user"),
						resource.TestCheckResourceAttrSet("data.scc_proxy_settings.scc_proxy_settings", "password"),
					),
				},
			},
		})

	})
}

func DataSourceProxySettings(datasourceName string) string {
	return fmt.Sprintf(`data "scc_proxy_settings" "%s" {}`, datasourceName)
}
