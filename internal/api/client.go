package api

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// maxErrorBodyBytes caps how many bytes of a non-2xx response body are read.
const maxErrorBodyBytes = 1 << 20 // 1 MiB

// credentialPattern scrubs JSON fields that commonly carry secrets.
var credentialPattern = regexp.MustCompile(
	`(?i)"(password|cloudPassword|cloudUser|authorizationData|user|keyPassword|passphrase)"\s*:\s*"[^"]*"`,
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

func NewRestApiClient(client *http.Client, baseURL *url.URL, username, password string, caCertBytes []byte, clientCertBytes []byte, clientCertKey []byte, skipSSLValidation bool) (*RestApiClient, diag.Diagnostics) {
	useBasicAuth, useCertAuth := isBasicAuthProvided(username, password), isCertAuthProvided(clientCertBytes, clientCertKey)

	diags := validateAuthMode(useBasicAuth, useCertAuth)
	if diags.HasError() {
		return nil, diags
	}

	// Reject plaintext HTTP when Basic Auth credentials are present — they
	// would be transmitted base64-encoded (trivially reversible) over the wire.
	if useBasicAuth && baseURL != nil && baseURL.Scheme == "http" {
		diags.AddError(
			"Insecure instance_url",
			"Basic Auth credentials must not be sent over a plain HTTP connection. Use an https:// URL or switch to certificate-based authentication.",
		)
		return nil, diags
	}

	tlsConfig, tlsDiags := buildTLSConfig(caCertBytes, clientCertBytes, clientCertKey, useCertAuth, skipSSLValidation)
	diags = append(diags, tlsDiags...)
	if diags.HasError() {
		return nil, diags
	}

	// Always work on a fresh copy of the caller's http.Client so that
	// multiple provider aliases never share (and mutate) the same transport.
	client = cloneHTTPClient(client)
	if tlsConfig != nil {
		client = withTLSClient(client, tlsConfig)
	}

	// Refuse redirects that downgrade from HTTPS to HTTP or change the
	// target host — either could hand Basic Auth to an unintended party.
	client.CheckRedirect = safeRedirectPolicy

	// An explicit timeout prevents terraform plan/apply from hanging
	// indefinitely when the SCC endpoint is slow or unresponsive.
	if client.Timeout == 0 {
		client.Timeout = 5 * time.Minute
	}

	return &RestApiClient{
		BaseURL:  baseURL,
		Client:   client,
		Username: username,
		Password: password,
	}, diags
}

// cloneHTTPClient returns a shallow copy of c (or a new client when c is nil).
// The Transport is cloned separately so that each provider instance gets its
// own independent TLS configuration.
func cloneHTTPClient(c *http.Client) *http.Client {
	if c == nil {
		return &http.Client{}
	}
	clone := *c // copy value fields (Timeout, Jar, CheckRedirect)
	if t, ok := c.Transport.(*http.Transport); ok {
		clone.Transport = t.Clone()
	} else if c.Transport != nil {
		// Unknown transport type — keep reference; withTLSClient will wrap it.
		clone.Transport = c.Transport
	}
	return &clone
}

// safeRedirectPolicy rejects any redirect that downgrades the scheme from
// https to http or changes the destination host.  SCC is a direct REST API
// so redirects are unexpected anyway; blocking them is the safest default.
func safeRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	prev := via[len(via)-1]
	if prev.URL.Scheme == "https" && req.URL.Scheme != "https" {
		return errors.New("redirect blocked: HTTPS→HTTP scheme downgrade would expose credentials")
	}
	if req.URL.Host != prev.URL.Host {
		return errors.New("redirect blocked: cross-host redirect would expose credentials")
	}
	return nil
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

func buildTLSConfig(caCert, clientCert, clientKey []byte, useCertAuth bool, skipSSLValidation bool) (*tls.Config, diag.Diagnostics) {
	// if Certificate based authentication without SSL skip, it is mandatory to provide CA Certificate
	var diags diag.Diagnostics
	if len(caCert) == 0 && useCertAuth && !skipSSLValidation {
		diags.AddError("Missing CA Certificate", "When using certificate-based authentication, a CA certificate must be provided.")
		return nil, diags
	}

	// No TLS customization needed — return nil so the caller's transport is
	// preserved unchanged.  Default Go TLS behaviour (system CAs, verify = true)
	// is identical to &tls.Config{InsecureSkipVerify: false}, so there is no
	// behavioural difference in production; but returning nil avoids overwriting
	// the RootCAs pool carried by a test-server client.
	if !skipSSLValidation && len(caCert) == 0 && !useCertAuth {
		return nil, diags
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipSSLValidation, //nolint:gosec // controlled by explicit user opt-in
	}

	if !skipSSLValidation && len(caCert) > 0 {
		caCertPool := x509.NewCertPool()

		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			diags.AddError(
				"Invalid CA Certificate",
				"The provided CA certificate is not valid PEM-encoded data.",
			)
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

	// Path joining preserves any base path on instance_url so that proxy deployments
	// with a URL prefix (e.g. "https://host/xp264-scc") work correctly.
	// Slashes are trimmed on both sides before joining to prevent double-slashes
	// (e.g. base="/xp264-scc/" + endpoint="/api/resource" → "/xp264-scc/api/resource").
	baseURL := *c.BaseURL
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/" + strings.TrimLeft(endpointURL.Path, "/")
	baseURL.RawQuery = endpointURL.RawQuery

	req, err := http.NewRequest(method, baseURL.String(), bytes.NewBuffer(body))
	if err != nil {
		diags.AddError("Failed to Create Request", fmt.Sprintf("Error creating request for %s %s: %v", method, baseURL.String(), err))
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
		diags.AddError("Request Failed", fmt.Sprintf("Error sending %s request to %s: %v", method, baseURL.String(), err))
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

	// Cap the body read to prevent memory exhaustion from a hostile
	// or misconfigured server returning a huge response.
	limited := io.LimitReader(response.Body, maxErrorBodyBytes)
	bodyBytes, _ := io.ReadAll(limited)
	if err := response.Body.Close(); err != nil {
		diags.AddWarning("Response Body Close Failed", fmt.Sprintf("Failed to close response body: %v", err))
	}
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Handle 401 Unauthorized explicitly
	if response.StatusCode == http.StatusUnauthorized {
		diags.AddError("Authentication Failed",
			fmt.Sprintf("Authentication rejected: HTTP %d for %s %s.",
				response.StatusCode, response.Request.Method, response.Request.URL))
		return diags
	}

	// Attempt to decode a structured error message.
	// Prefer the server's structured message field; never echo the
	// raw body which may contain reflected credentials.
	var errorResp ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Message != "" {
		diags.AddError("API Error", fmt.Sprintf("HTTP %s %s failed with status %d: %s",
			response.Request.Method, response.Request.URL, response.StatusCode,
			redactCredentials(errorResp.Message)))
		return diags
	}

	// Fallback: scrub the raw body before emitting it.
	safeBody := redactCredentials(string(bodyBytes))
	diags.AddError("API Error", fmt.Sprintf("HTTP %s %s failed with status %d. Response: %s",
		response.Request.Method, response.Request.URL, response.StatusCode, safeBody))

	return diags
}

// redactCredentials replaces the values of known credential JSON fields with
// "[REDACTED]" so that secrets are never stored in Terraform diagnostics or CI
// logs even when a hostile server reflects them back in an error body.
func redactCredentials(s string) string {
	return credentialPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Keep the field name, replace only the value.
		colonIdx := strings.Index(match, ":")
		if colonIdx < 0 {
			return match
		}
		return match[:colonIdx+1] + `"[REDACTED]"`
	})
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

func (c *RestApiClient) PatchRequest(endpoint string, body []byte) (*http.Response, diag.Diagnostics) {
	return c.DoRequest(http.MethodPatch, endpoint, body, "", "")
}
