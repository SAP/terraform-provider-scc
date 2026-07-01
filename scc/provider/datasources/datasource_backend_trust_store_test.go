package datasources_test

import (
	"fmt"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceBackendTrustStore(t *testing.T) {
	// To run this test, you need to have backend trust store configured with a certificate in your Cloud Connector Instance.
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_backend_trust_store")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceBackendTrustStore("scc_bts"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scc_backend_trust_store.scc_bts", "trust_all_backends"),
						resource.TestCheckResourceAttr("data.scc_backend_trust_store.scc_bts", "trusted_backends.#", "1"),

						resource.TestCheckResourceAttrSet("data.scc_backend_trust_store.scc_bts", "trusted_backends.0.alias"),
						resource.TestCheckResourceAttrSet("data.scc_backend_trust_store.scc_bts", "trusted_backends.0.subject_dn.cn"),
						resource.TestCheckResourceAttrSet("data.scc_backend_trust_store.scc_bts", "trusted_backends.0.issuer"),
						resource.TestMatchResourceAttr("data.scc_backend_trust_store.scc_bts", "trusted_backends.0.valid_to", tfutils.RegexpValidTimeStamp),
					),
				},
			},
		})

	})
}

func DataSourceBackendTrustStore(datasourceName string) string {
	return fmt.Sprintf(`data "scc_backend_trust_store" "%s" {}`, datasourceName)
}
