package tfutils

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/scc/provider/resources"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func NewTestClient(t *testing.T, server *httptest.Server) *api.RestApiClient {
	baseURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	return &api.RestApiClient{
		BaseURL: baseURL,
		Client:  server.Client(),
	}
}

func GenerateTestCert(t *testing.T) string {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		Subject: pkix.Name{
			CommonName: "test-cert",
		},
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&priv.PublicKey,
		priv,
	)
	require.NoError(t, err)

	var pemBuf bytes.Buffer
	err = pem.Encode(&pemBuf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	require.NoError(t, err)

	return pemBuf.String()
}

func GenerateValidDERCert(t *testing.T) []byte {
	cert := GenerateTestCert(t)
	block, _ := pem.Decode([]byte(cert))
	return block.Bytes
}

func BuildSignedChainPlan(
	ctx context.Context,
	r resource.Resource,
	chain string,
	includeSAN bool,
) tfsdk.Plan {

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	subjectDNType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"cn":    tftypes.String,
			"email": tftypes.String,
			"l":     tftypes.String,
			"ou":    tftypes.String,
			"o":     tftypes.String,
			"st":    tftypes.String,
			"c":     tftypes.String,
		},
	}

	attrTypes := map[string]tftypes.Type{
		"signed_chain":  tftypes.String,
		"subject_dn":    subjectDNType,
		"valid_to":      tftypes.String,
		"valid_from":    tftypes.String,
		"issuer":        tftypes.String,
		"serial_number": tftypes.String,
	}

	values := map[string]tftypes.Value{
		"signed_chain":  tftypes.NewValue(tftypes.String, chain),
		"subject_dn":    tftypes.NewValue(subjectDNType, nil),
		"valid_to":      tftypes.NewValue(tftypes.String, nil),
		"valid_from":    tftypes.NewValue(tftypes.String, nil),
		"issuer":        tftypes.NewValue(tftypes.String, nil),
		"serial_number": tftypes.NewValue(tftypes.String, nil),
	}

	if _, ok := schemaResp.Schema.Attributes["certificate_pem"]; ok {
		attrTypes["certificate_pem"] = tftypes.String
		values["certificate_pem"] = tftypes.NewValue(tftypes.String, nil)
	}

	if includeSAN {
		sanType := tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"type":  tftypes.String,
				"value": tftypes.String,
			},
		}

		attrTypes["subject_alternative_names"] = tftypes.List{
			ElementType: sanType,
		}

		values["subject_alternative_names"] = tftypes.NewValue(
			tftypes.List{ElementType: sanType},
			nil,
		)
	}

	return tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			values,
		),
	}
}

func BuildSignedChainState(ctx context.Context, r *resources.SystemCertificateSignedChainResource, chain string) tfsdk.State {
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	subjectDNType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"c":     tftypes.String,
			"cn":    tftypes.String,
			"email": tftypes.String,
			"l":     tftypes.String,
			"o":     tftypes.String,
			"ou":    tftypes.String,
			"st":    tftypes.String,
		},
	}

	return tfsdk.State{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"signed_chain":    tftypes.String,
					"certificate_pem": tftypes.String,
					"issuer":          tftypes.String,
					"serial_number":   tftypes.String,
					"subject_dn":      subjectDNType,
					"valid_from":      tftypes.String,
					"valid_to":        tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"signed_chain":    tftypes.NewValue(tftypes.String, chain),
				"certificate_pem": tftypes.NewValue(tftypes.String, ""),
				"issuer":          tftypes.NewValue(tftypes.String, ""),
				"serial_number":   tftypes.NewValue(tftypes.String, ""),
				"subject_dn": tftypes.NewValue(subjectDNType, map[string]tftypes.Value{
					"c":     tftypes.NewValue(tftypes.String, ""),
					"cn":    tftypes.NewValue(tftypes.String, ""),
					"email": tftypes.NewValue(tftypes.String, ""),
					"l":     tftypes.NewValue(tftypes.String, ""),
					"o":     tftypes.NewValue(tftypes.String, ""),
					"ou":    tftypes.NewValue(tftypes.String, ""),
					"st":    tftypes.NewValue(tftypes.String, ""),
				}),
				"valid_from": tftypes.NewValue(tftypes.String, ""),
				"valid_to":   tftypes.NewValue(tftypes.String, ""),
			},
		),
	}
}
