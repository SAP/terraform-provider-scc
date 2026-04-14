package provider

import (
	"context"
	"encoding/pem"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUICertificateSignedChain_Metadata(t *testing.T) {
	r := NewUICertificateSignedChainResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_ui_certificate_signed_chain", resp.TypeName)
}

func TestUICertificateSignedChain_Schema(t *testing.T) {
	r := NewUICertificateSignedChainResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema.Attributes["signed_chain"])
	assert.True(t, resp.Schema.Attributes["signed_chain"].IsRequired())
}

func TestUICertificateSignedChain_Schema_Attributes(t *testing.T) {
	r := NewUICertificateSignedChainResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["signed_chain"]

	assert.NotNil(t, attr)
	assert.True(t, attr.IsRequired())
}

func TestUICertificateSignedChain_Configure_Success(t *testing.T) {
	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Configure_WrongType(t *testing.T) {
	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Update_NoChange(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)

	plan, state := buildUpdatePlanStateUIWithAllAttrs(ctx, r, "same-cert").Plan, buildUpdatePlanStateUIWithAllAttrs(ctx, r, "same-cert").State

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

func TestUICertificateSignedChain_Update_WithChange(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldUpload := uploadSignedChainFunc
	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc
	oldValidate := validatePEMChainFunc
	oldModel := signedChainUICertificateResourceValueFromFunc

	defer func() {
		uploadSignedChainFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
		validatePEMChainFunc = oldValidate
		signedChainUICertificateResourceValueFromFunc = oldModel
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
		certPEM := generateTestCert(t)
		block, _ := pem.Decode([]byte(certPEM))
		require.NotNil(t, block)
		return block.Bytes, nil
	}

	validatePEMChainFunc = func(data string) diag.Diagnostics {
		return nil
	}

	signedChainUICertificateResourceValueFromFunc = func(
		ctx context.Context,
		resp apiobjects.Certificate,
	) (SignedChainUICertificateResourceConfig, diag.Diagnostics) {

		subjectDNValue, _ := types.ObjectValue(
			map[string]attr.Type{
				"cn":    types.StringType,
				"email": types.StringType,
				"l":     types.StringType,
				"ou":    types.StringType,
				"o":     types.StringType,
				"st":    types.StringType,
				"c":     types.StringType,
			},
			map[string]attr.Value{
				"cn":    types.StringValue("test"),
				"email": types.StringNull(),
				"l":     types.StringNull(),
				"ou":    types.StringNull(),
				"o":     types.StringNull(),
				"st":    types.StringNull(),
				"c":     types.StringNull(),
			},
		)

		return SignedChainUICertificateResourceConfig{
			SignedChain: types.StringValue("new-cert"),

			CertificateWithSANConfig: CertificateWithSANConfig{
				CertificateConfig: CertificateConfig{
					SubjectDN:    subjectDNValue,
					Issuer:       types.StringNull(),
					SerialNumber: types.StringNull(),
					ValidFrom:    types.StringNull(),
					ValidTo:      types.StringNull(),
				},

				SubjectAltNames: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":  types.StringType,
							"value": types.StringType,
						},
					},
				),
			},
		}, nil
	}

	// schema
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	psNew := buildUpdatePlanStateUIWithAllAttrs(ctx, r, "new-cert")
	psOld := buildUpdatePlanStateUIWithAllAttrs(ctx, r, "old-cert")

	req := resource.UpdateRequest{
		Plan:  psNew.Plan,
		State: psOld.State,
	}

	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(ctx, req, resp)

	for _, d := range resp.Diagnostics {
		t.Logf("Diagnostic: %s - %s", d.Summary(), d.Detail())
	}

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Delete(t *testing.T) {
	r := &UICertificateSignedChainResource{}

	resp := &resource.DeleteResponse{}
	req := resource.DeleteRequest{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Delete_NoState(t *testing.T) {
	r := NewUICertificateSignedChainResource()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Read(t *testing.T) {
	r := &UICertificateSignedChainResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc

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

	defer func() {
		requestAndUnmarshalFunc = oldReq
	}()

	resp := &resource.ReadResponse{}
	req := resource.ReadRequest{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Read_NoState(t *testing.T) {
	r := NewUICertificateSignedChainResource()

	req := resource.ReadRequest{}
	resp := &resource.ReadResponse{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Read_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)
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

func TestUICertificateSignedChain_Create_UploadFails(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)
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

func buildUICertificateSignedChainPlan(ctx context.Context, r *UICertificateSignedChainResource, chain string) tfsdk.Plan {
	return buildSignedChainPlan(ctx, r, chain, true)
}

func TestUICertificateSignedChain_Create_MetadataFails(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)
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
		Plan: buildUICertificateSignedChainPlan(ctx, r, "fake-cert"),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUICertificateSignedChain_Create_Success(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSignedChainResource().(*UICertificateSignedChainResource)
	r.client = &api.RestApiClient{}

	oldUpload := uploadSignedChainFunc
	oldReq := requestAndUnmarshalFunc
	oldModel := signedChainUICertificateResourceValueFromFunc

	defer func() {
		uploadSignedChainFunc = oldUpload
		requestAndUnmarshalFunc = oldReq
		signedChainUICertificateResourceValueFromFunc = oldModel
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

	signedChainUICertificateResourceValueFromFunc = SignedChainUICertificateResourceValueFrom

	// build schema-backed plan
	validChain := generateTestCert(t)
	req := resource.CreateRequest{
		Plan: buildUICertificateSignedChainPlan(ctx, r, validChain),
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

func buildUpdatePlanStateUIWithAllAttrs(
	ctx context.Context,
	r *UICertificateSignedChainResource,
	value string,
) struct {
	Plan  tfsdk.Plan
	State tfsdk.State
} {
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	attrTypes := map[string]tftypes.Type{
		"signed_chain":  tftypes.String,
		"issuer":        tftypes.String,
		"serial_number": tftypes.String,
		"valid_from":    tftypes.String,
		"valid_to":      tftypes.String,
		"subject_alternative_names": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"type":  tftypes.String,
			"value": tftypes.String,
		}}},
		"subject_dn": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"cn":    tftypes.String,
			"email": tftypes.String,
			"l":     tftypes.String,
			"ou":    tftypes.String,
			"o":     tftypes.String,
			"st":    tftypes.String,
			"c":     tftypes.String,
		}},
	}

	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: attrTypes},
		map[string]tftypes.Value{
			"signed_chain": tftypes.NewValue(tftypes.String, value),

			"issuer":        tftypes.NewValue(tftypes.String, nil),
			"serial_number": tftypes.NewValue(tftypes.String, nil),
			"valid_from":    tftypes.NewValue(tftypes.String, nil),
			"valid_to":      tftypes.NewValue(tftypes.String, nil),

			"subject_alternative_names": tftypes.NewValue(
				tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"type":  tftypes.String,
					"value": tftypes.String,
				}}},
				nil,
			),

			"subject_dn": tftypes.NewValue(
				tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"cn":    tftypes.String,
					"email": tftypes.String,
					"l":     tftypes.String,
					"ou":    tftypes.String,
					"o":     tftypes.String,
					"st":    tftypes.String,
					"c":     tftypes.String,
				}},
				nil,
			),
		},
	)

	return struct {
		Plan  tfsdk.Plan
		State tfsdk.State
	}{
		Plan: tfsdk.Plan{
			Schema: schemaResp.Schema,
			Raw:    raw,
		},
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    raw,
		},
	}
}
