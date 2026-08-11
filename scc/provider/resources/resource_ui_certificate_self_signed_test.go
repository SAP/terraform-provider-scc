package resources_test

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/SAP/terraform-provider-scc/scc/provider/resources"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestUICertificateSelfSigned_Metadata(t *testing.T) {
	r := resources.NewUICertificateSelfSignedResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_ui_certificate_self_signed", resp.TypeName)
}

func TestUICertificateSelfSigned_Schema(t *testing.T) {
	r := resources.NewUICertificateSelfSignedResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema.Attributes["key_size"])
	assert.NotNil(t, resp.Schema.Attributes["subject_dn"])
}

func TestUICertificateSelfSigned_Configure_Success(t *testing.T) {
	r := resources.NewUICertificateSelfSignedResource().(*resources.UICertificateSelfSignedResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}

	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Configure_WrongType(t *testing.T) {
	r := resources.NewUICertificateSelfSignedResource().(*resources.UICertificateSelfSignedResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}

	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Update_NoChange(t *testing.T) {
	plan := testValidSelfSignedUIPlan()
	state := testValidSelfSignedUIPlan()

	noChange := plan.KeySize == state.KeySize &&
		plan.SubjectDN.Equal(state.SubjectDN) &&
		plan.SubjectAltNames.Equal(state.SubjectAltNames)

	assert.True(t, noChange)
}

func TestUICertificateSelfSigned_Update_Change(t *testing.T) {
	plan := testValidSelfSignedUIPlan()
	state := testValidSelfSignedUIPlan()

	plan.KeySize = types.Int64Value(2048)

	changed := plan.KeySize != state.KeySize ||
		!plan.SubjectDN.Equal(state.SubjectDN) ||
		!plan.SubjectAltNames.Equal(state.SubjectAltNames)

	assert.True(t, changed)
}

func TestUICertificateSelfSigned_Delete(t *testing.T) {
	r := &resources.UICertificateSelfSignedResource{}

	resp := &resource.DeleteResponse{}
	req := resource.DeleteRequest{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Read(t *testing.T) {
	r := &resources.UICertificateSelfSignedResource{
		Client: &api.RestApiClient{},
	}

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

	defer func() {
		helpers.RequestAndUnmarshalCertificateFunc = old
	}()

	resp := &resource.ReadResponse{}
	req := resource.ReadRequest{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Read_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := resources.NewUICertificateSelfSignedResource().(*resources.UICertificateSelfSignedResource)
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

func TestUICertificateSelfSigned_Create_MissingSubjectDN(t *testing.T) {
	ctx := context.Background()

	r := resources.NewUICertificateSelfSignedResource().(*resources.UICertificateSelfSignedResource)
	r.Client = &api.RestApiClient{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	plan := tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key_size": tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"key_size": tftypes.NewValue(tftypes.Number, 2048),
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

func TestUICertificateSelfSigned_Create_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := resources.NewUICertificateSelfSignedResource().(*resources.UICertificateSelfSignedResource)
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
		d.AddError("create failed", "fail")
		return d
	}

	req := resource.CreateRequest{
		Plan: buildSelfSignedPlan(ctx, r),
	}

	resp := &resource.CreateResponse{}

	r.Create(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Create_Success(t *testing.T) {
	ctx := context.Background()

	r := resources.NewUICertificateSelfSignedResource().(*resources.UICertificateSelfSignedResource)
	r.Client = &api.RestApiClient{}

	old := helpers.RequestAndUnmarshalCertificateFunc
	oldModel := model.SelfSignedUICertificateResourceValueFromFunc

	defer func() {
		helpers.RequestAndUnmarshalCertificateFunc = old
		model.SelfSignedUICertificateResourceValueFromFunc = oldModel
	}()

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
		respObj.SubjectDN = "CN=test"

		return nil
	}

	// model.SelfSignedUICertificateResourceValueFromFunc = model.SelfSignedUICertificateResourceValueFromFunc

	req := resource.CreateRequest{
		Plan: buildSelfSignedPlan(ctx, r),
	}

	resp := &resource.CreateResponse{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	resp.State = tfsdk.State{
		Schema: schemaResp.Schema,
	}

	r.Create(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func buildSelfSignedPlan(ctx context.Context, r *resources.UICertificateSelfSignedResource) tfsdk.Plan {

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

	sanType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type":  tftypes.String,
			"value": tftypes.String,
		},
	}

	attrTypes := map[string]tftypes.Type{
		"key_size":                  tftypes.Number,
		"subject_dn":                subjectDNType,
		"valid_to":                  tftypes.String,
		"valid_from":                tftypes.String,
		"issuer":                    tftypes.String,
		"serial_number":             tftypes.String,
		"subject_alternative_names": tftypes.List{ElementType: sanType},
	}

	values := map[string]tftypes.Value{
		"key_size": tftypes.NewValue(tftypes.Number, 2048),
		"subject_dn": tftypes.NewValue(
			subjectDNType,
			map[string]tftypes.Value{
				"cn":    tftypes.NewValue(tftypes.String, "test-cert"),
				"email": tftypes.NewValue(tftypes.String, nil),
				"l":     tftypes.NewValue(tftypes.String, nil),
				"ou":    tftypes.NewValue(tftypes.String, nil),
				"o":     tftypes.NewValue(tftypes.String, nil),
				"st":    tftypes.NewValue(tftypes.String, nil),
				"c":     tftypes.NewValue(tftypes.String, nil),
			},
		),
		"valid_to":                  tftypes.NewValue(tftypes.String, nil),
		"valid_from":                tftypes.NewValue(tftypes.String, nil),
		"issuer":                    tftypes.NewValue(tftypes.String, nil),
		"serial_number":             tftypes.NewValue(tftypes.String, nil),
		"subject_alternative_names": tftypes.NewValue(tftypes.List{ElementType: sanType}, nil),
	}

	return tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			values,
		),
	}
}

func testValidSelfSignedUIPlan() model.SelfSignedUICertificateResourceConfig {
	return model.SelfSignedUICertificateResourceConfig{
		KeySize: types.Int64Value(4096),

		CertificateWithSANConfig: helpers.CertificateWithSANConfig{
			CertificateConfig: helpers.CertificateConfig{
				SubjectDN: types.ObjectValueMust(
					helpers.SubjectDNAttrTypes.AttrTypes,
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
			SubjectAltNames: types.ListNull(helpers.SubjectAlternativeNamesType),
		},
	}
}

func TestUICertificateSelfSigned_Read_ModelError(t *testing.T) {
	r := &resources.UICertificateSelfSignedResource{Client: &api.RestApiClient{}}

	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldValue := model.SelfSignedUICertificateResourceValueFromFunc
	defer func() {
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		model.SelfSignedUICertificateResourceValueFromFunc = oldValue
	}()

	helpers.RequestAndUnmarshalCertificateFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}
	model.SelfSignedUICertificateResourceValueFromFunc = func(ctx context.Context, obj apiobjects.Certificate, dn *helpers.CertificateSubjectDNConfig) (model.SelfSignedUICertificateResourceConfig, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("model error", "fail")
		return model.SelfSignedUICertificateResourceConfig{}, d
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	state := testValidSelfSignedUIPlan()
	diags := tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	req := resource.ReadRequest{State: tfState}
	resp := &resource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Read_Success_WithState(t *testing.T) {
	r := &resources.UICertificateSelfSignedResource{Client: &api.RestApiClient{}}

	oldReq := helpers.RequestAndUnmarshalCertificateFunc
	oldValue := model.SelfSignedUICertificateResourceValueFromFunc
	defer func() {
		helpers.RequestAndUnmarshalCertificateFunc = oldReq
		model.SelfSignedUICertificateResourceValueFromFunc = oldValue
	}()

	helpers.RequestAndUnmarshalCertificateFunc = func(*api.RestApiClient, *apiobjects.Certificate, string, string, map[string]any, bool) diag.Diagnostics {
		return nil
	}
	model.SelfSignedUICertificateResourceValueFromFunc = func(ctx context.Context, obj apiobjects.Certificate, dn *helpers.CertificateSubjectDNConfig) (model.SelfSignedUICertificateResourceConfig, diag.Diagnostics) {
		return testValidSelfSignedUIPlan(), nil
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	state := testValidSelfSignedUIPlan()
	diags := tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	req := resource.ReadRequest{State: tfState}
	r.Read(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Delete_WithState(t *testing.T) {
	r := &resources.UICertificateSelfSignedResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	state := testValidSelfSignedUIPlan()
	diags := tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	req := resource.DeleteRequest{State: tfState}
	// resp.State must have the schema so RemoveResource doesn't dereference a nil schema
	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Delete(context.Background(), req, resp)
	// UI cert delete only adds warning, no error
	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Update_ShouldUpdateFalse(t *testing.T) {
	r := &resources.UICertificateSelfSignedResource{Client: &api.RestApiClient{}}

	oldShould := helpers.ShouldUpdateSelfSignedCertificateFunc
	defer func() { helpers.ShouldUpdateSelfSignedCertificateFunc = oldShould }()

	helpers.ShouldUpdateSelfSignedCertificateFunc = func(
		planKeySize, stateKeySize types.Int64,
		planSubjectDN, stateSubjectDN types.Object,
		planSubjectAltNames, stateSubjectAltNames types.List,
	) bool {
		return false
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	plan := testValidSelfSignedUIPlan()
	state := testValidSelfSignedUIPlan()

	tfPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	diags := tfPlan.Set(context.Background(), &plan)
	assert.False(t, diags.HasError())
	diags = tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	req := resource.UpdateRequest{Plan: tfPlan, State: tfState}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Update_CreateFails(t *testing.T) {
	r := &resources.UICertificateSelfSignedResource{Client: &api.RestApiClient{}}

	oldShould := helpers.ShouldUpdateSelfSignedCertificateFunc
	oldCreate := resources.CreateSelfSignedUICertificateFunc
	defer func() {
		helpers.ShouldUpdateSelfSignedCertificateFunc = oldShould
		resources.CreateSelfSignedUICertificateFunc = oldCreate
	}()

	helpers.ShouldUpdateSelfSignedCertificateFunc = func(
		planKeySize, stateKeySize types.Int64,
		planSubjectDN, stateSubjectDN types.Object,
		planSubjectAltNames, stateSubjectAltNames types.List,
	) bool {
		return true
	}
	resources.CreateSelfSignedUICertificateFunc = func(r *resources.UICertificateSelfSignedResource, ctx context.Context, plan model.SelfSignedUICertificateResourceConfig) (*model.SelfSignedUICertificateResourceConfig, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("create failed", "error")
		return nil, d
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	plan := testValidSelfSignedUIPlan()
	plan.KeySize = types.Int64Value(2048)
	state := testValidSelfSignedUIPlan()

	tfPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	diags := tfPlan.Set(context.Background(), &plan)
	assert.False(t, diags.HasError())
	diags = tfState.Set(context.Background(), &state)
	assert.False(t, diags.HasError())

	req := resource.UpdateRequest{Plan: tfPlan, State: tfState}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}
