package resources_test

import (
	"context"
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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemCertificateSignedChain_Metadata(t *testing.T) {
	r := resources.NewSystemCertificateSignedChainResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_system_certificate_signed_chain", resp.TypeName)
}

func TestSystemCertificateSignedChain_Schema(t *testing.T) {
	r := resources.NewSystemCertificateSignedChainResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema.Attributes["signed_chain"])
	assert.True(t, resp.Schema.Attributes["signed_chain"].IsRequired())
}

func TestSystemCertificateSignedChain_Schema_Attributes(t *testing.T) {
	r := resources.NewSystemCertificateSignedChainResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["signed_chain"]

	assert.NotNil(t, attr)
	assert.True(t, attr.IsRequired())
}

func TestSystemCertificateSignedChain_Configure_Success(t *testing.T) {
	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Configure_WrongType(t *testing.T) {
	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
func TestSystemCertificateSignedChain_Create_InvalidPEM(t *testing.T) {
	// Invalid PEM should fail validation
	diags := helpers.ValidatePEMChainFunc("not a cert")
	if diags == nil || !diags.HasError() {
		t.Fatalf("expected validation error for invalid PEM")
	}
}

func TestSystemCertificateSignedChain_Update_NoChange(t *testing.T) {
	ctx := context.Background()

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)

	plan := tfutils.BuildSignedChainPlan(ctx, r, "same-cert", false)
	state := tfutils.BuildSignedChainState(ctx, r, "same-cert")

	req := resource.UpdateRequest{
		Plan:  plan,
		State: state,
	}

	resp := &resource.UpdateResponse{}

	r.Update(ctx, req, resp)

	for _, d := range resp.Diagnostics {
		t.Logf("Diagnostic: %s - %s", d.Summary(), d.Detail())
	}

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Delete(t *testing.T) {
	r := &resources.SystemCertificateSignedChainResource{
		Client: &api.RestApiClient{},
	}

	// Mock API call
	old := helpers.RequestAndUnmarshalCertificateFunc
	helpers.RequestAndUnmarshalCertificateFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}
	defer func() { helpers.RequestAndUnmarshalCertificateFunc = old }()

	resp := &resource.DeleteResponse{}
	req := resource.DeleteRequest{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Delete_NoState(t *testing.T) {
	r := resources.NewSystemCertificateSignedChainResource()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Read(t *testing.T) {
	r := &resources.SystemCertificateSignedChainResource{
		Client: &api.RestApiClient{},
	}

	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc

	helpers.RequestAndUnmarshalCertificateFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}

	helpers.GetCertificateBinaryFunc = func(
		client *api.RestApiClient,
		endpoint string,
	) ([]byte, diag.Diagnostics) {
		return []byte("fakecert"), nil
	}

	defer func() {
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
	}()

	resp := &resource.ReadResponse{}
	req := resource.ReadRequest{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Read_NoState(t *testing.T) {
	r := resources.NewSystemCertificateSignedChainResource()

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestValidatePEMChain_Valid(t *testing.T) {
	cert := tfutils.GenerateTestCert(t)

	diags := helpers.ValidatePEMChainFunc(cert)

	assert.False(t, diags.HasError())
}

func TestSystemCertificateSignedChain_Delete_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	r.Client = &api.RestApiClient{}

	old := helpers.RequestAndUnmarshalCertificateFunc
	defer func() { helpers.RequestAndUnmarshalCertificateFunc = old }()

	helpers.RequestAndUnmarshalCertificateFunc = func(
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

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	r.Client = &api.RestApiClient{}

	old := helpers.RequestAndUnmarshalCertificateFunc
	defer func() { helpers.RequestAndUnmarshalCertificateFunc = old }()

	helpers.RequestAndUnmarshalCertificateFunc = func(
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

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	r.Client = &api.RestApiClient{}

	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc

	defer func() {
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
	}()

	helpers.RequestAndUnmarshalCertificateFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}

	helpers.GetCertificateBinaryFunc = func(
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

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	r.Client = &api.RestApiClient{}

	oldUpload := helpers.UploadSignedChainFunc
	defer func() { helpers.UploadSignedChainFunc = oldUpload }()

	helpers.UploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
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

func buildSystemCertificateSignedChainPlan(ctx context.Context, r *resources.SystemCertificateSignedChainResource, chain string) tfsdk.Plan {
	return tfutils.BuildSignedChainPlan(ctx, r, chain, false)
}

func TestSystemCertificateSignedChain_Create_MetadataFails(t *testing.T) {
	ctx := context.Background()

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	r.Client = &api.RestApiClient{}

	oldUpload := helpers.UploadSignedChainFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc

	defer func() {
		helpers.UploadSignedChainFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
	}()

	helpers.UploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(
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
		Plan: buildSystemCertificateSignedChainPlan(ctx, r, "fake-cert"),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Create_BinaryFails(t *testing.T) {
	ctx := context.Background()

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	r.Client = &api.RestApiClient{}

	oldUpload := helpers.UploadSignedChainFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc

	defer func() {
		helpers.UploadSignedChainFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
	}()

	helpers.UploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.Certificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}

	helpers.GetCertificateBinaryFunc = func(client *api.RestApiClient, endpoint string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	req := resource.CreateRequest{
		Plan: buildSystemCertificateSignedChainPlan(ctx, r, "fake-cert"),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Create_Success(t *testing.T) {
	ctx := context.Background()

	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	r.Client = &api.RestApiClient{}

	oldUpload := helpers.UploadSignedChainFunc
	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldBin := helpers.GetCertificateBinaryFunc
	oldModel := model.SignedChainSystemCertificateResourceValueFromFunc
	oldValidate := helpers.ValidatePEMChainFunc

	defer func() {
		helpers.UploadSignedChainFunc = oldUpload
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		helpers.GetCertificateBinaryFunc = oldBin
		model.SignedChainSystemCertificateResourceValueFromFunc = oldModel
		helpers.ValidatePEMChainFunc = oldValidate
	}()

	helpers.UploadSignedChainFunc = func(client *api.RestApiClient, endpoint, chain string) diag.Diagnostics {
		return nil
	}

	helpers.RequestAndUnmarshalCertificateFunc = func(
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

	helpers.GetCertificateBinaryFunc = func(client *api.RestApiClient, endpoint string) ([]byte, diag.Diagnostics) {
		cert := tfutils.GenerateTestCert(t)
		block, _ := pem.Decode([]byte(cert))
		require.NotNil(t, block)
		return block.Bytes, nil
	}

	// model.SignedChainSystemCertificateResourceValueFromFunc = model.SignedChainSystemCertificateResourceValueFromFunc

	helpers.ValidatePEMChainFunc = func(data string) diag.Diagnostics {
		return nil
	}

	// build schema-backed plan
	validChain := tfutils.GenerateTestCert(t)
	req := resource.CreateRequest{
		Plan: buildSystemCertificateSignedChainPlan(ctx, r, validChain),
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
