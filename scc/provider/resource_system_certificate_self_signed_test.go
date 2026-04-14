package provider

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSystemCertificateSelfSigned_Metadata(t *testing.T) {
	r := NewSystemCertificateSelfSignedResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_system_certificate_self_signed", resp.TypeName)
}

func TestSystemCertificateSelfSigned_Schema(t *testing.T) {
	r := NewSystemCertificateSelfSignedResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.Contains(t, resp.Schema.Attributes, "key_size")
	assert.Contains(t, resp.Schema.Attributes, "subject_dn")
	assert.Contains(t, resp.Schema.Attributes, "certificate_pem")
}

func TestSystemCertificateSelfSigned_Configure_Success(t *testing.T) {
	r := NewSystemCertificateSelfSignedResource().(*SystemCertificateSelfSignedResource)

	client := &api.RestApiClient{}
	req := resource.ConfigureRequest{ProviderData: client}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, r.client)
}

func TestSystemCertificateSelfSigned_Configure_WrongType(t *testing.T) {
	r := NewSystemCertificateSelfSignedResource().(*SystemCertificateSelfSignedResource)

	req := resource.ConfigureRequest{ProviderData: "wrong"}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSelfSigned_Create_MissingSubjectDN(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{}

	plan := SelfSignedSystemCertificateResourceConfig{
		CertificateConfig: CertificateConfig{
			SubjectDN: types.ObjectNull(subjectDNAttrTypes.AttrTypes),
		},
	}

	_, diags := createSelfSignedSystemCertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestSystemCertificateSelfSigned_Create_RequestFails(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	defer func() { requestAndUnmarshalFunc = oldReq }()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("request failed", "fail")
		return d
	}

	plan := testValidSelfSignedSystemPlan()

	_, diags := createSelfSignedSystemCertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestSystemCertificateSelfSigned_Create_BinaryFails(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("binary error", "fail")
		return nil, d
	}

	plan := testValidSelfSignedSystemPlan()

	_, diags := createSelfSignedSystemCertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestSystemCertificateSelfSigned_Create_Success(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := selfSignedSystemCertificateResourceValueFromFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		selfSignedSystemCertificateResourceValueFromFunc = oldValue
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return generateValidDERCert(t), nil
	}

	selfSignedSystemCertificateResourceValueFromFunc = func(
		ctx context.Context,
		obj apiobjects.Certificate,
		dn *certificateSubjectDNConfig,
	) (SelfSignedSystemCertificateResourceConfig, diag.Diagnostics) {
		return testValidSelfSignedSystemPlan(), nil
	}

	plan := testValidSelfSignedSystemPlan()

	state, diags := createSelfSignedSystemCertificateFunc(r, context.Background(), plan)

	assert.False(t, diags.HasError())
	assert.NotNil(t, state)
}

func TestSystemCertificateSelfSigned_Create_InvalidPEM(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return []byte("invalid"), nil
	}

	plan := testValidSelfSignedSystemPlan()

	_, diags := createSelfSignedSystemCertificateFunc(r, context.Background(), plan)

	assert.False(t, diags.HasError())
}

func TestSystemCertificateSelfSigned_Read_RequestFails(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	defer func() { requestAndUnmarshalFunc = oldReq }()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("fail", "fail")
		return d
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	tfState := tfsdk.State{
		Schema: schemaResp.Schema,
	}

	state := testValidSelfSignedSystemPlan()

	diags := tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	req := resource.ReadRequest{
		State: tfState,
	}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSelfSigned_Read_BinaryFails(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("fail", "fail")
		return nil, d
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	tfState := tfsdk.State{
		Schema: schemaResp.Schema,
	}

	state := testValidSelfSignedSystemPlan()

	diags := tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	req := resource.ReadRequest{
		State: tfState,
	}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSelfSigned_Delete_Failure(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	defer func() { requestAndUnmarshalFunc = oldReq }()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("fail", "fail")
		return d
	}

	state := testValidSelfSignedSystemPlan()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	tfState := tfsdk.State{
		Schema: schemaResp.Schema,
	}

	diags := tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	req := resource.DeleteRequest{
		State: tfState,
	}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSelfSigned_Delete_Success(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	defer func() { requestAndUnmarshalFunc = oldReq }()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSelfSigned_Read_Success(t *testing.T) {
	r := &SystemCertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := selfSignedSystemCertificateResourceValueFromFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		selfSignedSystemCertificateResourceValueFromFunc = oldValue
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return generateValidDERCert(t), nil
	}

	selfSignedSystemCertificateResourceValueFromFunc = func(
		ctx context.Context,
		obj apiobjects.Certificate,
		dn *certificateSubjectDNConfig,
	) (SelfSignedSystemCertificateResourceConfig, diag.Diagnostics) {
		return testValidSelfSignedSystemPlan(), nil
	}

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSelfSigned_Update_NoChange(t *testing.T) {
	plan := testValidSelfSignedSystemPlan()
	state := testValidSelfSignedSystemPlan()

	changed := (plan.KeySize != state.KeySize &&
		plan.SubjectDN.Equal(state.SubjectDN))

	assert.False(t, changed)
}

func TestSystemCertificateSelfSigned_Update_Change(t *testing.T) {
	plan := testValidSelfSignedSystemPlan()
	state := testValidSelfSignedSystemPlan()

	plan.KeySize = types.Int64Value(2048)

	changed := shouldUpdateSelfSignedCertificate(plan.KeySize, state.KeySize, plan.SubjectDN, state.SubjectDN, types.ListNull(types.StringType), types.ListNull(types.StringType))

	assert.True(t, changed)
}

func testValidSelfSignedSystemPlan() SelfSignedSystemCertificateResourceConfig {
	return SelfSignedSystemCertificateResourceConfig{
		KeySize: types.Int64Value(4096),

		CertificateConfig: CertificateConfig{
			SubjectDN: types.ObjectValueMust(
				subjectDNAttrTypes.AttrTypes,
				map[string]attr.Value{
					"cn":    types.StringValue("example.com"),
					"email": types.StringNull(),
					"l":     types.StringNull(),
					"ou":    types.StringNull(),
					"o":     types.StringNull(),
					"st":    types.StringNull(),
					"c":     types.StringNull(),
				},
			),
		},
	}
}
