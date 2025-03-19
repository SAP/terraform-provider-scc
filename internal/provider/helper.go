package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
)

func sendGetRequest(client *api.RestApiClient, endpoint string) (*http.Response, error) {
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request to %s: %v", endpoint, err)
	}

	return response, nil
}

func sendPostOrPutRequest(client *api.RestApiClient, planBody map[string]string, endpoint string, action string) (*http.Response, error) {

	requestByteBody, err := json.Marshal(planBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal API request body from plan: %v", err)
	}

	if action != "Create" {
		response, err := client.PutRequest(endpoint, requestByteBody)
		if err != nil {
			return nil, fmt.Errorf("failed to send PUT request to %s: %v", endpoint, err)
		}
		return response, nil
	}

	response, err := client.PostRequest(endpoint, requestByteBody)
	if err != nil {
		return nil, fmt.Errorf("failed to send POST request to %s: %v", endpoint, err)
	}

	return response, nil
}

func sendDeleteRequest(client *api.RestApiClient, endpoint string) (*http.Response, error) {
	response, err := client.DeleteRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to send DELETE request to %s: %v", endpoint, err)
	}

	return response, nil
}

func requestAndUnmarshal[T any](client *api.RestApiClient, respObj *T, requestType string, endpoint string, planBody map[string]string, marshalResponse bool) error {
	var response *http.Response
	var err error
	switch requestType {
	case "GET":
		response, err = sendGetRequest(client, endpoint)
	case "POST":
		response, err = sendPostOrPutRequest(client, planBody, endpoint, "Create")
	case "PUT":
		response, err = sendPostOrPutRequest(client, planBody, endpoint, "Update")
	case "DELETE":
		response, err = sendDeleteRequest(client, endpoint)
	default:
		return fmt.Errorf("invalid request type: %s", requestType)
	}

	if err != nil {
		return err
	}

	if marshalResponse {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read API response body: %v", err)
		}

		defer response.Body.Close()

		err = json.Unmarshal(body, &respObj)
		if err != nil {
			return fmt.Errorf("failed to unmarshal API response body: %v", err)
		}
	}

	return nil

}
