package provider

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	actionGetRequest    = "Read"
	actionCreateRequest = "Create"
	actionUpdateRequest = "Update"
	actionDeleteRequest = "Delete"
)

type FormattedTimes struct {
	UTC          types.String
	WithTimezone types.String
}

func sendRequest(client *api.RestApiClient, planBody map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
	var response *http.Response
	var diags diag.Diagnostics
	var requestByteBody []byte
	var err error

	if planBody != nil {
		requestByteBody, err = json.Marshal(planBody)
		if err != nil {
			diags.AddError("Failed to Marshal Request Body", fmt.Sprintf("failed to marshal API request body from plan: %v", err))
			return nil, diags
		}
	}

	switch action {
	case actionGetRequest:
		response, diags = client.GetRequest(endpoint)
	case actionCreateRequest:
		response, diags = client.PostRequest(endpoint, requestByteBody)
	case actionUpdateRequest:
		response, diags = client.PutRequest(endpoint, requestByteBody)
	case actionDeleteRequest:
		response, diags = client.DeleteRequest(endpoint)
	default:
		diags.AddError("Invalid Action", fmt.Sprintf("unsupported action type: %s", action))
		return nil, diags
	}

	if diags.HasError() {
		return nil, diags
	}

	return response, diags
}

func requestAndUnmarshal[T any](client *api.RestApiClient, respObj *T, requestType string, endpoint string, planBody map[string]any, marshalResponse bool) diag.Diagnostics {
	var response *http.Response
	var diags diag.Diagnostics
	switch requestType {
	case "GET":
		response, diags = sendRequest(client, nil, endpoint, actionGetRequest)
	case "POST":
		response, diags = sendRequest(client, planBody, endpoint, actionCreateRequest)
	case "PUT":
		response, diags = sendRequest(client, planBody, endpoint, actionUpdateRequest)
	case "DELETE":
		response, diags = sendRequest(client, nil, endpoint, actionDeleteRequest)
	default:
		diags.AddError("Invalid Request Type", fmt.Sprintf("unsupported request type: %s", requestType))
		return diags
	}

	if diags.HasError() {
		return diags
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			diags.AddError("Failed to Close Response Body", fmt.Sprintf("failed to close response body: %v", err))
		}
	}()

	if marshalResponse {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			diags.AddError("Failed to Read Response Body", fmt.Sprintf("failed to read API response body: %v", err))
			return diags
		}

		err = json.Unmarshal(body, respObj)
		if err != nil {
			diags.AddError("Failed to Unmarshal Response Body", fmt.Sprintf("failed to unmarshal API response body: %v", err))
			return diags
		}
	}

	return diags

}

func ConvertMillisToTimes(millis any) FormattedTimes {
	var ms int64
	switch v := millis.(type) {
	case int64:
		ms = v
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return FormattedTimes{types.StringNull(), types.StringNull()}
		}
		ms = parsed
	default:
		return FormattedTimes{types.StringNull(), types.StringNull()}
	}

	if ms == 0 {
		return FormattedTimes{types.StringNull(), types.StringNull()}
	}

	// Build time
	t := time.UnixMilli(ms).UTC()

	return FormattedTimes{
		UTC:          types.StringValue(t.Format("2006-01-02 15:04:05")),
		WithTimezone: types.StringValue(t.Format("2006-01-02 15:04:05 -0700")),
	}
}

func GetCertificateBinary(client *api.RestApiClient, endpoint string) ([]byte, diag.Diagnostics) {
	response, diags := client.DoRequest(http.MethodGet, endpoint, nil, "application/pkix-cert")
	if diags.HasError() {
		return nil, diags
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			diags.AddWarning(
				"Failed to Close Response Body",
				fmt.Sprintf("error closing response body: %v", err),
			)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Failed to Read Response Body", fmt.Sprintf("failed to read response body: %v", err))
		return nil, diags
	}
	return body, diags
}

func validatePEMData(data string) diag.Diagnostics {
	var diags diag.Diagnostics

	if strings.TrimSpace(data) == "" {
		diags.AddError(
			"Empty PEM Data",
			"No certificate data provided.",
		)
		return diags
	}

	block, _ := pem.Decode([]byte(data))
	if block == nil {
		diags.AddError(
			"Invalid PEM Block",
			"Failed to decode PEM block. Ensure certificate is valid PEM format.",
		)
		return diags
	}

	// Only check supported types
	switch block.Type {
	case "CERTIFICATE",
		"PRIVATE KEY",
		"RSA PRIVATE KEY",
		"EC PRIVATE KEY":
		return diags
	default:
		diags.AddError(
			"Unsupported PEM Type",
			fmt.Sprintf("Unsupported PEM block type: %s", block.Type),
		)
	}

	return diags
}
