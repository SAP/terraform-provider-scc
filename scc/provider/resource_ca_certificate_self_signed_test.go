package provider

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCACertificateSelfSigned_Metadata(t *testing.T) {
	r := NewCACertificateSelfSignedResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_ca_certificate_self_signed", resp.TypeName)
}

func TestCACertificateSelfSigned_Schema(t *testing.T) {
	r := NewCACertificateSelfSignedResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.Contains(t, resp.Schema.Attributes, "key_size")
	assert.Contains(t, resp.Schema.Attributes, "subject_dn")
	assert.Contains(t, resp.Schema.Attributes, "certificate_pem")
}

func TestCACertificateSelfSigned_Configure_Success(t *testing.T) {
	r := NewCACertificateSelfSignedResource().(*CACertificateSelfSignedResource)

	client := &api.RestApiClient{}
	req := resource.ConfigureRequest{ProviderData: client}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, r.client)
}

func TestCACertificateSelfSigned_Configure_WrongType(t *testing.T) {
	r := NewCACertificateSelfSignedResource().(*CACertificateSelfSignedResource)

	req := resource.ConfigureRequest{ProviderData: "wrong"}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCACertificateSelfSigned_Create_MissingSubjectDN(t *testing.T) {
	r := &CACertificateSelfSignedResource{}

	plan := SelfSignedCACertificateResourceConfig{
		CertificateWithSANConfig: CertificateWithSANConfig{
			CertificateConfig: CertificateConfig{
				SubjectDN: types.ObjectNull(subjectDNAttrTypes.AttrTypes),
			},
		},
	}

	_, diags := createSelfSignedCACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificateSelfSigned_Create_RequestFails(t *testing.T) {
	r := &CACertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	defer func() { requestAndUnmarshalFunc = oldReq }()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("request failed", "fail")
		return d
	}

	plan := testValidSelfSignedCAPlan()

	_, diags := createSelfSignedCACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificateSelfSigned_Create_BinaryFails(t *testing.T) {
	r := &CACertificateSelfSignedResource{
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

	plan := testValidSelfSignedCAPlan()

	_, diags := createSelfSignedCACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificateSelfSigned_Create_Success(t *testing.T) {
	r := &CACertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := selfSignedCACertificateResourceValueFromFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		selfSignedCACertificateResourceValueFromFunc = oldValue
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return generateValidDERCert(t), nil
	}

	selfSignedCACertificateResourceValueFromFunc = func(
		ctx context.Context,
		obj apiobjects.Certificate,
		dn *certificateSubjectDNConfig,
	) (SelfSignedCACertificateResourceConfig, diag.Diagnostics) {
		return testValidSelfSignedCAPlan(), nil
	}

	plan := testValidSelfSignedCAPlan()

	state, diags := createSelfSignedCACertificateFunc(r, context.Background(), plan)

	assert.False(t, diags.HasError())
	assert.NotNil(t, state)
}

func TestCACertificateSelfSigned_Create_WithSANs(t *testing.T) {
	r := &CACertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := selfSignedCACertificateResourceValueFromFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		selfSignedCACertificateResourceValueFromFunc = oldValue
	}()

	var capturedBody map[string]any

	requestAndUnmarshalFunc = func(
		_ *api.RestApiClient,
		_ *apiobjects.Certificate,
		method string,
		_ string,
		body map[string]any,
		_ bool,
	) diag.Diagnostics {
		if method == "POST" {
			capturedBody = body
		}
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return generateValidDERCert(t), nil
	}

	selfSignedCACertificateResourceValueFromFunc = func(
		ctx context.Context,
		obj apiobjects.Certificate,
		dn *certificateSubjectDNConfig,
	) (SelfSignedCACertificateResourceConfig, diag.Diagnostics) {
		return testValidSelfSignedCAPlan(), nil
	}

	plan := testValidSelfSignedCAPlan()

	plan.SubjectAltNames = types.ListValueMust(
		subjectAlternativeNamesType,
		[]attr.Value{
			types.ObjectValueMust(
				subjectAlternativeNamesType.AttrTypes,
				map[string]attr.Value{
					"type":  types.StringValue("DNS"),
					"value": types.StringValue("example.com"),
				},
			),
		},
	)

	_, diags := createSelfSignedCACertificateFunc(r, context.Background(), plan)

	require.False(t, diags.HasError())

	assert.Contains(t, capturedBody, "subjectAltNames")
}

func TestCACertificateSelfSigned_Create_InvalidPEM(t *testing.T) {
	r := &CACertificateSelfSignedResource{
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

	plan := testValidSelfSignedCAPlan()

	_, diags := createSelfSignedCACertificateFunc(r, context.Background(), plan)

	assert.False(t, diags.HasError())
}

func TestCACertificateSelfSigned_Create_ModelFails(t *testing.T) {
	r := &CACertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := selfSignedCACertificateResourceValueFromFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		selfSignedCACertificateResourceValueFromFunc = oldValue
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return generateValidDERCert(t), nil
	}

	selfSignedCACertificateResourceValueFromFunc = func(
		ctx context.Context,
		obj apiobjects.Certificate,
		dn *certificateSubjectDNConfig,
	) (SelfSignedCACertificateResourceConfig, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("model error", "fail")
		return SelfSignedCACertificateResourceConfig{}, d
	}

	plan := testValidSelfSignedCAPlan()

	_, diags := createSelfSignedCACertificateFunc(r, context.Background(), plan)

	assert.True(t, diags.HasError())
}

func TestCACertificateSelfSigned_Delete_Success(t *testing.T) {
	r := &CACertificateSelfSignedResource{
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

func TestCACertificateSelfSigned_Read_Success(t *testing.T) {
	r := &CACertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValue := selfSignedCACertificateResourceValueFromFunc

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		selfSignedCACertificateResourceValueFromFunc = oldValue
	}()

	requestAndUnmarshalFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(*api.RestApiClient, string) ([]byte, diag.Diagnostics) {
		return generateValidDERCert(t), nil
	}

	selfSignedCACertificateResourceValueFromFunc = func(
		ctx context.Context,
		obj apiobjects.Certificate,
		dn *certificateSubjectDNConfig,
	) (SelfSignedCACertificateResourceConfig, diag.Diagnostics) {
		return testValidSelfSignedCAPlan(), nil
	}

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCACertificateSelfSigned_Update_NoChange(t *testing.T) {
	plan := testValidSelfSignedCAPlan()
	state := testValidSelfSignedCAPlan()

	noChange := plan.KeySize == state.KeySize &&
		plan.SubjectDN.Equal(state.SubjectDN) &&
		plan.SubjectAltNames.Equal(state.SubjectAltNames)

	assert.True(t, noChange)
}

func TestCACertificateSelfSigned_Update_Change(t *testing.T) {
	plan := testValidSelfSignedCAPlan()
	state := testValidSelfSignedCAPlan()

	plan.KeySize = types.Int64Value(2048)

	changed := plan.KeySize != state.KeySize ||
		!plan.SubjectDN.Equal(state.SubjectDN) ||
		!plan.SubjectAltNames.Equal(state.SubjectAltNames)

	assert.True(t, changed)
}

func testValidSelfSignedCAPlan() SelfSignedCACertificateResourceConfig {
	return SelfSignedCACertificateResourceConfig{
		KeySize: types.Int64Value(4096),

		CertificateWithSANConfig: CertificateWithSANConfig{
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
			SubjectAltNames: types.ListNull(subjectAlternativeNamesType),
		},
	}
}
