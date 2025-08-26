package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	actionCreateRequest = "Create"
	actionUpdateRequest = "Update"
)

func sendGetRequest(client *api.RestApiClient, endpoint string) (*http.Response, diag.Diagnostics) {
	response, diags := client.GetRequest(endpoint)
	if diags.HasError() {
		return nil, diags
	}

	return response, diags
}

func sendPostOrPutRequest(client *api.RestApiClient, planBody map[string]any, endpoint string, action string) (*http.Response, diag.Diagnostics) {
	var response *http.Response
	var diags diag.Diagnostics
	requestByteBody, err := json.Marshal(planBody)
	if err != nil {
		diags.AddError("Failed to Marshal Request Body", fmt.Sprintf("failed to marshal API request body from plan: %v", err))
		return nil, diags
	}

	switch action {
	case actionCreateRequest:
		response, diags = client.PostRequest(endpoint, requestByteBody)
	case actionUpdateRequest:
		response, diags = client.PutRequest(endpoint, requestByteBody)
	default:
		diags.AddError("Invalid Action", fmt.Sprintf("unsupported action type: %s", action))
		return nil, diags
	}

	if diags.HasError() {
		return nil, diags
	}

	return response, diags
}

func sendDeleteRequest(client *api.RestApiClient, endpoint string) (*http.Response, diag.Diagnostics) {
	response, diags := client.DeleteRequest(endpoint)
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
		response, diags = sendGetRequest(client, endpoint)
	case "POST":
		response, diags = sendPostOrPutRequest(client, planBody, endpoint, "Create")
	case "PUT":
		response, diags = sendPostOrPutRequest(client, planBody, endpoint, "Update")
	case "DELETE":
		response, diags = sendDeleteRequest(client, endpoint)
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
