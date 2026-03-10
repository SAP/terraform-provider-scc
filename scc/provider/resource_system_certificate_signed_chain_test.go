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

func TestSystemCertificateSignedChain_Metadata(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_system_certificate_signed_chain", resp.TypeName)
}

func TestSystemCertificateSignedChain_Schema(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema.Attributes["signed_chain"])
	assert.True(t, resp.Schema.Attributes["signed_chain"].IsRequired())
}

func TestSystemCertificateSignedChain_Schema_Attributes(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["signed_chain"]

	assert.NotNil(t, attr)
	assert.True(t, attr.IsRequired())
}

func TestSystemCertificateSignedChain_Configure_Success(t *testing.T) {
	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Configure_WrongType(t *testing.T) {
	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
func TestSystemCertificateSignedChain_Create_InvalidPEM(t *testing.T) {
	// Invalid PEM should fail validation
	diags := validatePEMChain("not a cert")
	if diags == nil || !diags.HasError() {
		t.Fatalf("expected validation error for invalid PEM")
	}
}

func TestSystemCertificateSignedChain_Update(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Delete(t *testing.T) {
	r := &SystemCertificateSignedChainResource{
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

func TestSystemCertificateSignedChain_Delete_NoState(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Read(t *testing.T) {
	r := &SystemCertificateSignedChainResource{
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

func TestSystemCertificateSignedChain_Read_NoState(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestValidatePEMChain_Valid(t *testing.T) {
	cert := generateTestCert(t)

	diags := validatePEMChain(cert)

	assert.False(t, diags.HasError())
}

func TestSystemCertificateSignedChain_Delete_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)
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

func TestSystemCertificateSignedChain_Read_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)
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

func TestSystemCertificateSignedChain_Read_BinaryFailure(t *testing.T) {
	ctx := context.Background()

	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)
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

func TestSystemCertificateSignedChain_Create_UploadFails(t *testing.T) {
	ctx := context.Background()

	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)
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

func buildSignedChainPlan(ctx context.Context, r *SystemCertificateSignedChainResource, chain string) tfsdk.Plan {
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

	return tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			values,
		),
	}
}

func TestSystemCertificateSignedChain_Create_MetadataFails(t *testing.T) {
	ctx := context.Background()

	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)
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
		Plan: buildSignedChainPlan(ctx, r, "fake-cert"),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Create_BinaryFails(t *testing.T) {
	ctx := context.Background()

	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)
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
		Plan: buildSignedChainPlan(ctx, r, "fake-cert"),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Create_Success(t *testing.T) {
	ctx := context.Background()

	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldUpload := uploadSignedChainFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldModel := signedChainSystemCertificateResourceValueFromFunc
	oldValidate := validatePEMChainFunc

	defer func() {
		uploadSignedChainFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		signedChainSystemCertificateResourceValueFromFunc = oldModel
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

	signedChainSystemCertificateResourceValueFromFunc = SignedChainSystemCertificateResourceValueFrom

	validatePEMChainFunc = func(data string) diag.Diagnostics {
		return nil
	}

	// build schema-backed plan
	validChain := generateTestCert(t)
	req := resource.CreateRequest{
		Plan: buildSignedChainPlan(ctx, r, validChain),
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
