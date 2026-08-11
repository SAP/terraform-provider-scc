package actions_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func TestCreateBackupAction_Metadata(t *testing.T) {
	a := actions.NewCreateBackupAction().(*actions.CreateBackupAction)

	req := action.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &action.MetadataResponse{}

	a.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_create_backup", resp.TypeName)
}

func TestCreateBackupAction_Configure_Success(t *testing.T) {
	a := actions.NewCreateBackupAction().(*actions.CreateBackupAction)

	req := action.ConfigureRequest{
		ProviderData: &api.RestApiClient{},
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.NotNil(t, a.Client)
}

func TestCreateBackupAction_Configure_InvalidType(t *testing.T) {
	a := actions.NewCreateBackupAction().(*actions.CreateBackupAction)

	req := action.ConfigureRequest{
		ProviderData: "wrong-type",
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCreateBackupAction_Invoke_Success(t *testing.T) {
	a := &actions.CreateBackupAction{
		Client: &api.RestApiClient{},
	}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	zipBytes := []byte("PK\x03\x04test zip content")

	helpers.SendRequestFunc = func(
		client *api.RestApiClient,
		body map[string]any,
		endpoint string,
		actionType string,
	) (*http.Response, diag.Diagnostics) {
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/zip"},
			},
			Body: io.NopCloser(strings.NewReader(string(zipBytes))),
		}, nil
	}

	resp := newTestResp()

	a.InvokeWithPlan(context.Background(), testBackupPlan(), resp)

	assert.False(t, resp.Diagnostics.HasError())

	files, err := filepath.Glob("scc_backup_*.zip")
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	for _, file := range files {
		_ = os.Remove(file)
	}
}

func TestCreateBackupAction_Invoke_MissingPassword(t *testing.T) {
	a := &actions.CreateBackupAction{
		Client: &api.RestApiClient{},
	}

	resp := newTestResp()

	plan := testBackupPlan()
	plan.Password = types.StringValue("")

	a.InvokeWithPlan(context.Background(), plan, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCreateBackupAction_Invoke_NilResponseBody(t *testing.T) {
	a := &actions.CreateBackupAction{
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
			StatusCode: 200,
		}, nil
	}

	resp := newTestResp()

	a.InvokeWithPlan(context.Background(), testBackupPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCreateBackupAction_Invoke_EmptyBackup(t *testing.T) {
	a := &actions.CreateBackupAction{
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
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/zip"},
			},
			Body: io.NopCloser(strings.NewReader("")),
		}, nil
	}

	resp := newTestResp()

	a.InvokeWithPlan(context.Background(), testBackupPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCreateBackupAction_Invoke_InvalidContentType(t *testing.T) {
	a := &actions.CreateBackupAction{
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
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"error":"test"}`)),
		}, nil
	}

	resp := newTestResp()

	a.InvokeWithPlan(context.Background(), testBackupPlan(), resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func testBackupPlan() model.BackupActionConfig {
	return model.BackupActionConfig{
		Password: types.StringValue("test-password"),
	}
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func TestCreateBackupAction_Schema(t *testing.T) {
	a := actions.NewCreateBackupAction().(*actions.CreateBackupAction)
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Schema.Attributes, "password")
}

// ---------------------------------------------------------------------------
// Configure — nil provider data
// ---------------------------------------------------------------------------

func TestCreateBackupAction_Configure_NilProviderData(t *testing.T) {
	a := actions.NewCreateBackupAction().(*actions.CreateBackupAction)
	resp := &action.ConfigureResponse{}
	a.Configure(context.Background(), action.ConfigureRequest{ProviderData: nil}, resp)
	assert.Nil(t, a.Client)
	assert.False(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Invoke — top-level (Config.Get → InvokeWithPlan)
// ---------------------------------------------------------------------------

func TestCreateBackupAction_Invoke_TopLevel(t *testing.T) {
	a := actions.NewCreateBackupAction().(*actions.CreateBackupAction)
	a.Client = &api.RestApiClient{}

	oldSend := helpers.SendRequestFunc
	defer func() { helpers.SendRequestFunc = oldSend }()

	helpers.SendRequestFunc = func(client *api.RestApiClient, body map[string]any, endpoint string, actionType string) (*http.Response, diag.Diagnostics) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/zip"}},
			Body:       io.NopCloser(strings.NewReader("PK\x03\x04zip-content")),
		}, nil
	}

	schemaResp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, schemaResp)

	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{"password": tftypes.String}},
		map[string]tftypes.Value{"password": tftypes.NewValue(tftypes.String, "mypassword")},
	)
	req := action.InvokeRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := newTestResp()
	a.Invoke(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())

	files, _ := filepath.Glob("scc_backup_*.zip")
	for _, f := range files {
		_ = os.Remove(f)
	}
}

func TestCreateBackupAction_Invoke_ConfigGetError(t *testing.T) {
	a := actions.NewCreateBackupAction().(*actions.CreateBackupAction)
	a.Client = &api.RestApiClient{}

	schemaResp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, schemaResp)

	raw := tftypes.NewValue(tftypes.String, "wrong-type")
	req := action.InvokeRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := newTestResp()
	a.Invoke(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
