package datasources_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/datasources"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
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

func TestDataSourceBackendTrustStore_Read_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ds := &datasources.BackendTrustStoreDataSource{Client: tfutils.NewTestClient(t, srv)}

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)

	schemaType := schemaResp.Schema.Type().TerraformType(context.Background())
	raw := tftypes.NewValue(schemaType, nil)

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := &datasource.ReadResponse{}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
