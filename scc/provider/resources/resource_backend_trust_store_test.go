package resources_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/SAP/terraform-provider-scc/scc/provider/resources"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackendTrustStore_Metadata(t *testing.T) {
	r := resources.NewBackendTrustStoreResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "scc"}, resp)
	assert.Equal(t, "scc_backend_trust_store", resp.TypeName)
}

func TestBackendTrustStore_Schema_Attributes(t *testing.T) {
	r := resources.NewBackendTrustStoreResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	for _, name := range []string{"certificate", "alias", "subject_dn", "valid_to", "issuer"} {
		assert.Contains(t, resp.Schema.Attributes, name)
	}
}

func TestBackendTrustStore_Schema_CertificateIsRequiredAndSensitive(t *testing.T) {
	r := resources.NewBackendTrustStoreResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	type requiredSensitive interface {
		IsRequired() bool
		IsSensitive() bool
	}
	cert, ok := resp.Schema.Attributes["certificate"]
	require.True(t, ok)
	rs, ok := cert.(requiredSensitive)
	require.True(t, ok)
	assert.True(t, rs.IsRequired())
	assert.True(t, rs.IsSensitive())
}

func TestBackendTrustStore_Schema_ComputedAttributes(t *testing.T) {
	r := resources.NewBackendTrustStoreResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	type computed interface{ IsComputed() bool }
	for _, name := range []string{"alias", "subject_dn", "valid_to", "issuer"} {
		attr, ok := resp.Schema.Attributes[name]
		require.True(t, ok, "attribute %q missing", name)
		c, ok := attr.(computed)
		require.True(t, ok)
		assert.True(t, c.IsComputed(), "attribute %q should be computed", name)
	}
}

func TestBackendTrustStore_Configure_Success(t *testing.T) {
	r := resources.NewBackendTrustStoreResource().(*resources.BackendTrustStoreResource)
	client := &api.RestApiClient{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, client, r.Client)
}

func TestBackendTrustStore_Configure_NilProviderData(t *testing.T) {
	r := resources.NewBackendTrustStoreResource().(*resources.BackendTrustStoreResource)
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: nil}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.Client)
}

func TestBackendTrustStore_Configure_WrongType(t *testing.T) {
	r := resources.NewBackendTrustStoreResource().(*resources.BackendTrustStoreResource)
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "wrong-type"}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Upload_InvalidPEM(t *testing.T) {
	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}

	_, diags := resources.UploadBackendCertificateFunc(r, context.Background(), "not-a-pem")

	assert.True(t, diags.HasError())
	assertDiagContains(t, diags, "Invalid Certificate")
}

func TestBackendTrustStore_Upload_UploadFails(t *testing.T) {
	old := helpers.UploadBackendTrustStoreCertificateFunc
	defer func() { helpers.UploadBackendTrustStoreCertificateFunc = old }()

	helpers.UploadBackendTrustStoreCertificateFunc = func(_ *api.RestApiClient, _, _ string) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("upload failed", "network error")
		return d
	}

	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}
	_, diags := resources.UploadBackendCertificateFunc(r, context.Background(), tfutils.GenerateTestCert(t))

	assert.True(t, diags.HasError())
}

func TestBackendTrustStore_Upload_ReadAfterUploadFails(t *testing.T) {
	old := helpers.UploadBackendTrustStoreCertificateFunc
	defer func() { helpers.UploadBackendTrustStoreCertificateFunc = old }()

	helpers.UploadBackendTrustStoreCertificateFunc = func(_ *api.RestApiClient, _, _ string) diag.Diagnostics {
		return nil
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	_, diags := resources.UploadBackendCertificateFunc(r, context.Background(), tfutils.GenerateTestCert(t))

	assert.True(t, diags.HasError())
}

func TestBackendTrustStore_Upload_CertNotFoundInStore(t *testing.T) {
	old := helpers.UploadBackendTrustStoreCertificateFunc
	defer func() { helpers.UploadBackendTrustStoreCertificateFunc = old }()

	helpers.UploadBackendTrustStoreCertificateFunc = func(_ *api.RestApiClient, _, _ string) diag.Diagnostics {
		return nil
	}

	// Return a trust store that does NOT contain the uploaded cert.
	emptyStore := apiobjects.BackendTrustStoreConfiguration{TrustedBackends: []apiobjects.TrustedBackends{}}
	srv := httptest.NewServer(jsonHandler(t, emptyStore))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	_, diags := resources.UploadBackendCertificateFunc(r, context.Background(), tfutils.GenerateTestCert(t))

	assert.True(t, diags.HasError())
	assertDiagContains(t, diags, "Failed to Find Uploaded Certificate")
}

func TestBackendTrustStore_Upload_Success(t *testing.T) {
	old := helpers.UploadBackendTrustStoreCertificateFunc
	defer func() { helpers.UploadBackendTrustStoreCertificateFunc = old }()

	helpers.UploadBackendTrustStoreCertificateFunc = func(_ *api.RestApiClient, _, _ string) diag.Diagnostics {
		return nil
	}

	certPEM := tfutils.GenerateTestCert(t)
	trustedBackend := buildMatchingTrustedBackend(t, certPEM)

	store := apiobjects.BackendTrustStoreConfiguration{
		TrustedBackends: []apiobjects.TrustedBackends{trustedBackend},
	}
	srv := httptest.NewServer(jsonHandler(t, store))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	state, diags := resources.UploadBackendCertificateFunc(r, context.Background(), certPEM)

	require.False(t, diags.HasError())
	require.NotNil(t, state)
	assert.Equal(t, trustedBackend.Alias, state.Alias.ValueString())
	assert.Equal(t, certPEM, state.Certificate.ValueString())
}

func TestBackendTrustStore_Read_NullState(t *testing.T) {
	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), resource.ReadRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Read_APIFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	state := buildTrustStoreState(t, r, "my-alias", "")

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Read_CertRemovedFromStore(t *testing.T) {
	store := apiobjects.BackendTrustStoreConfiguration{TrustedBackends: []apiobjects.TrustedBackends{}}
	srv := httptest.NewServer(jsonHandler(t, store))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	state := buildTrustStoreState(t, r, "gone-alias", "")

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	// Framework removes the resource when RemoveResource is called.
	assert.True(t, resp.State.Raw.IsNull())
}

func TestBackendTrustStore_Read_MatchingAlias(t *testing.T) {
	alias := "cert-alias"
	certPEM := tfutils.GenerateTestCert(t)
	store := apiobjects.BackendTrustStoreConfiguration{
		TrustedBackends: []apiobjects.TrustedBackends{
			{Alias: alias, SubjectDN: "CN=test-cert", Issuer: "CN=test-cert", ValidTo: 0},
		},
	}
	srv := httptest.NewServer(jsonHandler(t, store))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	state := buildTrustStoreState(t, r, alias, certPEM)

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Read_AliasNotInStoreRemovesResource(t *testing.T) {
	store := apiobjects.BackendTrustStoreConfiguration{
		TrustedBackends: []apiobjects.TrustedBackends{
			{Alias: "other-alias", SubjectDN: "CN=other", Issuer: "CN=other", ValidTo: 0},
		},
	}
	srv := httptest.NewServer(jsonHandler(t, store))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	state := buildTrustStoreState(t, r, "my-alias", "")

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.State.Raw.IsNull())
}

func TestBackendTrustStore_Update_SameCertificateIsNoop(t *testing.T) {
	certPEM := tfutils.GenerateTestCert(t)
	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}

	state := buildTrustStoreState(t, r, "alias", certPEM)
	plan := buildTrustStoreStatePlan(t, r, "alias", certPEM)

	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Update_DifferentCertificateErrors(t *testing.T) {
	cert1 := tfutils.GenerateTestCert(t)
	cert2 := tfutils.GenerateTestCert(t)
	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}

	state := buildTrustStoreState(t, r, "alias", cert1)
	plan := buildTrustStoreStatePlan(t, r, "alias", cert2)

	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assertDiagContains(t, resp.Diagnostics, "Updating Backend Trust Store Certificates is Not Supported")
}

func TestBackendTrustStore_Delete_NullState(t *testing.T) {
	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), resource.DeleteRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Delete_Success(t *testing.T) {
	var deletedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodDelete {
			deletedPath = req.URL.Path
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	alias := "my-alias"
	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	state := buildTrustStoreState(t, r, alias, "")

	// resp.State must carry the schema so RemoveResource can set Raw to null.
	resp := &resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Contains(t, deletedPath, alias)
}

func TestBackendTrustStore_Delete_APIFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	r := &resources.BackendTrustStoreResource{Client: tfutils.NewTestClient(t, srv)}
	state := buildTrustStoreState(t, r, "my-alias", "")

	resp := &resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Create_UploadFuncFails(t *testing.T) {
	old := resources.UploadBackendCertificateFunc
	defer func() { resources.UploadBackendCertificateFunc = old }()

	resources.UploadBackendCertificateFunc = func(_ *resources.BackendTrustStoreResource, _ context.Context, _ string) (*model.BackendTrustStoreResourceConfig, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("upload error", "fail")
		return nil, d
	}

	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}
	certPEM := tfutils.GenerateTestCert(t)
	plan := buildTrustStoreStatePlan(t, r, "alias", certPEM)

	resp := &resource.CreateResponse{State: buildTrustStoreState(t, r, "alias", certPEM)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendTrustStore_Create_UploadFuncSucceeds(t *testing.T) {
	old := resources.UploadBackendCertificateFunc
	defer func() { resources.UploadBackendCertificateFunc = old }()

	certPEM := tfutils.GenerateTestCert(t)
	returned := testValidBackendTrustStoreConfig(certPEM, "returned-alias")

	resources.UploadBackendCertificateFunc = func(_ *resources.BackendTrustStoreResource, _ context.Context, _ string) (*model.BackendTrustStoreResourceConfig, diag.Diagnostics) {
		return &returned, nil
	}

	r := &resources.BackendTrustStoreResource{Client: &api.RestApiClient{}}
	plan := buildTrustStoreStatePlan(t, r, "alias", certPEM)

	resp := &resource.CreateResponse{State: buildTrustStoreState(t, r, "alias", certPEM)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func jsonHandler(t *testing.T, v any) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(v); err != nil {
			t.Errorf("jsonHandler: encode failed: %v", err)
		}
	})
}

// buildMatchingTrustedBackend creates a TrustedBackends whose fields exactly
// match the certificate encoded in certPEM, so MatchesBackendCertificate returns true.
func buildMatchingTrustedBackend(t *testing.T, certPEM string) apiobjects.TrustedBackends {
	t.Helper()
	parsed := parsePEMCert(t, certPEM)
	return apiobjects.TrustedBackends{
		Alias:     "test-alias",
		SubjectDN: fmt.Sprintf("CN=%s", parsed.Subject.CommonName),
		Issuer:    fmt.Sprintf("CN=%s", parsed.Issuer.CommonName),
		ValidTo:   parsed.NotAfter.UnixMilli(),
	}
}

func parsePEMCert(t *testing.T, certPEM string) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode([]byte(certPEM))
	require.NotNil(t, block, "failed to decode PEM block")
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	return cert
}

// tftypes shapes for the resource schema.
var btsSubjectDNType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"cn": tftypes.String, "email": tftypes.String, "l": tftypes.String,
		"ou": tftypes.String, "o": tftypes.String, "st": tftypes.String, "c": tftypes.String,
	},
}

var btsObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"certificate": tftypes.String,
		"alias":       tftypes.String,
		"subject_dn":  btsSubjectDNType,
		"valid_to":    tftypes.String,
		"issuer":      tftypes.String,
	},
}

func btsSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := resources.NewBackendTrustStoreResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	return *resp
}

func btsRawValue(alias, certPEM string) tftypes.Value {
	return tftypes.NewValue(btsObjectType, map[string]tftypes.Value{
		"certificate": tftypes.NewValue(tftypes.String, certPEM),
		"alias":       tftypes.NewValue(tftypes.String, alias),
		"subject_dn":  tftypes.NewValue(btsSubjectDNType, nil),
		"valid_to":    tftypes.NewValue(tftypes.String, nil),
		"issuer":      tftypes.NewValue(tftypes.String, nil),
	})
}

func buildTrustStoreState(t *testing.T, _ *resources.BackendTrustStoreResource, alias, certPEM string) tfsdk.State {
	t.Helper()
	s := btsSchema(t)
	return tfsdk.State{Schema: s.Schema, Raw: btsRawValue(alias, certPEM)}
}

func buildTrustStoreStatePlan(t *testing.T, _ *resources.BackendTrustStoreResource, alias, certPEM string) tfsdk.Plan {
	t.Helper()
	s := btsSchema(t)
	return tfsdk.Plan{Schema: s.Schema, Raw: btsRawValue(alias, certPEM)}
}

func testValidBackendTrustStoreConfig(certPEM, alias string) model.BackendTrustStoreResourceConfig {
	dn := helpers.ParseSubjectDNFunc("CN=test-cert")
	return model.BackendTrustStoreResourceConfig{
		Certificate: types.StringValue(certPEM),
		Alias:       types.StringValue(alias),
		SubjectDN:   helpers.BuildSubjectDNObjectFunc(dn),
		Issuer:      types.StringValue("CN=test-issuer"),
		ValidTo:     types.StringNull(),
	}
}

func assertDiagContains(t *testing.T, diags diag.Diagnostics, substr string) {
	t.Helper()
	for _, d := range diags {
		if strings.Contains(d.Summary(), substr) || strings.Contains(d.Detail(), substr) {
			return
		}
	}
	t.Errorf("expected diagnostics to contain %q, got: %v", substr, diags)
}
