package api

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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

func NewRestApiClient(client *http.Client, baseURL *url.URL, username, password string, caCertBytes []byte, clientCertBytes []byte, clientCertKey []byte) (*RestApiClient, error) {
	useCertAuth := len(clientCertBytes) > 0 && len(clientCertKey) > 0
	useBasicAuth := username != "" && password != ""

	if useBasicAuth && useCertAuth {
		return nil, fmt.Errorf("cannot use both certificate-based and basic authentication simultaneously")
	}

	if !useBasicAuth && !useCertAuth {
		return nil, fmt.Errorf("either certificate-based or basic authentication must be provided")
	}

	tlsConfig := &tls.Config{}

	// Create TLS config only if caCertPEM is provided
	if len(caCertBytes) > 0 {
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
			return nil, fmt.Errorf("failed to parse CA certificate: input is not valid PEM-encoded data")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// If certificate auth is used, load the cert/key pair
	if useCertAuth {
		if block, _ := pem.Decode(clientCertBytes); block == nil {
			return nil, fmt.Errorf("client certificate is not valid PEM-encoded data")
		}
		if block, _ := pem.Decode(clientCertKey); block == nil {
			return nil, fmt.Errorf("client key is not valid PEM-encoded data")
		}
		cert, err := tls.X509KeyPair(clientCertBytes, clientCertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate/key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Only create a transport if we have a non-default TLS config
	if tlsConfig.RootCAs != nil || len(tlsConfig.Certificates) > 0 {
		if client == nil {
			client = &http.Client{}
		}
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	} else if client == nil {
		client = &http.Client{}
	}

	return &RestApiClient{
		BaseURL:  baseURL,
		Client:   client,
		Username: username,
		Password: password,
	}, nil
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

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

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
	if response.StatusCode == http.StatusOK ||
		response.StatusCode == http.StatusCreated ||
		response.StatusCode == http.StatusNoContent {
		return nil
	}

	bodyBytes, _ := io.ReadAll(response.Body)
	var errorResp ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err != nil || errorResp.Message == "" {
		return fmt.Errorf("HTTP %s %s failed with status %d. Raw response: %s",
			response.Request.Method, response.Request.URL, response.StatusCode, string(bodyBytes))
	}

	return fmt.Errorf("HTTP %s %s failed with status %d: %s",
		response.Request.Method, response.Request.URL, response.StatusCode, errorResp.Message)
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
