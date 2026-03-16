package provider

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestUICertificateSelfSigned_Metadata(t *testing.T) {
	r := NewUICertificateSelfSignedResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_ui_certificate_self_signed", resp.TypeName)
}

func TestUICertificateSelfSigned_Schema(t *testing.T) {
	r := NewUICertificateSelfSignedResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema.Attributes["key_size"])
	assert.NotNil(t, resp.Schema.Attributes["subject_dn"])
}

func TestUICertificateSelfSigned_Configure_Success(t *testing.T) {
	r := NewUICertificateSelfSignedResource().(*UICertificateSelfSignedResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}

	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Configure_WrongType(t *testing.T) {
	r := NewUICertificateSelfSignedResource().(*UICertificateSelfSignedResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}

	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Update(t *testing.T) {
	r := NewUICertificateSelfSignedResource()

	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Delete(t *testing.T) {
	r := &UICertificateSelfSignedResource{}

	resp := &resource.DeleteResponse{}
	req := resource.DeleteRequest{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Read(t *testing.T) {
	r := &UICertificateSelfSignedResource{
		client: &api.RestApiClient{},
	}

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

	defer func() {
		requestAndUnmarshalFunc = old
	}()

	resp := &resource.ReadResponse{}
	req := resource.ReadRequest{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUICertificateSelfSigned_Read_APIFailure(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSelfSignedResource().(*UICertificateSelfSignedResource)
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

func TestUICertificateSelfSigned_Create_MissingSubjectDN(t *testing.T) {
	ctx := context.Background()

	r := NewUICertificateSelfSignedResource().(*UICertificateSelfSignedResource)
	r.client = &api.RestApiClient{}

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

	r := NewUICertificateSelfSignedResource().(*UICertificateSelfSignedResource)
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

	r := NewUICertificateSelfSignedResource().(*UICertificateSelfSignedResource)
	r.client = &api.RestApiClient{}

	old := requestAndUnmarshalFunc
	oldModel := selfSignedUICertificateResourceValueFromFunc

	defer func() {
		requestAndUnmarshalFunc = old
		selfSignedUICertificateResourceValueFromFunc = oldModel
	}()

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
		respObj.SubjectDN = "CN=test"

		return nil
	}

	selfSignedUICertificateResourceValueFromFunc = SelfSignedUICertificateResourceValueFrom

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

func buildSelfSignedPlan(ctx context.Context, r *UICertificateSelfSignedResource) tfsdk.Plan {

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
