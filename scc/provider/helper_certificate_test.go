package provider

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, server *httptest.Server) *api.RestApiClient {
	baseURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	return &api.RestApiClient{
		BaseURL: baseURL,
		Client:  server.Client(),
	}
}

// Tests for parseSubjectDN function
func TestParseSubjectDN_AllFields(t *testing.T) {
	dn := "CN=testCert,EMAIL=test@example.com,L=Bangalore,OU=Engineering,O=SAP,ST=KA,C=IN"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("test@example.com"), result.Email)
	assert.Equal(t, types.StringValue("Bangalore"), result.Locality)
	assert.Equal(t, types.StringValue("Engineering"), result.OrganizationalUnit)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("KA"), result.State)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_MissingOptionalFields(t *testing.T) {
	dn := "CN=testCert,O=SAP"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)

	assert.True(t, result.Email.IsNull())
	assert.True(t, result.Locality.IsNull())
	assert.True(t, result.OrganizationalUnit.IsNull())
	assert.True(t, result.State.IsNull())
	assert.True(t, result.Country.IsNull())
}

func TestParseSubjectDN_EmptyDN(t *testing.T) {
	result := parseSubjectDN("")

	assert.True(t, result.CommonName.IsNull())
	assert.True(t, result.Email.IsNull())
	assert.True(t, result.Locality.IsNull())
	assert.True(t, result.OrganizationalUnit.IsNull())
	assert.True(t, result.Organization.IsNull())
	assert.True(t, result.State.IsNull())
	assert.True(t, result.Country.IsNull())
}

func TestParseSubjectDN_UnknownFieldsIgnored(t *testing.T) {
	dn := "CN=testCert,XYZ=value,O=SAP"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
}

func TestParseSubjectDN_CaseInsensitiveKeys(t *testing.T) {
	dn := "cn=testCert,o=SAP,c=in"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("in"), result.Country)
}

func TestParseSubjectDN_SpacesTrimmed(t *testing.T) {
	dn := " CN = testCert , O = SAP , C = IN "

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_MultipleOU_LastWins(t *testing.T) {
	dn := "CN=testCert,OU=Team1,OU=Team2"

	result := parseSubjectDN(dn)

	// Current implementation keeps last OU
	assert.Equal(t, types.StringValue("Team2"), result.OrganizationalUnit)
}

func TestParseSubjectDN_EmptyPartsIgnored(t *testing.T) {
	dn := "CN=test,,O=SAP, ,C=IN"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("test"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_DuplicateCN_LastWins(t *testing.T) {
	dn := "CN=one,CN=two"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("two"), result.CommonName)
}

// Tests for buildSubjectDN function
func TestBuildSubjectDN_AllFields(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName:         types.StringValue("testCert"),
		Email:              types.StringValue("test@sap.com"),
		Locality:           types.StringValue("Bangalore"),
		OrganizationalUnit: types.StringValue("BTP"),
		Organization:       types.StringValue("SAP"),
		State:              types.StringValue("KA"),
		Country:            types.StringValue("IN"),
	}

	result := buildSubjectDN(subject)

	expected := "CN=testCert,EMAIL=test@sap.com,L=Bangalore,OU=BTP,O=SAP,ST=KA,C=IN"
	assert.Equal(t, expected, result)
}

func TestBuildSubjectDN_OnlyCN(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName: types.StringValue("testCert"),
	}

	result := buildSubjectDN(subject)

	assert.Equal(t, "CN=testCert", result)
}

func TestBuildSubjectDN_NilSubject(t *testing.T) {
	result := buildSubjectDN(nil)
	assert.Equal(t, "", result)
}

func TestBuildSubjectDN_CNNull(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName: types.StringNull(),
	}

	result := buildSubjectDN(subject)
	assert.Equal(t, "", result)
}

func TestBuildSubjectDN_EmptyOptionalFieldsIgnored(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName:         types.StringValue("testCert"),
		Email:              types.StringValue(""),
		Locality:           types.StringNull(),
		OrganizationalUnit: types.StringValue(" "),
		Organization:       types.StringValue("SAP"),
	}

	result := buildSubjectDN(subject)

	assert.Equal(t, "CN=testCert,O=SAP", result)
}

func TestBuildSubjectDN_SpacesTrimmed(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName:   types.StringValue(" testCert "),
		Organization: types.StringValue(" SAP "),
		Country:      types.StringValue(" IN "),
	}

	result := buildSubjectDN(subject)

	assert.Equal(t, "CN=testCert,O=SAP,C=IN", result)
}

func TestBuildSubjectDN_FieldOrder(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Country:      types.StringValue("IN"),
		Organization: types.StringValue("SAP"),
		Email:        types.StringValue("a@sap.com"),
		Locality:     types.StringValue("BLR"),
	}

	result := buildSubjectDN(subject)

	expected := "CN=cert,EMAIL=a@sap.com,L=BLR,O=SAP,C=IN"
	assert.Equal(t, expected, result)
}

func TestBuildSubjectDN_OptionalFieldsNull(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Email:        types.StringNull(),
		Organization: types.StringNull(),
	}

	result := buildSubjectDN(subject)

	assert.Equal(t, "CN=cert", result)
}

func TestBuildSubjectDN_SpecialCharacters(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName:   types.StringValue("test-cert_123"),
		Organization: types.StringValue("SAP-SE"),
	}

	result := buildSubjectDN(subject)

	assert.Equal(t, "CN=test-cert_123,O=SAP-SE", result)
}

func TestBuildSubjectDN_RoundTrip(t *testing.T) {
	input := &certificateSubjectDNConfig{
		CommonName:   types.StringValue("testCert"),
		Organization: types.StringValue("SAP"),
		Country:      types.StringValue("IN"),
	}

	dn := buildSubjectDN(input)
	parsed := parseSubjectDN(dn)

	assert.Equal(t, input.CommonName, parsed.CommonName)
	assert.Equal(t, input.Organization, parsed.Organization)
	assert.Equal(t, input.Country, parsed.Country)
}

func TestBuildSubjectDN_UnknownValuesIgnored(t *testing.T) {
	subject := &certificateSubjectDNConfig{
		CommonName: types.StringValue("cert"),
		Email:      types.StringUnknown(),
	}

	result := buildSubjectDN(subject)

	assert.Equal(t, "CN=cert", result)
}

func TestBuildSubjectDNObject_Valid(t *testing.T) {
	dn := &certificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Organization: types.StringValue("SAP"),
		Country:      types.StringValue("IN"),
	}

	obj := buildSubjectDNObject(dn)

	assert.False(t, obj.IsNull())

	var result certificateSubjectDNConfig
	diags := obj.As(context.Background(), &result, basetypes.ObjectAsOptions{})

	assert.False(t, diags.HasError())
	assert.Equal(t, "cert", result.CommonName.ValueString())
	assert.Equal(t, "SAP", result.Organization.ValueString())
	assert.Equal(t, "IN", result.Country.ValueString())
}

func TestBuildSubjectDNObject_Nil(t *testing.T) {
	obj := buildSubjectDNObject(nil)

	assert.True(t, obj.IsNull())
}

// Tests for expandSubjectDN function
func TestExpandSubjectDN_Null(t *testing.T) {
	ctx := context.Background()
	obj := types.ObjectNull(subjectDNAttrTypes.AttrTypes)

	res, diags := expandSubjectDN(ctx, obj)

	assert.Nil(t, res)
	assert.False(t, diags.HasError())
}
func TestExpandSubjectDN_Valid(t *testing.T) {
	ctx := context.Background()

	obj, diags := types.ObjectValue(
		subjectDNAttrTypes.AttrTypes,
		map[string]attr.Value{
			"cn":    types.StringValue("cert"),
			"email": types.StringNull(),
			"l":     types.StringNull(),
			"ou":    types.StringNull(),
			"o":     types.StringValue("SAP"),
			"st":    types.StringNull(),
			"c":     types.StringNull(),
		},
	)

	assert.False(t, diags.HasError())

	res, diags := expandSubjectDN(ctx, obj)

	assert.False(t, diags.HasError())
	assert.Equal(t, "cert", res.CommonName.ValueString())
	assert.Equal(t, "SAP", res.Organization.ValueString())
}

func TestExpandSubjectDN_Unknown(t *testing.T) {
	ctx := context.Background()
	obj := types.ObjectUnknown(subjectDNAttrTypes.AttrTypes)

	res, diags := expandSubjectDN(ctx, obj)

	assert.Nil(t, res)
	assert.False(t, diags.HasError())
}

// Tests for validatePEMData function
func TestValidatePEMData_Empty(t *testing.T) {
	diags := validatePEMData("")
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_InvalidPEM(t *testing.T) {
	diags := validatePEMData("not a pem")
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_UnsupportedType(t *testing.T) {
	pem := `-----BEGIN FOO-----
abcd
-----END FOO-----`

	diags := validatePEMData(pem)
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_ValidCert(t *testing.T) {
	var validCert = generateTestCert(t)
	diags := validatePEMData(validCert)
	assert.False(t, diags.HasError())
}

// Tests for validatePEMChain function
func TestValidatePEMChain_Empty(t *testing.T) {
	diags := validatePEMChain("")
	assert.True(t, diags.HasError())
}

func TestValidatePEMChain_InvalidBlock(t *testing.T) {
	data := `-----BEGIN PRIVATE KEY-----
abcd
-----END PRIVATE KEY-----`

	diags := validatePEMChain(data)
	assert.True(t, diags.HasError())
}

func TestValidatePEMChain_MultipleCerts(t *testing.T) {
	var validCert = generateTestCert(t)
	data := validCert + "\n" + validCert
	diags := validatePEMChain(data)
	assert.False(t, diags.HasError())
}

func generateTestCert(t *testing.T) string {
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

func TestValidatePEMChain_InvalidCertificateBytes(t *testing.T) {
	data := `-----BEGIN CERTIFICATE-----
abcd
-----END CERTIFICATE-----`

	diags := validatePEMChain(data)
	assert.True(t, diags.HasError())
}

// Tests for uploadSignedChain function
func TestUploadSignedChain_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)

		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		file, _, err := r.FormFile("signedCertificate")
		require.NoError(t, err)
		defer func() {
			if err := file.Close(); err != nil {
				t.Errorf("failed to close file: %v", err)
			}
		}()

		content, err := io.ReadAll(file)
		require.NoError(t, err)

		assert.Contains(t, string(content), "CERTIFICATE")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)

	diags := uploadSignedChain(client, "", "-----BEGIN CERTIFICATE-----test")
	assert.False(t, diags.HasError())
}

func TestUploadSignedChain_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("INVALID_REQUEST")); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server)

	diags := uploadSignedChain(client, "", "test")
	assert.True(t, diags.HasError())
}

// Tests for getCertificateBinary function
func TestGetCertificateBinary_Success(t *testing.T) {
	expected := []byte("binary-cert")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/pkix-cert", r.Header.Get("Accept"))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(expected)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := newTestClient(t, server)

	body, diags := getCertificateBinaryFunc(client, "")
	assert.False(t, diags.HasError())
	assert.Equal(t, expected, body)
}

func TestGetCertificateBinary_ReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)

	body, diags := getCertificateBinaryFunc(client, "")
	assert.False(t, diags.HasError())
	assert.NotNil(t, body)
}

// Tests for uploadPKCS12Certificate function
func TestUploadPKCS12Certificate_Success_WithKeyPassword(t *testing.T) {
	expectedBytes := []byte("dummy-p12-content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		// Validate password fields
		assert.Equal(t, "storepass", r.FormValue("password"))
		assert.Equal(t, "keypass", r.FormValue("keyPassword"))

		// Validate file field
		file, _, err := r.FormFile("pkcs12")
		require.NoError(t, err)
		defer func() {
			if err := file.Close(); err != nil {
				t.Errorf("failed to close file: %v", err)
			}
		}()

		body, err := io.ReadAll(file)
		require.NoError(t, err)

		assert.Equal(t, expectedBytes, body)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)

	diags := uploadPKCS12Certificate(
		client,
		"",
		expectedBytes,
		"storepass",
		"keypass",
	)

	assert.False(t, diags.HasError())
}

func TestUploadPKCS12Certificate_Success_WithoutKeyPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		assert.Equal(t, "storepass", r.FormValue("password"))
		assert.Equal(t, "", r.FormValue("keyPassword"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)

	diags := uploadPKCS12Certificate(
		client,
		"",
		[]byte("data"),
		"storepass",
		"",
	)

	assert.False(t, diags.HasError())
}

func TestUploadPKCS12Certificate_HTTP400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("INVALID_REQUEST")); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server)

	diags := uploadPKCS12Certificate(
		client,
		"",
		[]byte("data"),
		"pass",
		"",
	)

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "INVALID_REQUEST")
}

func TestUploadPKCS12Certificate_HTTP500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("SERVER_ERROR")); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server)

	diags := uploadPKCS12Certificate(
		client,
		"",
		[]byte("data"),
		"pass",
		"",
	)

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "SERVER_ERROR")
}

func TestUploadPKCS12Certificate_DoRequestFailure(t *testing.T) {
	// Invalid server to force client error
	client := &api.RestApiClient{
		BaseURL: &url.URL{
			Scheme: "http",
			Host:   "invalid-host",
		},
		Client: &http.Client{},
	}

	diags := uploadPKCS12Certificate(
		client,
		"",
		[]byte("data"),
		"pass",
		"",
	)

	assert.True(t, diags.HasError())
}

func TestUploadPKCS12Certificate_EmptyPKCS12Bytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		file, _, err := r.FormFile("pkcs12")
		require.NoError(t, err)
		defer func() {
			if err := file.Close(); err != nil {
				t.Errorf("failed to close file: %v", err)
			}
		}()

		body, err := io.ReadAll(file)
		require.NoError(t, err)

		assert.Equal(t, []byte{}, body)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)

	diags := uploadPKCS12Certificate(
		client,
		"",
		[]byte{},
		"pass",
		"",
	)

	assert.False(t, diags.HasError())
}

// Tests for validatePKCS12Inputs function
func TestValidatePKCS12Inputs_Base64Input(t *testing.T) {
	raw := []byte("dummy-p12")
	encoded := base64.StdEncoding.EncodeToString(raw)

	plan := PKCS12SystemCertificateResourceConfig{
		PKCS12Certificate: types.StringValue(encoded),
		KeyPassword:       types.StringNull(),
	}

	data, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, raw, data)
}

func TestValidatePKCS12Inputs_RawInput(t *testing.T) {
	plan := PKCS12SystemCertificateResourceConfig{
		PKCS12Certificate: types.StringValue("rawdata"),
		KeyPassword:       types.StringNull(),
	}

	data, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("rawdata"), data)
}

func TestValidatePKCS12Inputs_EmptyCertificate(t *testing.T) {
	plan := PKCS12SystemCertificateResourceConfig{
		PKCS12Certificate: types.StringValue(""),
		KeyPassword:       types.StringNull(),
	}

	_, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.True(t, diags.HasError())
}

// Tests for buildCertificateModel function
func TestBuildCertificateModel(t *testing.T) {
	ctx := context.Background()

	cert := apiobjects.Certificate{
		Issuer:             "TestCA",
		SerialNumber:       "12345",
		SubjectDN:          "CN=test,O=SAP,C=IN",
		NotBeforeTimeStamp: time.Now().UnixMilli(),
		NotAfterTimeStamp:  time.Now().Add(time.Hour).UnixMilli(),
	}

	model, diags := buildCertificateModel(ctx, cert, []byte("pem-data"))

	assert.False(t, diags.HasError())

	assert.Equal(t, "TestCA", model.Issuer.ValueString())
	assert.Equal(t, "12345", model.SerialNumber.ValueString())
	assert.Equal(t, "pem-data", model.CertificatePEM.ValueString())

	assert.False(t, model.SubjectDN.IsNull())
}

func TestBuildCertificateModel_NoSubjectDN(t *testing.T) {
	ctx := context.Background()

	cert := apiobjects.Certificate{
		Issuer:             "TestCA",
		SerialNumber:       "123",
		NotBeforeTimeStamp: time.Now().UnixMilli(),
		NotAfterTimeStamp:  time.Now().UnixMilli(),
	}

	model, diags := buildCertificateModel(ctx, cert, []byte("pem"))

	assert.False(t, diags.HasError())
	assert.True(t, model.SubjectDN.IsNull())
}

// Tests for buildCertificateModelWithSAN function
func TestBuildCertificateModelWithSAN(t *testing.T) {
	ctx := context.Background()

	cert := apiobjects.Certificate{
		Issuer:             "TestCA",
		SerialNumber:       "123",
		NotBeforeTimeStamp: time.Now().UnixMilli(),
		NotAfterTimeStamp:  time.Now().UnixMilli(),
		SubjectDN:          "CN=test",
		SubjectAltNames: []apiobjects.SubjectAltNames{
			{
				Type:  "DNS",
				Value: "example.com",
			},
			{
				Type:  "IP",
				Value: "127.0.0.1",
			},
		},
	}

	model, diags := buildCertificateModelWithSAN(ctx, cert, []byte("pem"))

	assert.False(t, diags.HasError())
	assert.False(t, model.SubjectAltNames.IsNull())
}

func TestBuildCertificateModelWithSAN_NoSAN(t *testing.T) {
	ctx := context.Background()

	cert := apiobjects.Certificate{
		Issuer:             "TestCA",
		SerialNumber:       "123",
		NotBeforeTimeStamp: time.Now().UnixMilli(),
		NotAfterTimeStamp:  time.Now().UnixMilli(),
	}

	model, diags := buildCertificateModelWithSAN(ctx, cert, []byte("pem"))

	assert.False(t, diags.HasError())
	assert.True(t, model.SubjectAltNames.IsNull())
}

func buildSignedChainPlan(
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
		"signed_chain":    tftypes.String,
		"subject_dn":      subjectDNType,
		"valid_to":        tftypes.String,
		"valid_from":      tftypes.String,
		"issuer":          tftypes.String,
		"serial_number":   tftypes.String,
		"certificate_pem": tftypes.String,
	}

	values := map[string]tftypes.Value{
		"signed_chain":    tftypes.NewValue(tftypes.String, chain),
		"subject_dn":      tftypes.NewValue(subjectDNType, nil),
		"valid_to":        tftypes.NewValue(tftypes.String, nil),
		"valid_from":      tftypes.NewValue(tftypes.String, nil),
		"issuer":          tftypes.NewValue(tftypes.String, nil),
		"serial_number":   tftypes.NewValue(tftypes.String, nil),
		"certificate_pem": tftypes.NewValue(tftypes.String, nil),
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

func generateValidDERCert(t *testing.T) []byte {
	cert := generateTestCert(t)
	block, _ := pem.Decode([]byte(cert))
	return block.Bytes
}
