package provider

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func newTestResp() *action.InvokeResponse {
	resp := &action.InvokeResponse{}
	resp.Diagnostics = diag.Diagnostics{}
	return resp
}

func TestGenerateCSRAction_Metadata(t *testing.T) {
	a := NewGenerateCSRAction()

	req := action.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &action.MetadataResponse{}

	a.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_generate_csr", resp.TypeName)
}

func TestGenerateCSRAction_Configure_Success(t *testing.T) {
	a := NewGenerateCSRAction().(*GenerateCSRAction)

	req := action.ConfigureRequest{
		ProviderData: &api.RestApiClient{},
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.NotNil(t, a.client)
}

func TestGenerateCSRAction_Configure_InvalidType(t *testing.T) {
	a := NewGenerateCSRAction().(*GenerateCSRAction)

	req := action.ConfigureRequest{
		ProviderData: "wrong-type",
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGenerateCSRAction_Invoke_Success(t *testing.T) {

	a := &GenerateCSRAction{
		client: &api.RestApiClient{},
	}

	oldSend := sendRequestFunc
	defer func() { sendRequestFunc = oldSend }()

	sendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
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
	a := &GenerateCSRAction{
		client: &api.RestApiClient{},
	}

	oldSend := sendRequestFunc
	defer func() { sendRequestFunc = oldSend }()

	sendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
		return &http.Response{
			Body: io.NopCloser(strings.NewReader("")),
		}, nil
	}

	resp := &action.InvokeResponse{}

	a.InvokeWithPlan(context.Background(), testCSRPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGenerateCSRAction_Invoke_InvalidType(t *testing.T) {
	a := &GenerateCSRAction{
		client: &api.RestApiClient{},
	}

	resp := &action.InvokeResponse{}

	plan := testCSRPlan()
	plan.Type = types.StringValue("invalid")

	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGenerateCSRAction_Invoke_WithSANs(t *testing.T) {
	a := &GenerateCSRAction{
		client: &api.RestApiClient{},
	}

	var capturedBody map[string]any

	oldSend := sendRequestFunc
	defer func() { sendRequestFunc = oldSend }()

	sendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
		capturedBody = body

		return &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader("-----BEGIN CERTIFICATE REQUEST-----\nTEST\n-----END CERTIFICATE REQUEST-----")),
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

func testCSRPlan() CSRActionConfig {
	return CSRActionConfig{
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
