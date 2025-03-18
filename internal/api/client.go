package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type RestApiClient struct {
	Client   *http.Client
	BaseURL  *url.URL
	Username string
	Password string
}

type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewRestApiClient(client *http.Client, baseURL *url.URL, username string, password string) *RestApiClient {
	if client.Transport == nil {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}
	return &RestApiClient{
		BaseURL:  baseURL,
		Client:   client,
		Username: username,
		Password: password,
	}
}

func (c *RestApiClient) DoRequest(method string, endpoint string, body []byte) (*http.Response, error) {
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	finalURL := c.BaseURL.ResolveReference(endpointURL)

	req, err := http.NewRequest(method, finalURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	err = validateResponse(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func validateResponse(response *http.Response) error {
	switch response.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	default:
		bodyBytes, _ := io.ReadAll(response.Body)
		var errorResp ErrorResponse
		err := json.Unmarshal(bodyBytes, &errorResp)
		if err != nil {
			return fmt.Errorf("error while unmarshaling error response")
		}
		return fmt.Errorf("status: %d, error while sending the request: %s", response.StatusCode, errorResp.Message)
	}
}

func (c *RestApiClient) GetRequest(endpoint string) (*http.Response, error) {
	return c.DoRequest(http.MethodGet, endpoint, nil)
}

func (c *RestApiClient) PostRequest(endpoint string, body []byte) (*http.Response, error) {
	return c.DoRequest(http.MethodPost, endpoint, body)
}

func (c *RestApiClient) PutRequest(endpoint string, body []byte) (*http.Response, error) {
	return c.DoRequest(http.MethodPut, endpoint, body)
}

func (c *RestApiClient) DeleteRequest(endpoint string) (*http.Response, error) {
	return c.DoRequest(http.MethodDelete, endpoint, nil)
}
