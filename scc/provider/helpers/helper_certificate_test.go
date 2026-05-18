package helpers_test

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for parseSubjectDN function
func TestParseSubjectDN_AllFields(t *testing.T) {
	dn := "CN=testCert,EMAIL=test@example.com,L=Bangalore,OU=Engineering,O=SAP,ST=KA,C=IN"

	result := helpers.ParseSubjectDNFunc(dn)

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

	result := helpers.ParseSubjectDNFunc(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)

	assert.True(t, result.Email.IsNull())
	assert.True(t, result.Locality.IsNull())
	assert.True(t, result.OrganizationalUnit.IsNull())
	assert.True(t, result.State.IsNull())
	assert.True(t, result.Country.IsNull())
}

func TestParseSubjectDN_EmptyDN(t *testing.T) {
	result := helpers.ParseSubjectDNFunc("")

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

	result := helpers.ParseSubjectDNFunc(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
}

func TestParseSubjectDN_CaseInsensitiveKeys(t *testing.T) {
	dn := "cn=testCert,o=SAP,c=in"

	result := helpers.ParseSubjectDNFunc(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("in"), result.Country)
}

func TestParseSubjectDN_SpacesTrimmed(t *testing.T) {
	dn := " CN = testCert , O = SAP , C = IN "

	result := helpers.ParseSubjectDNFunc(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_MultipleOU_LastWins(t *testing.T) {
	dn := "CN=testCert,OU=Team1,OU=Team2"

	result := helpers.ParseSubjectDNFunc(dn)

	// Current implementation keeps last OU
	assert.Equal(t, types.StringValue("Team2"), result.OrganizationalUnit)
}

func TestParseSubjectDN_EmptyPartsIgnored(t *testing.T) {
	dn := "CN=test,,O=SAP, ,C=IN"

	result := helpers.ParseSubjectDNFunc(dn)

	assert.Equal(t, types.StringValue("test"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_DuplicateCN_LastWins(t *testing.T) {
	dn := "CN=one,CN=two"

	result := helpers.ParseSubjectDNFunc(dn)

	assert.Equal(t, types.StringValue("two"), result.CommonName)
}

// Tests for buildSubjectDN function
func TestBuildSubjectDN_AllFields(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName:         types.StringValue("testCert"),
		Email:              types.StringValue("test@sap.com"),
		Locality:           types.StringValue("Bangalore"),
		OrganizationalUnit: types.StringValue("BTP"),
		Organization:       types.StringValue("SAP"),
		State:              types.StringValue("KA"),
		Country:            types.StringValue("IN"),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	expected := "CN=testCert,EMAIL=test@sap.com,L=Bangalore,OU=BTP,O=SAP,ST=KA,C=IN"
	assert.Equal(t, expected, result)
}

func TestBuildSubjectDN_OnlyCN(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName: types.StringValue("testCert"),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	assert.Equal(t, "CN=testCert", result)
}

func TestBuildSubjectDN_NilSubject(t *testing.T) {
	result := helpers.BuildSubjectDNFunc(nil)
	assert.Equal(t, "", result)
}

func TestBuildSubjectDN_CNNull(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName: types.StringNull(),
	}

	result := helpers.BuildSubjectDNFunc(subject)
	assert.Equal(t, "", result)
}

func TestBuildSubjectDN_EmptyOptionalFieldsIgnored(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName:         types.StringValue("testCert"),
		Email:              types.StringValue(""),
		Locality:           types.StringNull(),
		OrganizationalUnit: types.StringValue(" "),
		Organization:       types.StringValue("SAP"),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	assert.Equal(t, "CN=testCert,O=SAP", result)
}

func TestBuildSubjectDN_SpacesTrimmed(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName:   types.StringValue(" testCert "),
		Organization: types.StringValue(" SAP "),
		Country:      types.StringValue(" IN "),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	assert.Equal(t, "CN=testCert,O=SAP,C=IN", result)
}

func TestBuildSubjectDN_FieldOrder(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Country:      types.StringValue("IN"),
		Organization: types.StringValue("SAP"),
		Email:        types.StringValue("a@sap.com"),
		Locality:     types.StringValue("BLR"),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	expected := "CN=cert,EMAIL=a@sap.com,L=BLR,O=SAP,C=IN"
	assert.Equal(t, expected, result)
}

func TestBuildSubjectDN_OptionalFieldsNull(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Email:        types.StringNull(),
		Organization: types.StringNull(),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	assert.Equal(t, "CN=cert", result)
}

func TestBuildSubjectDN_SpecialCharacters(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName:   types.StringValue("test-cert_123"),
		Organization: types.StringValue("SAP-SE"),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	assert.Equal(t, "CN=test-cert_123,O=SAP-SE", result)
}

func TestBuildSubjectDN_RoundTrip(t *testing.T) {
	input := &helpers.CertificateSubjectDNConfig{
		CommonName:   types.StringValue("testCert"),
		Organization: types.StringValue("SAP"),
		Country:      types.StringValue("IN"),
	}

	dn := helpers.BuildSubjectDNFunc(input)
	parsed := helpers.ParseSubjectDNFunc(dn)

	assert.Equal(t, input.CommonName, parsed.CommonName)
	assert.Equal(t, input.Organization, parsed.Organization)
	assert.Equal(t, input.Country, parsed.Country)
}

func TestBuildSubjectDN_UnknownValuesIgnored(t *testing.T) {
	subject := &helpers.CertificateSubjectDNConfig{
		CommonName: types.StringValue("cert"),
		Email:      types.StringUnknown(),
	}

	result := helpers.BuildSubjectDNFunc(subject)

	assert.Equal(t, "CN=cert", result)
}

func TestBuildSubjectDNObject_Valid(t *testing.T) {
	dn := &helpers.CertificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Organization: types.StringValue("SAP"),
		Country:      types.StringValue("IN"),
	}

	obj := helpers.BuildSubjectDNObjectFunc(dn)

	assert.False(t, obj.IsNull())

	var result helpers.CertificateSubjectDNConfig
	diags := obj.As(context.Background(), &result, basetypes.ObjectAsOptions{})

	assert.False(t, diags.HasError())
	assert.Equal(t, "cert", result.CommonName.ValueString())
	assert.Equal(t, "SAP", result.Organization.ValueString())
	assert.Equal(t, "IN", result.Country.ValueString())
}

func TestBuildSubjectDNObject_Nil(t *testing.T) {
	obj := helpers.BuildSubjectDNObjectFunc(nil)

	assert.True(t, obj.IsNull())
}

// Tests for expandSubjectDN function
func TestExpandSubjectDN_Null(t *testing.T) {
	ctx := context.Background()
	obj := types.ObjectNull(helpers.SubjectDNAttrTypes.AttrTypes)

	res, diags := helpers.ExpandSubjectDNFunc(ctx, obj)

	assert.Nil(t, res)
	assert.False(t, diags.HasError())
}
func TestExpandSubjectDN_Valid(t *testing.T) {
	ctx := context.Background()

	obj, diags := types.ObjectValue(
		helpers.SubjectDNAttrTypes.AttrTypes,
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

	res, diags := helpers.ExpandSubjectDNFunc(ctx, obj)

	assert.False(t, diags.HasError())
	assert.Equal(t, "cert", res.CommonName.ValueString())
	assert.Equal(t, "SAP", res.Organization.ValueString())
}

func TestExpandSubjectDN_Unknown(t *testing.T) {
	ctx := context.Background()
	obj := types.ObjectUnknown(helpers.SubjectDNAttrTypes.AttrTypes)

	res, diags := helpers.ExpandSubjectDNFunc(ctx, obj)

	assert.Nil(t, res)
	assert.False(t, diags.HasError())
}

// Tests for validatePEMData function
func TestValidatePEMData_Empty(t *testing.T) {
	diags := helpers.ValidatePEMDataFunc("")

	assert.True(t, diags.HasError())
}

func TestValidatePEMData_InvalidPEM(t *testing.T) {
	diags := helpers.ValidatePEMDataFunc("not a pem")
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_UnsupportedType(t *testing.T) {
	pem := `-----BEGIN FOO-----
abcd
-----END FOO-----`

	diags := helpers.ValidatePEMDataFunc(pem)
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_ValidCert(t *testing.T) {
	var validCert = tfutils.GenerateTestCert(t)
	diags := helpers.ValidatePEMDataFunc(validCert)
	assert.False(t, diags.HasError())
}

// Tests for validatePEMChain function
func TestValidatePEMChain_Empty(t *testing.T) {
	diags := helpers.ValidatePEMChainFunc("")
	assert.True(t, diags.HasError())
}

func TestValidatePEMChain_InvalidBlock(t *testing.T) {
	data := `-----BEGIN PRIVATE KEY-----
abcd
-----END PRIVATE KEY-----`

	diags := helpers.ValidatePEMChainFunc(data)
	assert.True(t, diags.HasError())
}

func TestValidatePEMChain_MultipleCerts(t *testing.T) {
	var validCert = tfutils.GenerateTestCert(t)
	data := validCert + "\n" + validCert
	diags := helpers.ValidatePEMChainFunc(data)
	assert.False(t, diags.HasError())
}

func TestValidatePEMChain_InvalidCertificateBytes(t *testing.T) {
	data := `-----BEGIN CERTIFICATE-----
abcd
-----END CERTIFICATE-----`

	diags := helpers.ValidatePEMChainFunc(data)
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

	client := tfutils.NewTestClient(t, server)

	diags := helpers.UploadSignedChainFunc(client, "", "-----BEGIN CERTIFICATE-----test")
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

	client := tfutils.NewTestClient(t, server)

	diags := helpers.UploadSignedChainFunc(client, "", "test")
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

	client := tfutils.NewTestClient(t, server)

	body, diags := helpers.GetCertificateBinaryFunc(client, "")
	assert.False(t, diags.HasError())
	assert.Equal(t, expected, body)
}

func TestGetCertificateBinary_ReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := tfutils.NewTestClient(t, server)

	body, diags := helpers.GetCertificateBinaryFunc(client, "")
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

	client := tfutils.NewTestClient(t, server)

	diags := helpers.UploadPKCS12CertificateFunc(
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

	client := tfutils.NewTestClient(t, server)

	diags := helpers.UploadPKCS12CertificateFunc(
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

	client := tfutils.NewTestClient(t, server)

	diags := helpers.UploadPKCS12CertificateFunc(
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

	client := tfutils.NewTestClient(t, server)

	diags := helpers.UploadPKCS12CertificateFunc(
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

	diags := helpers.UploadPKCS12CertificateFunc(
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

	client := tfutils.NewTestClient(t, server)

	diags := helpers.UploadPKCS12CertificateFunc(
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

	plan := model.PKCS12SystemCertificateResourceConfig{
		PKCS12Certificate: types.StringValue(encoded),
		KeyPassword:       types.StringNull(),
	}

	data, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, raw, data)
}

func TestValidatePKCS12Inputs_RawInput(t *testing.T) {
	plan := model.PKCS12SystemCertificateResourceConfig{
		PKCS12Certificate: types.StringValue("rawdata"),
		KeyPassword:       types.StringNull(),
	}

	data, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("rawdata"), data)
}

func TestValidatePKCS12Inputs_EmptyCertificate(t *testing.T) {
	plan := model.PKCS12SystemCertificateResourceConfig{
		PKCS12Certificate: types.StringValue(""),
		KeyPassword:       types.StringNull(),
	}

	_, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

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

	model, diags := helpers.BuildCertificateModelFunc(ctx, cert)

	assert.False(t, diags.HasError())

	assert.Equal(t, "TestCA", model.Issuer.ValueString())
	assert.Equal(t, "12345", model.SerialNumber.ValueString())

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

	model, diags := helpers.BuildCertificateModelFunc(ctx, cert)

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

	model, diags := helpers.BuildCertificateModelWithSANFunc(ctx, cert)

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

	model, diags := helpers.BuildCertificateModelWithSANFunc(ctx, cert)

	assert.False(t, diags.HasError())
	assert.True(t, model.SubjectAltNames.IsNull())
}
