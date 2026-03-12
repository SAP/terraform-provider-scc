package provider

import (
	"context"
	"encoding/pem"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCACertificateSignedChain_Metadata(t *testing.T) {
	r := NewCACertificateSignedChainResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_ca_certificate_signed_chain", resp.TypeName)
}

func TestCACertificateSignedChain_Schema(t *testing.T) {
	r := NewCACertificateSignedChainResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema.Attributes["signed_chain"])
	assert.True(t, resp.Schema.Attributes["signed_chain"].IsRequired())
}

func TestCACertificateSignedChain_Schema_Attributes(t *testing.T) {
	r := NewCACertificateSignedChainResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["signed_chain"]

	assert.NotNil(t, attr)
	assert.True(t, attr.IsRequired())
}

func TestCACertificateSignedChain_Configure_Success(t *testing.T) {
	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Configure_WrongType(t *testing.T) {
	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
func TestCACertificateSignedChain_Create_InvalidPEM(t *testing.T) {
	// Invalid PEM should fail validation
	diags := validatePEMChain("not a cert")
	if diags == nil || !diags.HasError() {
		t.Fatalf("expected validation error for invalid PEM")
	}
}

func TestCACertificateSignedChain_Update(t *testing.T) {
	r := NewCACertificateSignedChainResource()

	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Delete(t *testing.T) {
	r := &CACertificateSignedChainResource{
		client: &api.RestApiClient{},
	}

	// Mock API call
	old := requestAndUnmarshalFunc
	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}
	defer func() { requestAndUnmarshalFunc = old }()

	resp := &resource.DeleteResponse{}
	req := resource.DeleteRequest{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Delete_NoState(t *testing.T) {
	r := NewCACertificateSignedChainResource()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Read(t *testing.T) {
	r := &CACertificateSignedChainResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(
		client *api.RestApiClient,
		endpoint string,
	) ([]byte, diag.Diagnostics) {
		return []byte("fakecert"), nil
	}

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	resp := &resource.ReadResponse{}
	req := resource.ReadRequest{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Read_NoState(t *testing.T) {
	r := NewCACertificateSignedChainResource()

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Delete_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	old := requestAndUnmarshalFunc
	defer func() { requestAndUnmarshalFunc = old }()

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("delete failed", "fail")
		return d
	}

	// get resource schema
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{},
			map[string]tftypes.Value{},
		),
	}

	req := resource.DeleteRequest{
		State: state,
	}

	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Read_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	old := requestAndUnmarshalFunc
	defer func() { requestAndUnmarshalFunc = old }()

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("api error", "fail")
		return d
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
	}

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{}

	r.Read(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Read_BinaryFailure(t *testing.T) {
	ctx := context.Background()

	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(
		client *api.RestApiClient,
		endpoint string,
	) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
	}

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{}

	r.Read(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Create_UploadFails(t *testing.T) {
	ctx := context.Background()

	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldUpload := uploadSignedChainFunc
	defer func() { uploadSignedChainFunc = oldUpload }()

	uploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("upload failed", "fail")
		return d
	}

	// get resource schema
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	// build valid plan
	plan := tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"signed_chain": tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"signed_chain": tftypes.NewValue(tftypes.String, "fake-cert"),
			},
		),
	}

	req := resource.CreateRequest{
		Plan: plan,
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func buildCACertificateSignedChainPlan(ctx context.Context, r *CACertificateSignedChainResource, chain string) tfsdk.Plan {
	return buildSignedChainPlan(ctx, r, chain, true)
}

func TestCACertificateSignedChain_Create_MetadataFails(t *testing.T) {
	ctx := context.Background()

	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldUpload := uploadSignedChainFunc
	oldReq := requestAndUnmarshalFunc

	defer func() {
		uploadSignedChainFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
	}()

	uploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("metadata error", "fail")
		return d
	}

	req := resource.CreateRequest{
		Plan: buildCACertificateSignedChainPlan(ctx, r, "fake-cert"),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Create_BinaryFails(t *testing.T) {
	ctx := context.Background()

	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldUpload := uploadSignedChainFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc

	defer func() {
		uploadSignedChainFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	uploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(client *api.RestApiClient, endpoint string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	req := resource.CreateRequest{
		Plan: buildCACertificateSignedChainPlan(ctx, r, "fake-cert"),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificateSignedChain_Create_Success(t *testing.T) {
	ctx := context.Background()

	r := NewCACertificateSignedChainResource().(*CACertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldUpload := uploadSignedChainFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldModel := signedChainCertificateResourceValueFromFunc
	oldValidate := validatePEMChainFunc

	defer func() {
		uploadSignedChainFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		signedChainCertificateResourceValueFromFunc = oldModel
		validatePEMChainFunc = oldValidate
	}()

	uploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
		return nil
	}

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {

		respObj.Issuer = "Test CA"
		respObj.SerialNumber = "123"
		respObj.NotBeforeTimeStamp = 1700000000000
		respObj.NotAfterTimeStamp = 1800000000000
		respObj.SubjectDN = "CN=test.example.com"

		return nil
	}

	getCertificateBinaryFunc = func(client *api.RestApiClient, endpoint string) ([]byte, diag.Diagnostics) {
		cert := generateTestCert(t)
		block, _ := pem.Decode([]byte(cert))
		require.NotNil(t, block)
		return block.Bytes, nil
	}

	signedChainCertificateResourceValueFromFunc = SignedChainCertificateResourceValueFrom

	validatePEMChainFunc = func(data string) diag.Diagnostics {
		return nil
	}

	// build schema-backed plan
	validChain := generateTestCert(t)
	req := resource.CreateRequest{
		Plan: buildCACertificateSignedChainPlan(ctx, r, validChain),
	}

	resp := &resource.CreateResponse{}

	// IMPORTANT: attach schema so State.Set() works
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	resp.State = tfsdk.State{
		Schema: schemaResp.Schema,
	}

	r.Create(ctx, req, resp)

	for _, d := range resp.Diagnostics {
		t.Logf("Diagnostic: %s - %s", d.Summary(), d.Detail())
	}

	assert.False(t, resp.Diagnostics.HasError())
}
