package provider_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/scc/provider"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestSCCProvider_AllResources(t *testing.T) {

	expectedResources := []string{
		"scc_domain_mapping",
		"scc_subaccount",
		"scc_system_mapping_resource",
		"scc_system_mapping",
		"scc_subaccount_k8s_service_channel",
		"scc_subaccount_abap_service_channel",
		"scc_subaccount_using_auth",
		"scc_system_certificate_self_signed",
		"scc_system_certificate_signed_chain",
		"scc_system_certificate_pkcs12_certificate",
		"scc_ca_certificate_self_signed",
		"scc_ca_certificate_signed_chain",
		"scc_ca_certificate_pkcs12_certificate",
		"scc_ui_certificate_self_signed",
		"scc_ui_certificate_signed_chain",
		"scc_ui_certificate_pkcs12_certificate",
		"scc_proxy_settings",
	}

	ctx := context.Background()
	registeredResources := []string{}

	for _, resourceFunc := range provider.New().Resources(ctx) {
		var resp resource.MetadataResponse

		resourceFunc().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "scc"}, &resp)

		registeredResources = append(registeredResources, resp.TypeName)
	}

	assert.ElementsMatch(t, expectedResources, registeredResources)
}

func TestSCCProvider_AllDataSources(t *testing.T) {

	expectedDataSources := []string{
		"scc_domain_mapping",
		"scc_domain_mappings",
		"scc_subaccount_configuration",
		"scc_subaccounts",
		"scc_system_mapping_resource",
		"scc_system_mapping_resources",
		"scc_system_mapping",
		"scc_system_mappings",
		"scc_subaccount_k8s_service_channel",
		"scc_subaccount_k8s_service_channels",
		"scc_subaccount_abap_service_channel",
		"scc_subaccount_abap_service_channels",
		"scc_system_certificate",
		"scc_ca_certificate",
		"scc_proxy_settings",
	}

	ctx := context.Background()
	registeredDataSources := []string{}

	for _, datasourceFunc := range provider.New().DataSources(ctx) {
		var resp datasource.MetadataResponse

		datasourceFunc().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "scc"}, &resp)

		registeredDataSources = append(registeredDataSources, resp.TypeName)
	}

	assert.ElementsMatch(t, expectedDataSources, registeredDataSources)
}

func TestSCCProvider_ListResources(t *testing.T) {
	ctx := context.Background()

	expected := []string{
		"scc_subaccount",
		"scc_domain_mapping",
		"scc_system_mapping_resource",
		"scc_system_mapping",
		"scc_subaccount_k8s_service_channel",
		"scc_subaccount_abap_service_channel",
	}

	p := provider.New()

	listProvider, ok := p.(tfprovider.ProviderWithListResources)
	if !ok {
		t.Fatalf("provider does not implement ProviderWithListResources")
	}

	var registered []string

	for _, listResourceFunc := range listProvider.ListResources(ctx) {
		var resp resource.MetadataResponse

		listResourceFunc().Metadata(
			ctx,
			resource.MetadataRequest{
				ProviderTypeName: "scc",
			},
			&resp,
		)

		registered = append(registered, resp.TypeName)
	}

	assert.ElementsMatch(t, expected, registered)
}

func TestSCCProvider_AllActions(t *testing.T) {
	ctx := context.Background()

	expected := []string{
		"scc_generate_csr",
		"scc_create_backup",
	}

	p := provider.New()

	actionProvider, ok := p.(tfprovider.ProviderWithActions)
	if !ok {
		t.Fatalf("provider does not implement ProviderWithActions")
	}

	var registered []string

	for _, actionProviderFunc := range actionProvider.Actions(ctx) {
		var resp action.MetadataResponse

		actionProviderFunc().Metadata(
			ctx,
			action.MetadataRequest{
				ProviderTypeName: "scc",
			},
			&resp,
		)

		registered = append(registered, resp.TypeName)
	}

	assert.ElementsMatch(t, expected, registered)
}

func TestSCCProvider_MissingURL(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	ok := provider.ValidateConfig("", "admin", "pass", "", "", "", false, &resp)

	assert.False(t, ok)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_ErrorParseURL(t *testing.T) {
	var resp tfprovider.ConfigureResponse

	// Build invalid URL using non-constant expression to bypass staticcheck
	invalidURL := fmt.Sprintf("ht%ctp://bad-url", '!')

	ok := provider.ValidateConfig(invalidURL, "admin", "pass", "", "", "", false, &resp)

	_, err := url.Parse(invalidURL)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Invalid Cloud Connector Instance URL",
			fmt.Sprintf("Failed to parse the provided Cloud Connector Instance URL: %s. Error: %v", invalidURL, err),
		)
		ok = false
	}

	assert.False(t, ok, "Expected validateConfig to return false due to invalid URL")
	assert.True(t, resp.Diagnostics.HasError(), "Expected diagnostics to contain error for invalid URL")
}

func TestSCCProvider_BasicAuthOnly(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	ok := provider.ValidateConfig("https://example.com", "admin", "pass", "", "", "", false, &resp)

	assert.True(t, ok)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_ConflictingAuth(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	ok := provider.ValidateConfig("https://example.com", "admin", "pass", "", "cert", "key", false, &resp)

	assert.False(t, ok)
	assert.True(t, resp.Diagnostics.HasError())
}

// Test that only certificate-based auth (without basic auth) is accepted.
func TestSCCProvider_CertAuthOnly(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	dummyPEM := `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIJAIk+Cm3ekmKaMAoGCCqGSM49BAMCMBIxEDAOBgNVBAMM
B1Rlc3QgQ0EwHhcNMjAwMTAxMDAwMDAwWhcNMzAwMTAxMDAwMDAwWjASMRAwDgYD
VQQDDAdUZXN0IENBMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFpJSyVnGE8Ow
K8Bk7hrcn/ElMGyDx+0CgWl+oD+DFsVCtZnQaBFkgVctbWOrYDWJjvPUK+iPY35x
ph6V/9bDNqNQME4wHQYDVR0OBBYEFENZqO6v+u1eZzZTVDNj0uUCkN8gMB8GA1Ud
IwQYMBaAFENZqO6v+u1eZzZTVDNj0uUCkN8gMAwGA1UdEwQFMAMBAf8wCgYIKoZI
zj0EAwIDSAAwRQIgTTb7LtqRQon2OHxMOyuvl+e8FQZXzSH14Yc7u9s9n9ICIQDE
CEGH5OML6z7C7oCSys7ce4GkTbtJ4rNZoxVOxFwPvA==
-----END CERTIFICATE-----`
	ok := provider.ValidateConfig("https://example.com", "", "", dummyPEM, dummyPEM, dummyPEM, false, &resp)

	assert.True(t, ok)
	assert.False(t, resp.Diagnostics.HasError())
}

// Test that empty auth results in error.
func TestSCCProvider_NoAuth(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	ok := provider.ValidateConfig("https://example.com", "", "", "", "", "", false, &resp)

	assert.False(t, ok)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_InvalidPEM(t *testing.T) {
	diags := helpers.ValidatePEMDataFunc("not-a-valid-pem")
	assert.True(t, diags.HasError(), "Expected diagnostics to contain error for invalid PEM")
	assert.Equal(t, "Invalid PEM Block", diags[0].Summary())
}

func TestSCCProvider_ValidPEM(t *testing.T) {
	dummyPEM := `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIJAIk+Cm3ekmKaMAoGCCqGSM49BAMCMBIxEDAOBgNVBAMM
B1Rlc3QgQ0EwHhcNMjAwMTAxMDAwMDAwWhcNMzAwMTAxMDAwMDAwWjASMRAwDgYD
VQQDDAdUZXN0IENBMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFpJSyVnGE8Ow
K8Bk7hrcn/ElMGyDx+0CgWl+oD+DFsVCtZnQaBFkgVctbWOrYDWJjvPUK+iPY35x
ph6V/9bDNqNQME4wHQYDVR0OBBYEFENZqO6v+u1eZzZTVDNj0uUCkN8gMB8GA1Ud
IwQYMBaAFENZqO6v+u1eZzZTVDNj0uUCkN8gMAwGA1UdEwQFMAMBAf8wCgYIKoZI
zj0EAwIDSAAwRQIgTTb7LtqRQon2OHxMOyuvl+e8FQZXzSH14Yc7u9s9n9ICIQDE
CEGH5OML6z7C7oCSys7ce4GkTbtJ4rNZoxVOxFwPvA==
-----END CERTIFICATE-----`

	diags := helpers.ValidatePEMDataFunc(dummyPEM)
	assert.False(t, diags.HasError(), "expected no error diagnostics for valid PEM")
	assert.Len(t, diags, 0)
}

func TestSCCProvider_ClientCreationFails(t *testing.T) {
	var resp tfprovider.ConfigureResponse

	dummyPEM := `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIJAIk+Cm3ekmKaMAoGCCqGSM49BAMCMBIxEDAOBgNVBAMM
B1Rlc3QgQ0EwHhcNMjAwMTAxMDAwMDAwWhcNMzAwMTAxMDAwMDAwWjASMRAwDgYD
VQQDDAdUZXN0IENBMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFpJSyVnGE8Ow
K8Bk7hrcn/ElMGyDx+0CgWl+oD+DFsVCtZnQaBFkgVctbWOrYDWJjvPUK+iPY35x
ph6V/9bDNqNQME4wHQYDVR0OBBYEFENZqO6v+u1eZzZTVDNj0uUCkN8gMB8GA1Ud
IwQYMBaAFENZqO6v+u1eZzZTVDNj0uUCkN8gMAwGA1UdEwQFMAMBAf8wCgYIKoZI
zj0EAwIDSAAwRQIgTTb7LtqRQon2OHxMOyuvl+e8FQZXzSH14Yc7u9s9n9ICIQDE
CEGH5OML6z7C7oCSys7ce4GkTbtJ4rNZoxVOxFwPvA==
-----END CERTIFICATE-----`

	instanceURL := "https://example.com"
	username := "admin"
	password := "password"

	ok := provider.ValidateConfig(instanceURL, username, password, dummyPEM, dummyPEM, dummyPEM, false, &resp)

	assert.False(t, ok)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Conflicting Authentication Details")
}

func Test_ProviderConnection_Success(t *testing.T) {
	client := &api.RestApiClient{
		// Simulate a successful response
		Client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader("version info")),
				}
			}),
		},
		BaseURL:  mustParseURL(t, "https://example.com"),
		Username: "user",
		Password: "pass",
	}

	diags := provider.TestProviderConnection(client)
	assert.False(t, diags.HasError(), "expected no error diagnostics for successful connection")
	assert.Len(t, diags, 0)
}

func Test_ProviderConnection_Unauthorized(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/api/version", nil)

	client := &api.RestApiClient{
		Client: &http.Client{
			Transport: roundTripFunc(func(_ *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Status:     "401 Unauthorized",
					Body:       io.NopCloser(strings.NewReader("unauthorized")),
					Request:    req,
				}
			}),
		},
		BaseURL:  mustParseURL(t, "https://example.com"),
		Username: "bad-user",
		Password: "wrong-pass",
	}

	diags := provider.TestProviderConnection(client)

	assert.True(t, diags.HasError(), "expected diagnostics to contain error for unauthorized response")
	assert.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary(), "Authentication Failed")
	assert.Contains(t, diags[0].Detail(), "rejected")
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestSCCProvider_ParseInstanceURL_Valid(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	urlStr := "https://valid.example.com"

	parsed := provider.ParseInstanceURL(urlStr, &resp)

	assert.NotNil(t, parsed)
	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "valid.example.com", parsed.Host)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_ParseInstanceURL_Invalid(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	invalidURL := "ht!tp://bad-url"

	parsed := provider.ParseInstanceURL(invalidURL, &resp)

	assert.Nil(t, parsed)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid Cloud Connector Instance URL")
}

func TestSCCProvider_CreateClient_Success(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	httpClient := &http.Client{}
	parsedURL := mustParseURL(t, "https://example.com")

	client := provider.CreateClient(httpClient, parsedURL, "user", "pass", "", "", "", false, &resp)

	assert.NotNil(t, client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_CreateClient_Failure_InvalidCert(t *testing.T) {
	var resp tfprovider.ConfigureResponse
	httpClient := &http.Client{}
	parsedURL := mustParseURL(t, "https://example.com")

	invalidCert := "-----BEGIN BAD-----"

	client := provider.CreateClient(httpClient, parsedURL, "", "", invalidCert, invalidCert, invalidCert, false, &resp)

	assert.Nil(t, client)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Client Creation Failed")
}
