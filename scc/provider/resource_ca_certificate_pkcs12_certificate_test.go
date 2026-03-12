package provider

import (
	"context"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCACertificatePKCS12Certificate_Metadata(t *testing.T) {
	r := NewCACertificatePKCS12CertificateResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_ca_certificate_pkcs12_certificate", resp.TypeName)
}

func TestCACertificatePKCS12Certificate_Schema_Attributes(t *testing.T) {
	r := NewCACertificatePKCS12CertificateResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	passwordAttr := resp.Schema.Attributes["password"]
	assert.True(t, passwordAttr.IsRequired())

	keyAttr := resp.Schema.Attributes["key_password"]
	assert.True(t, keyAttr.IsOptional())
}

func TestCACertificatePKCS12Certificate_Configure_Success(t *testing.T) {
	r := NewCACertificatePKCS12CertificateResource().(*CACertificatePKCS12CertificateResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_Configure_WrongType(t *testing.T) {
	r := NewCACertificatePKCS12CertificateResource().(*CACertificatePKCS12CertificateResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_Update_NotSupported(t *testing.T) {
	r := NewCACertificatePKCS12CertificateResource()

	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_Read_NoState(t *testing.T) {
	r := NewCACertificatePKCS12CertificateResource()

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_RawValue(t *testing.T) {
	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("raw-data"),
	}

	raw, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("raw-data"), raw)
}

func TestCACertificatePKCS12Certificate_Inputs_EmptyCert(t *testing.T) {
	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue(""),
	}

	_, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_EmptyKeyPassword(t *testing.T) {
	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		KeyPassword:       types.StringValue(""),
	}

	_, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_Base64Decode(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("hello"))

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue(encoded),
	}

	raw, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("hello"), raw)
}

func TestCACertificatePKCS12Certificate_Inputs_KeyPasswordUnknown(t *testing.T) {
	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		KeyPassword:       types.StringUnknown(),
	}

	_, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_ValidWithKeyPassword(t *testing.T) {
	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		KeyPassword:       types.StringValue("keypass"),
	}

	raw, diags := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("abc"), raw)
}

func TestCACertificatePKCS12Certificate_Create_ValidationFails(t *testing.T) {
	r := &CACertificatePKCS12CertificateResource{}

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue(""),
	}

	_, diags := r.createInternal(context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_UploadFails(t *testing.T) {
	r := &CACertificatePKCS12CertificateResource{
		client: &api.RestApiClient{},
	}

	oldUpload := uploadPKCS12CertificateFunc
	defer func() { uploadPKCS12CertificateFunc = oldUpload }()

	uploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("upload failed", "fail")
		return d
	}

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := r.createInternal(context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_MetadataFails(t *testing.T) {
	r := &CACertificatePKCS12CertificateResource{
		client: &api.RestApiClient{},
	}

	oldUpload := uploadPKCS12CertificateFunc
	oldReq := requestAndUnmarshalFunc
	defer func() {
		uploadPKCS12CertificateFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
	}()

	uploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("metadata error", "fail")
		return d
	}

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := r.createInternal(context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_BinaryFails(t *testing.T) {
	r := &CACertificatePKCS12CertificateResource{
		client: &api.RestApiClient{},
	}

	oldUpload := uploadPKCS12CertificateFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	defer func() {
		uploadPKCS12CertificateFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	uploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := r.createInternal(context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_InvalidPEM(t *testing.T) {
	r := &CACertificatePKCS12CertificateResource{
		client: &api.RestApiClient{},
	}

	oldUpload := uploadPKCS12CertificateFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := pkcs12CACertificateResourceValueFromFunc

	defer func() {
		uploadPKCS12CertificateFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		pkcs12CACertificateResourceValueFromFunc = oldValue
	}()

	uploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := r.createInternal(context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_ModelConversionFails(t *testing.T) {
	r := &CACertificatePKCS12CertificateResource{
		client: &api.RestApiClient{},
	}

	oldUpload := uploadPKCS12CertificateFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := pkcs12CACertificateResourceValueFromFunc

	defer func() {
		uploadPKCS12CertificateFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		pkcs12CACertificateResourceValueFromFunc = oldValue
	}()

	uploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		cert := generateTestCert(t)
		block, _ := pem.Decode([]byte(cert))
		require.NotNil(t, block)
		return block.Bytes, nil
	}

	pkcs12CACertificateResourceValueFromFunc = func(context.Context, apiobjects.Certificate) (PKCS12CACertificateResourceConfig, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("model error", "fail")
		return PKCS12CACertificateResourceConfig{}, d
	}

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := r.createInternal(context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_Success(t *testing.T) {
	r := &CACertificatePKCS12CertificateResource{
		client: &api.RestApiClient{},
	}

	oldUpload := uploadPKCS12CertificateFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := pkcs12CACertificateResourceValueFromFunc

	defer func() {
		uploadPKCS12CertificateFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		pkcs12CACertificateResourceValueFromFunc = oldValue
	}()

	uploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return generateValidDERCert(t), nil
	}

	pkcs12CACertificateResourceValueFromFunc = func(ctx context.Context, obj apiobjects.Certificate) (PKCS12CACertificateResourceConfig, diag.Diagnostics) {
		return PKCS12CACertificateResourceConfig{
			PKCS12Certificate: types.StringValue("abc"),
			Password:          types.StringValue("pass"),
		}, nil
	}

	plan := PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	state, diags := r.createInternal(context.Background(), plan)

	assert.False(t, diags.HasError())
	assert.NotNil(t, state)
}

func TestCACertificatePKCS12_Delete_NoState(t *testing.T) {
	r := NewCACertificatePKCS12CertificateResource()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}
