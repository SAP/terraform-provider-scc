package datasources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/datasources"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestDataSourceCACertificate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_ca_certficate")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceCACertificate("scc_ca_cert"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scc_ca_certificate.scc_ca_cert", "subject_dn.cn"),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "issuer", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "valid_from", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "valid_to", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "serial_number", tfutils.RegexpValidSerialNumber),
						resource.TestCheckResourceAttrSet("data.scc_ca_certificate.scc_ca_cert", "certificate_pem"),
					),
				},
			},
		})

	})
}

func DataSourceCACertificate(datasourceName string) string {
	return fmt.Sprintf(`data "scc_ca_certificate" "%s" {}`, datasourceName)
}

func TestDataSourceCACertificate_Read_MetadataAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ds := &datasources.CACertificateDataSource{Client: tfutils.NewTestClient(t, srv)}

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)

	schemaType := schemaResp.Schema.Type().TerraformType(context.Background())
	raw := tftypes.NewValue(schemaType, nil)

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := &datasource.ReadResponse{}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestDataSourceCACertificate_Read_BinaryAPIError(t *testing.T) {
	// Metadata succeeds but binary download fails.
	metadataResponse := map[string]any{
		"subjectDN":    "CN=test-ca",
		"issuer":       "CN=test-ca",
		"notAfter":     0,
		"notBefore":    0,
		"serialNumber": "aa:bb:cc",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept == "application/pkix-cert" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metadataResponse)
	}))
	defer srv.Close()

	ds := &datasources.CACertificateDataSource{Client: tfutils.NewTestClient(t, srv)}

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)

	schemaType := schemaResp.Schema.Type().TerraformType(context.Background())
	raw := tftypes.NewValue(schemaType, nil)

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := &datasource.ReadResponse{}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
