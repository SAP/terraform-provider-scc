package actions_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/scc/provider/actions"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestChangeTrustStoreAction_Metadata(t *testing.T) {
	a := actions.NewChangeTrustStoreAction().(*actions.ChangeTrustStoreAction)

	req := action.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &action.MetadataResponse{}

	a.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_change_trust_store", resp.TypeName)
}

func TestChangeTrustStoreAction_Configure_Success(t *testing.T) {
	a := actions.NewChangeTrustStoreAction().(*actions.ChangeTrustStoreAction)

	req := action.ConfigureRequest{
		ProviderData: &api.RestApiClient{},
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.NotNil(t, a.Client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Configure_NilProviderData(t *testing.T) {
	a := actions.NewChangeTrustStoreAction().(*actions.ChangeTrustStoreAction)

	req := action.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.Nil(t, a.Client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Configure_InvalidType(t *testing.T) {
	a := actions.NewChangeTrustStoreAction().(*actions.ChangeTrustStoreAction)

	req := action.ConfigureRequest{
		ProviderData: "wrong-type",
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Invoke_Success_TrustAll(t *testing.T) {
	a := &actions.ChangeTrustStoreAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	var capturedBody map[string]any

	helpers.SendRequestFunc = func(
		client *api.RestApiClient,
		body map[string]any,
		endpoint string,
		actionType string,
	) (*http.Response, diag.Diagnostics) {
		capturedBody = body
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), trustStorePlan(true), resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, true, capturedBody["trustAllBackends"])
}

func TestChangeTrustStoreAction_Invoke_Success_TrustSpecific(t *testing.T) {
	a := &actions.ChangeTrustStoreAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	var capturedBody map[string]any

	helpers.SendRequestFunc = func(
		client *api.RestApiClient,
		body map[string]any,
		endpoint string,
		actionType string,
	) (*http.Response, diag.Diagnostics) {
		capturedBody = body
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), trustStorePlan(false), resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, false, capturedBody["trustAllBackends"])
}

func TestChangeTrustStoreAction_Invoke_NullTrustAllBackends(t *testing.T) {
	a := &actions.ChangeTrustStoreAction{
		Client: &api.RestApiClient{},
	}

	resp := newTestResp()

	plan := model.BackendTrustStoreActionConfig{
		TrustAllBackends: types.BoolNull(),
	}

	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Invoke_UnknownTrustAllBackends(t *testing.T) {
	a := &actions.ChangeTrustStoreAction{
		Client: &api.RestApiClient{},
	}

	resp := newTestResp()

	plan := model.BackendTrustStoreActionConfig{
		TrustAllBackends: types.BoolUnknown(),
	}

	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Invoke_SendRequestError(t *testing.T) {
	a := &actions.ChangeTrustStoreAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(
		client *api.RestApiClient,
		body map[string]any,
		endpoint string,
		actionType string,
	) (*http.Response, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("Request Failed", "simulated network error")
		return nil, d
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), trustStorePlan(true), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Invoke_UnexpectedStatusCode(t *testing.T) {
	a := &actions.ChangeTrustStoreAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(
		client *api.RestApiClient,
		body map[string]any,
		endpoint string,
		actionType string,
	) (*http.Response, diag.Diagnostics) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"error":"internal server error"}`)),
		}, nil
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), trustStorePlan(true), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Invoke_UsesCorrectActionType(t *testing.T) {
	a := &actions.ChangeTrustStoreAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	var capturedActionType string

	helpers.SendRequestFunc = func(
		client *api.RestApiClient,
		body map[string]any,
		endpoint string,
		actionType string,
	) (*http.Response, diag.Diagnostics) {
		capturedActionType = actionType
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	resp := newTestResp()
	a.InvokeWithPlan(context.Background(), trustStorePlan(true), resp)

	assert.Equal(t, helpers.ActionPatchRequest, capturedActionType)
}

func trustStorePlan(trustAll bool) model.BackendTrustStoreActionConfig {
	return model.BackendTrustStoreActionConfig{
		TrustAllBackends: types.BoolValue(trustAll),
	}
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func TestChangeTrustStoreAction_Schema(t *testing.T) {
	a := actions.NewChangeTrustStoreAction().(*actions.ChangeTrustStoreAction)
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Schema.Attributes, "trust_all_backends")
}

// ---------------------------------------------------------------------------
// Invoke — top-level (Config.Get → InvokeWithPlan)
// ---------------------------------------------------------------------------

func TestChangeTrustStoreAction_Invoke_TopLevel(t *testing.T) {
	a := actions.NewChangeTrustStoreAction().(*actions.ChangeTrustStoreAction)
	a.Client = &api.RestApiClient{}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	schemaResp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, schemaResp)

	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{"trust_all_backends": tftypes.Bool}},
		map[string]tftypes.Value{
			"trust_all_backends": tftypes.NewValue(tftypes.Bool, true),
		},
	)
	req := action.InvokeRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := newTestResp()
	a.Invoke(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestChangeTrustStoreAction_Invoke_ConfigGetError(t *testing.T) {
	a := actions.NewChangeTrustStoreAction().(*actions.ChangeTrustStoreAction)
	a.Client = &api.RestApiClient{}

	schemaResp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, schemaResp)

	// Wrong type — Config.Get will fail
	raw := tftypes.NewValue(tftypes.String, "wrong")
	req := action.InvokeRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := newTestResp()
	a.Invoke(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
