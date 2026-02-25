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

	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func NewRestApiClient(client *http.Client, baseURL *url.URL, username, password string, caCertBytes []byte, clientCertBytes []byte, clientCertKey []byte) (*RestApiClient, diag.Diagnostics) {
	useBasicAuth, useCertAuth := isBasicAuthProvided(username, password), isCertAuthProvided(clientCertBytes, clientCertKey)

	diags := validateAuthMode(useBasicAuth, useCertAuth)
	if diags.HasError() {
		return nil, diags
	}

	tlsConfig, diags := buildTLSConfig(caCertBytes, clientCertBytes, clientCertKey, useCertAuth)
	if diags.HasError() {
		return nil, diags
	}

	if tlsConfig != nil {
		client = withTLSClient(client, tlsConfig)
	} else if client == nil {
		client = &http.Client{}
	}

	return &RestApiClient{
		BaseURL:  baseURL,
		Client:   client,
		Username: username,
		Password: password,
	}, diags
}

func isBasicAuthProvided(username, password string) bool {
	return username != "" && password != ""
}

func isCertAuthProvided(cert, key []byte) bool {
	return len(cert) > 0 && len(key) > 0
}

func validateAuthMode(useBasicAuth, useCertAuth bool) diag.Diagnostics {
	var diags diag.Diagnostics
	switch {
	case useBasicAuth && useCertAuth:
		// If both authentication methods are provided, return an error
		diags.AddError("Authentication Conflict", "Cannot use both certificate-based and basic authentication simultaneously. Please choose one method.")
		return diags
	case !useBasicAuth && !useCertAuth:
		// If neither authentication method is provided, return an error
		diags.AddError("Authentication Required", "Either certificate-based or basic authentication must be provided.")
		return diags
	default:
		return diags
	}
}

func buildTLSConfig(caCert, clientCert, clientKey []byte, useCertAuth bool) (*tls.Config, diag.Diagnostics) {
	// if Certificate based authentication, it is mandatory to provide CA Certificate
	var diags diag.Diagnostics
	if len(caCert) == 0 && useCertAuth {
		diags.AddError("Missing CA Certificate", "When using certificate-based authentication, a CA certificate must be provided.")
		return nil, diags
	}

	tlsConfig := &tls.Config{}

	if len(caCert) > 0 {
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			// If the CA certificate is not valid PEM-encoded data, return an error
			diags.AddError("Invalid CA Certificate", "The provided CA certificate is not valid PEM-encoded data.")
			return nil, diags
		}
		tlsConfig.RootCAs = caCertPool
	}

	if useCertAuth {
		clientCertDiags := validatePEMBlock(clientCert, "client certificate")
		diags = append(diags, clientCertDiags...)
		if diags.HasError() {
			return nil, diags
		}
		clientKeyDiags := validatePEMBlock(clientKey, "client key")
		diags = append(diags, clientKeyDiags...)
		if diags.HasError() {
			return nil, diags
		}
		// Load the client certificate and key
		cert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			diags.AddError("Invalid Client Certificate/Key", fmt.Sprintf("Failed to load client certificate/key: %v", err))
			return nil, diags
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, diags
}

func validatePEMBlock(data []byte, label string) diag.Diagnostics {
	var diags diag.Diagnostics
	if block, _ := pem.Decode(data); block == nil {
		diags.AddError("Invalid PEM Data", fmt.Sprintf("%s is not valid PEM-encoded data", label))
	}
	return diags
}

func withTLSClient(client *http.Client, tlsConfig *tls.Config) *http.Client {
	if client == nil {
		client = &http.Client{}
	}

	if client.Transport == nil {
		client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
		return client
	}

	if base, ok := client.Transport.(*http.Transport); ok {
		clone := base.Clone()
		clone.TLSClientConfig = tlsConfig
		client.Transport = clone
		return client
	}

	return client
}

func (c *RestApiClient) DoRequest(method string, endpoint string, body []byte, acceptType string, contentType string) (*http.Response, diag.Diagnostics) {
	var diags diag.Diagnostics
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		diags.AddError("Invalid Endpoint URL", fmt.Sprintf("Error parsing endpoint URL %s: %v", endpoint, err))
		return nil, diags
	}

	finalURL := c.BaseURL.ResolveReference(endpointURL)

	req, err := http.NewRequest(method, finalURL.String(), bytes.NewBuffer(body))
	if err != nil {
		diags.AddError("Failed to Create Request", fmt.Sprintf("Error creating request for %s %s: %v", method, finalURL.String(), err))
		return nil, diags
	}

	if acceptType != "" {
		req.Header.Set("Accept", acceptType)
	} else {
		req.Header.Set("Accept", "*/*")
	}

	if body != nil {
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		diags.AddError("Request Failed", fmt.Sprintf("Error sending %s request to %s: %v", method, finalURL.String(), err))
		return nil, diags
	}
	validateDiags := validateResponse(resp)
	diags = append(diags, validateDiags...)
	if diags.HasError() {
		return nil, diags
	}
	return resp, diags

}

func validateResponse(response *http.Response) diag.Diagnostics {
	var diags diag.Diagnostics
	if response.StatusCode == http.StatusOK ||
		response.StatusCode == http.StatusCreated ||
		response.StatusCode == http.StatusNoContent {
		return diags
	}

	bodyBytes, _ := io.ReadAll(response.Body)
	if err := response.Body.Close(); err != nil {
		diags.AddWarning("Response Body Close Failed", fmt.Sprintf("Failed to close response body: %v", err))
	}
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Handle 401 Unauthorized explicitly
	if response.StatusCode == http.StatusUnauthorized {
		diags.AddError("Authentication Failed",
			fmt.Sprintf("Authentication rejected: HTTP %d for %s %s. Response: %s",
				response.StatusCode, response.Request.Method, response.Request.URL, string(bodyBytes)))
		return diags
	}

	// Attempt to decode a structured error message
	var errorResp ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Message != "" {
		diags.AddError("API Error", fmt.Sprintf("HTTP %s %s failed with status %d: %s",
			response.Request.Method, response.Request.URL, response.StatusCode, errorResp.Message))
		return diags
	}

	// Fallback to raw body
	diags.AddError("API Error", fmt.Sprintf("HTTP %s %s failed with status %d. Raw response: %s",
		response.Request.Method, response.Request.URL, response.StatusCode, string(bodyBytes)))

	return diags
}

func (c *RestApiClient) GetRequest(endpoint string) (*http.Response, diag.Diagnostics) {
	return c.DoRequest(http.MethodGet, endpoint, nil, "", "")
}

func (c *RestApiClient) PostRequest(endpoint string, body []byte) (*http.Response, diag.Diagnostics) {
	return c.DoRequest(http.MethodPost, endpoint, body, "", "")
}

func (c *RestApiClient) PutRequest(endpoint string, body []byte) (*http.Response, diag.Diagnostics) {
	return c.DoRequest(http.MethodPut, endpoint, body, "", "")
}

func (c *RestApiClient) DeleteRequest(endpoint string) (*http.Response, diag.Diagnostics) {
	return c.DoRequest(http.MethodDelete, endpoint, nil, "", "")
}
