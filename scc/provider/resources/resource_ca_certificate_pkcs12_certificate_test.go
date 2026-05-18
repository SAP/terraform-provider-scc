package resources_test

import (
	"context"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/SAP/terraform-provider-scc/scc/provider/resources"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCACertificatePKCS12Certificate_Metadata(t *testing.T) {
	r := resources.NewCACertificatePKCS12CertificateResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_ca_certificate_pkcs12_certificate", resp.TypeName)
}

func TestCACertificatePKCS12Certificate_Schema_Attributes(t *testing.T) {
	r := resources.NewCACertificatePKCS12CertificateResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	passwordAttr := resp.Schema.Attributes["password"]
	assert.True(t, passwordAttr.IsRequired())

	keyAttr := resp.Schema.Attributes["key_password"]
	assert.True(t, keyAttr.IsOptional())
}

func TestCACertificatePKCS12Certificate_Configure_Success(t *testing.T) {
	r := resources.NewCACertificatePKCS12CertificateResource().(*resources.CACertificatePKCS12CertificateResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_Configure_WrongType(t *testing.T) {
	r := resources.NewCACertificatePKCS12CertificateResource().(*resources.CACertificatePKCS12CertificateResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_Read_NoState(t *testing.T) {
	r := resources.NewCACertificatePKCS12CertificateResource()

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_RawValue(t *testing.T) {
	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("raw-data"),
	}

	raw, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("raw-data"), raw)
}

func TestCACertificatePKCS12Certificate_Inputs_EmptyCert(t *testing.T) {
	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue(""),
	}

	_, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_EmptyKeyPassword(t *testing.T) {
	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		KeyPassword:       types.StringValue(""),
	}

	_, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_Base64Decode(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("hello"))

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue(encoded),
	}

	raw, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("hello"), raw)
}

func TestCACertificatePKCS12Certificate_Inputs_KeyPasswordUnknown(t *testing.T) {
	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		KeyPassword:       types.StringUnknown(),
	}

	_, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Inputs_ValidWithKeyPassword(t *testing.T) {
	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		KeyPassword:       types.StringValue("keypass"),
	}

	raw, diags := helpers.ValidatePKCS12InputsFunc(plan.PKCS12Certificate, plan.KeyPassword)

	assert.False(t, diags.HasError())
	assert.Equal(t, []byte("abc"), raw)
}

func TestCACertificatePKCS12Certificate_Create_ValidationFails(t *testing.T) {
	r := &resources.CACertificatePKCS12CertificateResource{}

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue(""),
	}

	_, diags := resources.CreatePKCS12CACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_UploadFails(t *testing.T) {
	r := &resources.CACertificatePKCS12CertificateResource{
		Client: &api.RestApiClient{},
	}

	oldUpload := helpers.UploadPKCS12CertificateFunc
	defer func() { helpers.UploadPKCS12CertificateFunc = oldUpload }()

	helpers.UploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("upload failed", "fail")
		return d
	}

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := resources.CreatePKCS12CACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_MetadataFails(t *testing.T) {
	r := &resources.CACertificatePKCS12CertificateResource{
		Client: &api.RestApiClient{},
	}

	oldUpload := helpers.UploadPKCS12CertificateFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	defer func() {
		helpers.UploadPKCS12CertificateFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
	}()

	helpers.UploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("metadata error", "fail")
		return d
	}

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := resources.CreatePKCS12CACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_BinaryFails(t *testing.T) {
	r := &resources.CACertificatePKCS12CertificateResource{
		Client: &api.RestApiClient{},
	}

	oldUpload := helpers.UploadPKCS12CertificateFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc
	defer func() {
		helpers.UploadPKCS12CertificateFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
	}()

	helpers.UploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	helpers.GetCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := resources.CreatePKCS12CACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_InvalidPEM(t *testing.T) {
	r := &resources.CACertificatePKCS12CertificateResource{
		Client: &api.RestApiClient{},
	}

	oldUpload := helpers.UploadPKCS12CertificateFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc
	oldValue := model.PKCS12CACertificateResourceValueFromFunc

	defer func() {
		helpers.UploadPKCS12CertificateFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
		model.PKCS12CACertificateResourceValueFromFunc = oldValue
	}()

	helpers.UploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	helpers.GetCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := resources.CreatePKCS12CACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_ModelConversionFails(t *testing.T) {
	r := &resources.CACertificatePKCS12CertificateResource{
		Client: &api.RestApiClient{},
	}

	oldUpload := helpers.UploadPKCS12CertificateFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc
	oldValue := model.PKCS12CACertificateResourceValueFromFunc

	defer func() {
		helpers.UploadPKCS12CertificateFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
		model.PKCS12CACertificateResourceValueFromFunc = oldValue
	}()

	helpers.UploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	helpers.GetCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		cert := tfutils.GenerateTestCert(t)
		block, _ := pem.Decode([]byte(cert))
		require.NotNil(t, block)
		return block.Bytes, nil
	}

	model.PKCS12CACertificateResourceValueFromFunc = func(context.Context, apiobjects.Certificate) (model.PKCS12CACertificateResourceConfig, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("model error", "fail")
		return model.PKCS12CACertificateResourceConfig{}, d
	}

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	_, diags := resources.CreatePKCS12CACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificatePKCS12Certificate_Create_Success(t *testing.T) {
	r := &resources.CACertificatePKCS12CertificateResource{
		Client: &api.RestApiClient{},
	}

	oldUpload := helpers.UploadPKCS12CertificateFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc
	oldValue := model.PKCS12CACertificateResourceValueFromFunc

	defer func() {
		helpers.UploadPKCS12CertificateFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
		model.PKCS12CACertificateResourceValueFromFunc = oldValue
	}()

	helpers.UploadPKCS12CertificateFunc = func(*api.RestApiClient, string, []byte, string, string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	helpers.GetCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return tfutils.GenerateValidDERCert(t), nil
	}

	model.PKCS12CACertificateResourceValueFromFunc = func(ctx context.Context, obj apiobjects.Certificate) (model.PKCS12CACertificateResourceConfig, diag.Diagnostics) {
		return model.PKCS12CACertificateResourceConfig{
			PKCS12Certificate: types.StringValue("abc"),
			Password:          types.StringValue("pass"),
		}, nil
	}

	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	state, diags := resources.CreatePKCS12CACertificateFunc(r, context.Background(), plan)

	assert.False(t, diags.HasError())
	assert.NotNil(t, state)
}

func TestCACertificatePKCS12_Delete_NoState(t *testing.T) {
	r := resources.NewCACertificatePKCS12CertificateResource()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificatePKCS12Certificate_ShouldUpdate_NoChange(t *testing.T) {
	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("abc"),
		Password:          types.StringValue("pass"),
	}

	state := plan

	result := helpers.ShouldUpdatePKCS12Func(plan.PKCS12Certificate, state.PKCS12Certificate, plan.Password, state.Password, plan.KeyPassword, state.KeyPassword)

	assert.False(t, result)
}

func TestCACertificatePKCS12Certificate_ShouldUpdate_Change(t *testing.T) {
	plan := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("new"),
		Password:          types.StringValue("pass"),
	}

	state := model.PKCS12CACertificateResourceConfig{
		PKCS12Certificate: types.StringValue("old"),
		Password:          types.StringValue("pass"),
	}

	result := helpers.ShouldUpdatePKCS12Func(plan.PKCS12Certificate, state.PKCS12Certificate, plan.Password, state.Password, plan.KeyPassword, state.KeyPassword)

	assert.True(t, result)
}
