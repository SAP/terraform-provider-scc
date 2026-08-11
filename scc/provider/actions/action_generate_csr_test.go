package actions_test

import (
	"context"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/scc/provider/actions"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func newTestResp() *action.InvokeResponse {
	resp := &action.InvokeResponse{}
	resp.Diagnostics = diag.Diagnostics{}
	return resp
}

func TestGenerateCSRAction_Metadata(t *testing.T) {
	a := actions.NewGenerateCSRAction().(*actions.GenerateCSRAction)

	req := action.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &action.MetadataResponse{}

	a.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_generate_csr", resp.TypeName)
}

func TestGenerateCSRAction_Configure_Success(t *testing.T) {
	a := actions.NewGenerateCSRAction().(*actions.GenerateCSRAction)

	req := action.ConfigureRequest{
		ProviderData: &api.RestApiClient{},
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.NotNil(t, a.Client)
}

func TestGenerateCSRAction_Configure_InvalidType(t *testing.T) {
	a := actions.NewGenerateCSRAction().(*actions.GenerateCSRAction)

	req := action.ConfigureRequest{
		ProviderData: "wrong-type",
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGenerateCSRAction_Invoke_Success(t *testing.T) {

	a := &actions.GenerateCSRAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
		csr := "-----BEGIN CERTIFICATE REQUEST-----\nTEST\n-----END CERTIFICATE REQUEST-----"

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(csr)),
		}, nil
	}

	resp := newTestResp()

	a.InvokeWithPlan(context.Background(), testCSRPlan(), resp)

	assert.False(t, resp.Diagnostics.HasError())

	_, err := os.Stat("system_csr.pem")
	assert.NoError(t, err)

	if err := os.Remove("system_csr.pem"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove test file: %v", err)
	}
}
func TestGenerateCSRAction_Invoke_EmptyCSR(t *testing.T) {
	a := &actions.GenerateCSRAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
		return &http.Response{
			Body: io.NopCloser(strings.NewReader("")),
		}, nil
	}

	resp := &action.InvokeResponse{}

	a.InvokeWithPlan(context.Background(), testCSRPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGenerateCSRAction_Invoke_InvalidType(t *testing.T) {
	a := &actions.GenerateCSRAction{
		Client: &api.RestApiClient{},
	}

	resp := &action.InvokeResponse{}

	plan := testCSRPlan()
	plan.Type = types.StringValue("invalid")

	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGenerateCSRAction_Invoke_WithSANs(t *testing.T) {
	a := &actions.GenerateCSRAction{
		Client: &api.RestApiClient{},
	}

	var capturedBody map[string]any

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
		capturedBody = body

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("-----BEGIN CERTIFICATE REQUEST-----\nTEST\n-----END CERTIFICATE REQUEST-----")),
		}, nil
	}

	plan := testCSRPlan()

	plan.SubjectAlternativeNames = types.ListValueMust(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":  types.StringType,
				"value": types.StringType,
			},
		},
		[]attr.Value{
			types.ObjectValueMust(
				map[string]attr.Type{
					"type":  types.StringType,
					"value": types.StringType,
				},
				map[string]attr.Value{
					"type":  types.StringValue("DNS"),
					"value": types.StringValue("example.com"),
				},
			),
		},
	)

	resp := newTestResp()

	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.False(t, resp.Diagnostics.HasError())

	sans, ok := capturedBody["subjectAltNames"]
	assert.True(t, ok, "subjectAltNames key should exist")

	sanList, ok := sans.([]map[string]string)
	assert.True(t, ok, "subjectAltNames should be []map[string]string")

	assert.Len(t, sanList, 1)
	assert.Equal(t, "DNS", sanList[0]["type"])
	assert.Equal(t, "example.com", sanList[0]["value"])

	_ = os.Remove("system_csr.pem")
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func TestGenerateCSRAction_Schema(t *testing.T) {
	a := actions.NewGenerateCSRAction().(*actions.GenerateCSRAction)
	resp := &action.SchemaResponse{}

	a.Schema(context.Background(), action.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotEmpty(t, resp.Schema.Attributes)
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "key_size")
	assert.Contains(t, resp.Schema.Attributes, "subject_dn")
	assert.Contains(t, resp.Schema.Attributes, "subject_alternative_names")
}

// ---------------------------------------------------------------------------
// Configure — nil provider data
// ---------------------------------------------------------------------------

func TestGenerateCSRAction_Configure_NilProviderData(t *testing.T) {
	a := actions.NewGenerateCSRAction().(*actions.GenerateCSRAction)

	a.Configure(context.Background(), action.ConfigureRequest{ProviderData: nil}, &action.ConfigureResponse{})

	assert.Nil(t, a.Client)
}

// ---------------------------------------------------------------------------
// InvokeWithPlan — remaining branches
// ---------------------------------------------------------------------------

func TestGenerateCSRAction_Invoke_NullSubjectDN(t *testing.T) {
	a := &actions.GenerateCSRAction{Client: &api.RestApiClient{}}

	plan := testCSRPlan()
	plan.SubjectDN = types.ObjectNull(plan.SubjectDN.AttributeTypes(context.Background()))

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Missing Subject DN")
}

func TestGenerateCSRAction_Invoke_CACertType(t *testing.T) {
	a := &actions.GenerateCSRAction{Client: &api.RestApiClient{}}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	var capturedEndpoint string
	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		capturedEndpoint = endpoint
		csr := "-----BEGIN CERTIFICATE REQUEST-----\nTEST\n-----END CERTIFICATE REQUEST-----"
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(csr))}, nil
	}

	plan := testCSRPlan()
	plan.Type = types.StringValue("ca")

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotEmpty(t, capturedEndpoint)
	_ = os.Remove("ca_csr.pem")
}

func TestGenerateCSRAction_Invoke_UICertType(t *testing.T) {
	a := &actions.GenerateCSRAction{Client: &api.RestApiClient{}}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	var capturedEndpoint string
	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		capturedEndpoint = endpoint
		csr := "-----BEGIN CERTIFICATE REQUEST-----\nTEST\n-----END CERTIFICATE REQUEST-----"
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(csr))}, nil
	}

	plan := testCSRPlan()
	plan.Type = types.StringValue("ui")

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotEmpty(t, capturedEndpoint)
	_ = os.Remove("ui_csr.pem")
}

func TestGenerateCSRAction_Invoke_SendRequestError(t *testing.T) {
	a := &actions.GenerateCSRAction{Client: &api.RestApiClient{}}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("API Error", "connection refused")
		return nil, d
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), testCSRPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGenerateCSRAction_Invoke_NilResponseBody(t *testing.T) {
	a := &actions.GenerateCSRAction{Client: &api.RestApiClient{}}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		return &http.Response{StatusCode: 200, Body: nil}, nil
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), testCSRPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid API Response")
}

func TestGenerateCSRAction_Invoke_NilResponse(t *testing.T) {
	a := &actions.GenerateCSRAction{Client: &api.RestApiClient{}}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		return nil, nil
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), testCSRPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid API Response")
}

func TestGenerateCSRAction_Invoke_NullSANs(t *testing.T) {
	a := &actions.GenerateCSRAction{Client: &api.RestApiClient{}}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	var capturedBody map[string]any
	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		capturedBody = body
		csr := "-----BEGIN CERTIFICATE REQUEST-----\nTEST\n-----END CERTIFICATE REQUEST-----"
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(csr))}, nil
	}

	plan := testCSRPlan()
	plan.SubjectAlternativeNames = types.ListNull(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":  types.StringType,
			"value": types.StringType,
		},
	})

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.False(t, resp.Diagnostics.HasError())
	_, hasSANs := capturedBody["subjectAltNames"]
	assert.False(t, hasSANs, "subjectAltNames should not be set when SANs list is null")
	_ = os.Remove("system_csr.pem")
}

// ---------------------------------------------------------------------------
// Invoke — top-level entry point (exercises req.Config.Get → InvokeWithPlan path)
// ---------------------------------------------------------------------------

func TestGenerateCSRAction_Invoke_TopLevel(t *testing.T) {
	a := actions.NewGenerateCSRAction().(*actions.GenerateCSRAction)
	a.Client = &api.RestApiClient{}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		csr := "-----BEGIN CERTIFICATE REQUEST-----\nTEST\n-----END CERTIFICATE REQUEST-----"
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(csr))}, nil
	}

	schemaResp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, schemaResp)

	subjectDNType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"cn": tftypes.String, "email": tftypes.String, "l": tftypes.String,
			"ou": tftypes.String, "o": tftypes.String, "st": tftypes.String, "c": tftypes.String,
		},
	}
	sanType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{"type": tftypes.String, "value": tftypes.String},
	}
	raw := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type":                      tftypes.String,
			"key_size":                  tftypes.Number,
			"subject_dn":                subjectDNType,
			"subject_alternative_names": tftypes.List{ElementType: sanType},
		},
	}, map[string]tftypes.Value{
		"type":     tftypes.NewValue(tftypes.String, "system"),
		"key_size": tftypes.NewValue(tftypes.Number, big.NewFloat(2048)),
		"subject_dn": tftypes.NewValue(subjectDNType, map[string]tftypes.Value{
			"cn":    tftypes.NewValue(tftypes.String, "example.com"),
			"email": tftypes.NewValue(tftypes.String, nil),
			"l":     tftypes.NewValue(tftypes.String, nil),
			"ou":    tftypes.NewValue(tftypes.String, nil),
			"o":     tftypes.NewValue(tftypes.String, "SAP"),
			"st":    tftypes.NewValue(tftypes.String, nil),
			"c":     tftypes.NewValue(tftypes.String, "IN"),
		}),
		"subject_alternative_names": tftypes.NewValue(tftypes.List{ElementType: sanType}, []tftypes.Value{}),
	})

	req := action.InvokeRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw},
	}
	resp := newTestResp()
	a.Invoke(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	_ = os.Remove("system_csr.pem")
}

func TestGenerateCSRAction_Invoke_ConfigGetError(t *testing.T) {
	a := actions.NewGenerateCSRAction().(*actions.GenerateCSRAction)
	a.Client = &api.RestApiClient{}

	schemaResp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, schemaResp)

	// Pass a raw value whose type doesn't match the schema — Config.Get will error
	raw := tftypes.NewValue(tftypes.String, "wrong-type")
	req := action.InvokeRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw},
	}
	resp := newTestResp()
	a.Invoke(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func testCSRPlan() model.CSRActionConfig {
	return model.CSRActionConfig{
		Type:    types.StringValue("system"),
		KeySize: types.Int64Value(2048),

		SubjectDN: types.ObjectValueMust(
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
				"cn":    types.StringValue("example.com"),
				"email": types.StringNull(),
				"l":     types.StringNull(),
				"ou":    types.StringNull(),
				"o":     types.StringValue("SAP"),
				"st":    types.StringNull(),
				"c":     types.StringValue("IN"),
			},
		),
		SubjectAlternativeNames: types.ListValueMust(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type":  types.StringType,
					"value": types.StringType,
				},
			},
			[]attr.Value{},
		),
	}
}

// ---------------------------------------------------------------------------
// registry.All
// ---------------------------------------------------------------------------

func TestRegistry_All(t *testing.T) {
	all := actions.All()
	assert.Len(t, all, 3)

	ctx := context.Background()
	names := make([]string, 0, len(all))
	for _, fn := range all {
		resp := &action.MetadataResponse{}
		fn().Metadata(ctx, action.MetadataRequest{ProviderTypeName: "scc"}, resp)
		names = append(names, resp.TypeName)
	}

	assert.Contains(t, names, "scc_generate_csr")
	assert.Contains(t, names, "scc_create_backup")
	assert.Contains(t, names, "scc_change_trust_store")
}
