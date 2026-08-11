package helpers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// sendRequest / RequestAndUnmarshal helpers
// ---------------------------------------------------------------------------

func newJSONServer(t *testing.T, method string, status int, payload any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, method, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != nil {
			b, _ := json.Marshal(payload)
			_, _ = w.Write(b)
		}
	}))
}

func TestRequestAndUnmarshal_GET(t *testing.T) {
	t.Parallel()
	want := apiobjects.Certificate{Issuer: "TestCA", SerialNumber: "01"}
	srv := newJSONServer(t, http.MethodGet, http.StatusOK, want)
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "GET", "/", nil, true)

	require.False(t, diags.HasError())
	assert.Equal(t, "TestCA", got.Issuer)
	assert.Equal(t, "01", got.SerialNumber)
}

func TestRequestAndUnmarshal_POST(t *testing.T) {
	t.Parallel()
	want := apiobjects.Certificate{Issuer: "PostCA"}
	srv := newJSONServer(t, http.MethodPost, http.StatusOK, want)
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "POST", "/", map[string]any{"key": "val"}, true)

	require.False(t, diags.HasError())
	assert.Equal(t, "PostCA", got.Issuer)
}

func TestRequestAndUnmarshal_PUT(t *testing.T) {
	t.Parallel()
	want := apiobjects.Certificate{Issuer: "PutCA"}
	srv := newJSONServer(t, http.MethodPut, http.StatusOK, want)
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "PUT", "/", map[string]any{"k": "v"}, true)

	require.False(t, diags.HasError())
	assert.Equal(t, "PutCA", got.Issuer)
}

func TestRequestAndUnmarshal_DELETE(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodDelete, http.StatusNoContent, nil)
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "DELETE", "/", nil, false)

	assert.False(t, diags.HasError())
}

func TestRequestAndUnmarshal_PATCH(t *testing.T) {
	t.Parallel()
	want := apiobjects.Certificate{Issuer: "PatchCA"}
	srv := newJSONServer(t, http.MethodPatch, http.StatusOK, want)
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "PATCH", "/", map[string]any{"k": "v"}, true)

	require.False(t, diags.HasError())
	assert.Equal(t, "PatchCA", got.Issuer)
}

func TestRequestAndUnmarshal_InvalidRequestType(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodGet, http.StatusOK, nil)
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "INVALID", "/", nil, false)

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Invalid Request Type")
}

func TestRequestAndUnmarshal_WithoutMarshal(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodGet, http.StatusOK, map[string]string{"ignored": "yes"})
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "GET", "/", nil, false)

	assert.False(t, diags.HasError())
	assert.Empty(t, got.Issuer) // body not unmarshalled
}

func TestRequestAndUnmarshal_BadJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "GET", "/", nil, true)

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Failed to Unmarshal Response Body")
}

func TestRequestAndUnmarshal_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	var got apiobjects.Certificate
	diags := helpers.RequestAndUnmarshal(client, &got, "GET", "/", nil, true)

	assert.True(t, diags.HasError())
}

// sendRequest — default/unknown action branch
func TestSendRequestFunc_UnknownAction(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodGet, http.StatusOK, nil)
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)
	resp, diags := helpers.SendRequestFunc(client, nil, "/", "UNKNOWN_ACTION")

	assert.Nil(t, resp)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Invalid Action")
}

// sendRequest — each HTTP verb dispatched correctly
func TestSendRequestFunc_GET(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodGet, http.StatusOK, nil)
	defer srv.Close()

	resp, diags := helpers.SendRequestFunc(tfutils.NewTestClient(t, srv), nil, "/", helpers.ActionGetRequest)
	require.False(t, diags.HasError())
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSendRequestFunc_POST(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodPost, http.StatusCreated, nil)
	defer srv.Close()

	resp, diags := helpers.SendRequestFunc(tfutils.NewTestClient(t, srv), map[string]any{"a": 1}, "/", helpers.ActionCreateRequest)
	require.False(t, diags.HasError())
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestSendRequestFunc_PUT(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodPut, http.StatusOK, nil)
	defer srv.Close()

	resp, diags := helpers.SendRequestFunc(tfutils.NewTestClient(t, srv), map[string]any{"a": 1}, "/", helpers.ActionUpdateRequest)
	require.False(t, diags.HasError())
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSendRequestFunc_DELETE(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodDelete, http.StatusNoContent, nil)
	defer srv.Close()

	resp, diags := helpers.SendRequestFunc(tfutils.NewTestClient(t, srv), nil, "/", helpers.ActionDeleteRequest)
	require.False(t, diags.HasError())
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestSendRequestFunc_PATCH(t *testing.T) {
	t.Parallel()
	srv := newJSONServer(t, http.MethodPatch, http.StatusOK, nil)
	defer srv.Close()

	resp, diags := helpers.SendRequestFunc(tfutils.NewTestClient(t, srv), map[string]any{"a": 1}, "/", helpers.ActionPatchRequest)
	require.False(t, diags.HasError())
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ConvertMillisToTimes
// ---------------------------------------------------------------------------

func TestConvertMillisToTimes_Int64Valid(t *testing.T) {
	t.Parallel()
	result := helpers.ConvertMillisToTimes(int64(1700000000000))

	assert.False(t, result.UTC.IsNull())
	assert.False(t, result.WithTimezone.IsNull())
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, result.UTC.ValueString())
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [+-]\d{4}$`, result.WithTimezone.ValueString())
}

func TestConvertMillisToTimes_Int64Zero(t *testing.T) {
	t.Parallel()
	result := helpers.ConvertMillisToTimes(int64(0))

	assert.True(t, result.UTC.IsNull())
	assert.True(t, result.WithTimezone.IsNull())
}

func TestConvertMillisToTimes_StringValid(t *testing.T) {
	t.Parallel()
	result := helpers.ConvertMillisToTimes("1700000000000")

	assert.False(t, result.UTC.IsNull())
	assert.False(t, result.WithTimezone.IsNull())
}

func TestConvertMillisToTimes_StringInvalid(t *testing.T) {
	t.Parallel()
	result := helpers.ConvertMillisToTimes("not-a-number")

	assert.True(t, result.UTC.IsNull())
	assert.True(t, result.WithTimezone.IsNull())
}

func TestConvertMillisToTimes_StringZero(t *testing.T) {
	t.Parallel()
	result := helpers.ConvertMillisToTimes("0")

	assert.True(t, result.UTC.IsNull())
	assert.True(t, result.WithTimezone.IsNull())
}

func TestConvertMillisToTimes_UnknownType(t *testing.T) {
	t.Parallel()
	result := helpers.ConvertMillisToTimes(3.14)

	assert.True(t, result.UTC.IsNull())
	assert.True(t, result.WithTimezone.IsNull())
}

// ---------------------------------------------------------------------------
// SafeProgress
// ---------------------------------------------------------------------------

func TestSafeProgress_NilResponse(t *testing.T) {
	t.Parallel()
	// Should not panic
	helpers.SafeProgress(nil, "test message")
}

func TestSafeProgress_WithSendProgress(t *testing.T) {
	t.Parallel()
	var received string
	resp := &action.InvokeResponse{
		SendProgress: func(e action.InvokeProgressEvent) {
			received = e.Message
		},
	}

	helpers.SafeProgress(resp, "hello")
	assert.Equal(t, "hello", received)
}

func TestSafeProgress_NilSendProgress_NoPanic(t *testing.T) {
	t.Parallel()
	// SendProgress is nil — SafeProgress must recover from the resulting panic
	resp := &action.InvokeResponse{SendProgress: nil}

	assert.NotPanics(t, func() {
		helpers.SafeProgress(resp, "msg")
	})
}

// ---------------------------------------------------------------------------
// StringValueOrNull
// ---------------------------------------------------------------------------

func TestStringValueOrNull_EmptyString(t *testing.T) {
	t.Parallel()
	result := helpers.StringValueOrNull("")

	assert.Equal(t, types.StringNull(), result)
}

func TestStringValueOrNull_NonEmptyString(t *testing.T) {
	t.Parallel()
	result := helpers.StringValueOrNull("hello")

	assert.Equal(t, types.StringValue("hello"), result)
}
