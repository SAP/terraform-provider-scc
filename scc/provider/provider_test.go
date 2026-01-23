package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

var (
	regexpValidUUID        = uuidvalidator.UuidRegexp
	regexValidTimeStamp    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?: \+\d{4})?$`)
	regexValidSerialNumber = regexp.MustCompile(`^(?:[0-9a-fA-F]{2}:){14,}[0-9a-fA-F]{1,2}$`)
)

type User struct {
	InstanceUsername string
	InstancePassword string
	InstanceURL      string
	// For adding subaccount to the cloud connector
	CloudUsername           string
	CloudPassword           string
	CloudAuthenticationData string
	// For adding K8S service channel to subaccount
	K8SCluster          string
	K8SService          string
	ABAPCloudTenantHost string
}

var redactedTestUser = User{
	InstanceUsername:        "test-user@example.com",
	InstancePassword:        "REDACTED_INSTANCE_PASSWORD",
	InstanceURL:             "https://redacted.instance.url",
	CloudUsername:           "cloud-user@example.com",
	CloudPassword:           "REDACTED_CLOUD_PASSWORD",
	CloudAuthenticationData: "REDACTED_SUBACCOUNT_AUTHENTICATION_DATA",
	K8SCluster:              "REDACTED_K8S_CLUSTER_HOST",
	K8SService:              "REDACTED_K8S_SERVICE_ID",
	ABAPCloudTenantHost:     "REDACTED_ABAP_CLOUD_TENANT_HOST",
}

func providerConfig(testUser User) string {
	return fmt.Sprintf(`
	provider "scc" {
	instance_url= "%s"
	username= "%s"
	password= "%s"
	}
	`, testUser.InstanceURL, testUser.InstanceUsername, testUser.InstancePassword)
}

func getTestProviders(httpClient *http.Client) map[string]func() (tfprotov6.ProviderServer, error) {
	cloudconnectorProvider := NewWithClient(httpClient).(*cloudConnectorProvider)

	return map[string]func() (tfprotov6.ProviderServer, error){
		"scc": providerserver.NewProtocol6WithError(cloudconnectorProvider),
	}
}

func setupVCR(t *testing.T, cassetteName string) (*recorder.Recorder, User) {
	t.Helper()

	mode := recorder.ModeRecordOnce
	if testRecord, _ := strconv.ParseBool(os.Getenv("TEST_RECORD")); testRecord {
		mode = recorder.ModeRecordOnly
	}

	user := redactedTestUser

	rec, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassetteName,
		Mode:               mode,
		SkipRequestLatency: true,
		RealTransport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})

	if rec.IsRecording() {
		t.Logf("ATTENTION: Recording '%s'", cassetteName)
		// Get environment variables for initiating provider
		user.InstanceUsername = os.Getenv("SCC_USERNAME")
		user.InstancePassword = os.Getenv("SCC_PASSWORD")
		user.InstanceURL = os.Getenv("SCC_INSTANCE_URL")

		// Get environment variables for recording test fixtures
		user.CloudUsername = os.Getenv("TF_VAR_cloud_user")
		user.CloudPassword = os.Getenv("TF_VAR_cloud_password")
		user.CloudAuthenticationData = os.Getenv("TF_VAR_authentication_data")
		user.K8SCluster = os.Getenv("TF_VAR_k8s_cluster_host")
		user.K8SService = os.Getenv("TF_VAR_k8s_service_id")
		user.ABAPCloudTenantHost = os.Getenv("TF_VAR_abap_cloud_tenant_host")
		if len(user.InstanceUsername) == 0 || len(user.InstancePassword) == 0 || len(user.InstanceURL) == 0 {
			t.Fatal("Env vars SCC_USERNAME, SCC_PASSWORD and SCC_INSTANCE_URL are required when recording test fixtures")
		}
	} else {
		t.Logf("Replaying '%s'", cassetteName)
	}

	if err != nil {
		t.Fatal()
	}

	rec.SetMatcher(requestMatcher(t))
	rec.AddHook(hookRedactSensitiveCredentials(), recorder.BeforeSaveHook)
	rec.AddHook(hookRedactBodyLinks(), recorder.BeforeSaveHook)
	rec.AddHook(hookRedactSensitiveBody(), recorder.BeforeSaveHook)

	return rec, user
}

func requestMatcher(t *testing.T) cassette.MatcherFunc {
	return func(r *http.Request, i cassette.Request) bool {
		if r.Method != i.Method || r.URL.String() != i.URL {
			return false
		}

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal("Unable to read body from request")
		}

		r.Body = io.NopCloser(strings.NewReader(string(bytes)))
		return string(bytes) == i.Body
	}
}

func hookRedactSensitiveCredentials() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		redact := func(headers map[string][]string) {
			for key := range headers {
				if strings.Contains(strings.ToLower(key), "x-csrf-token") ||
					strings.Contains(strings.ToLower(key), "set-cookie") ||
					strings.Contains(strings.ToLower(key), "authorization") ||
					strings.Contains(strings.ToLower(key), "location") {
					headers[key] = []string{"redacted"}
				}
			}
		}

		ipOrHostRegex := regexp.MustCompile(`https://(?:[a-zA-Z0-9\-\.]+|\d{1,3}(?:\.\d{1,3}){3})(?::\d+)?`)
		i.Request.URL = ipOrHostRegex.ReplaceAllString(i.Request.URL, redactedTestUser.InstanceURL)

		hostRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9\-\.]+|\d{1,3}(?:\.\d{1,3}){3})(?::\d+)?$`)
		i.Request.Host = hostRegex.ReplaceAllString(i.Request.Host, redactedTestUser.InstanceURL)

		redact(i.Request.Headers)
		redact(i.Response.Headers)

		return nil
	}
}

func hookRedactSensitiveBody() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		if strings.Contains(i.Request.Body, "cloudPassword") {
			reBindingSecret := regexp.MustCompile(`"cloudPassword":"(.*?)"`)
			i.Request.Body = reBindingSecret.ReplaceAllString(i.Request.Body, `"cloudPassword":"`+redactedTestUser.CloudPassword+`"`)
		}

		if strings.Contains(i.Request.Body, "cloudUser") {
			reBindingSecret := regexp.MustCompile(`"cloudUser":"(.*?)"`)
			i.Request.Body = reBindingSecret.ReplaceAllString(i.Request.Body, `"cloudUser":"`+redactedTestUser.CloudUsername+`"`)
		}

		if strings.Contains(i.Request.Body, "authenticationData") {
			reBindingSecret := regexp.MustCompile(`"authenticationData":"(.*?)"`)
			i.Request.Body = reBindingSecret.ReplaceAllString(i.Request.Body, `"authenticationData":"`+redactedTestUser.CloudAuthenticationData+`"`)
		}

		if strings.Contains(i.Request.Body, "k8sCluster") {
			reBindingSecret := regexp.MustCompile(`"k8sCluster":"(.*?)"`)
			i.Request.Body = reBindingSecret.ReplaceAllString(i.Request.Body, `"k8sCluster":"`+redactedTestUser.K8SCluster+`"`)
		}

		if strings.Contains(i.Request.Body, "k8sService") {
			reBindingSecret := regexp.MustCompile(`"k8sService":"(.*?)"`)
			i.Request.Body = reBindingSecret.ReplaceAllString(i.Request.Body, `"k8sService":"`+redactedTestUser.K8SService+`"`)
		}

		if strings.Contains(i.Response.Body, "k8sCluster") {
			reBindingSecret := regexp.MustCompile(`"k8sCluster":"(.*?)"`)
			i.Response.Body = reBindingSecret.ReplaceAllString(i.Response.Body, `"k8sCluster":"`+redactedTestUser.K8SCluster+`"`)
		}

		if strings.Contains(i.Response.Body, "k8sService") {
			reBindingSecret := regexp.MustCompile(`"k8sService":"(.*?)"`)
			i.Response.Body = reBindingSecret.ReplaceAllString(i.Response.Body, `"k8sService":"`+redactedTestUser.K8SService+`"`)
		}

		if strings.Contains(i.Request.Body, "abapCloudTenantHost") {
			reBindingSecret := regexp.MustCompile(`"abapCloudTenantHost":"(.*?)"`)
			i.Request.Body = reBindingSecret.ReplaceAllString(i.Request.Body, `"abapCloudTenantHost":"`+redactedTestUser.ABAPCloudTenantHost+`"`)
		}

		if strings.Contains(i.Response.Body, "abapCloudTenantHost") {
			reBindingSecret := regexp.MustCompile(`"abapCloudTenantHost":"(.*?)"`)
			i.Response.Body = reBindingSecret.ReplaceAllString(i.Response.Body, `"abapCloudTenantHost":"`+redactedTestUser.ABAPCloudTenantHost+`"`)
		}

		if strings.Contains(i.Response.Body, "subaccountCertificate") {
			reSubjectDN := regexp.MustCompile(`"subjectDN"\s*:\s*".*?"`)
			i.Response.Body = reSubjectDN.ReplaceAllString(i.Response.Body, `"subjectDN": "CN=redacted,L=redacted,OU=redacted,OU=redacted,O=redacted,C=redacted"`)

			reIssuer := regexp.MustCompile(`"issuer"\s*:\s*".*?"`)
			i.Response.Body = reIssuer.ReplaceAllString(i.Response.Body, `"issuer": "CN=redacted,OU=SAP Cloud Platform Clients,O=redacted,L=redacted,C=redacted"`)

			reSerial := regexp.MustCompile(`"serialNumber"\s*:\s*"(?:[0-9a-fA-F]{2}:){15}[0-9a-fA-F]{2}"`)
			i.Response.Body = reSerial.ReplaceAllString(i.Response.Body, `"serialNumber": "aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa"`)
		}

		if strings.Contains(i.Response.Body, "tunnel") {
			reUser := regexp.MustCompile(`"user"\s*:\s*".*?"`)
			i.Response.Body = reUser.ReplaceAllString(i.Response.Body, `"user":"`+redactedTestUser.CloudUsername+`"`)
		}

		return nil
	}
}

func hookRedactBodyLinks() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		if strings.Contains(i.Response.Body, "_links") {
			// Redact all href URLs under _links
			reHref := regexp.MustCompile(`"href"\s*:\s*"https://[^"]+"`)
			i.Response.Body = reHref.ReplaceAllString(i.Response.Body, `"href": "https://redacted.url/path"`)
		}

		return nil
	}
}

func stopQuietly(rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		panic(err)
	}
}

func TestSCCProvider_AllResources(t *testing.T) {

	expectedResources := []string{
		"scc_domain_mapping",
		"scc_subaccount",
		"scc_system_mapping_resource",
		"scc_system_mapping",
		"scc_subaccount_k8s_service_channel",
		"scc_subaccount_abap_service_channel",
		"scc_subaccount_using_auth",
	}

	ctx := context.Background()
	registeredResources := []string{}

	for _, resourceFunc := range New().Resources(ctx) {
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
	}

	ctx := context.Background()
	registeredDataSources := []string{}

	for _, datasourceFunc := range New().DataSources(ctx) {
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
	}

	p := New()

	listProvider, ok := p.(provider.ProviderWithListResources)
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

func TestSCCProvider_MissingURL(t *testing.T) {
	var resp provider.ConfigureResponse
	ok := validateConfig("", "admin", "pass", "", "", "", &resp)

	assert.False(t, ok)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_ErrorParseURL(t *testing.T) {
	var resp provider.ConfigureResponse

	// Build invalid URL using non-constant expression to bypass staticcheck
	invalidURL := fmt.Sprintf("ht%ctp://bad-url", '!')

	ok := validateConfig(invalidURL, "admin", "pass", "", "", "", &resp)

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
	var resp provider.ConfigureResponse
	ok := validateConfig("https://example.com", "admin", "pass", "", "", "", &resp)

	assert.True(t, ok)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_ConflictingAuth(t *testing.T) {
	var resp provider.ConfigureResponse
	ok := validateConfig("https://example.com", "admin", "pass", "", "cert", "key", &resp)

	assert.False(t, ok)
	assert.True(t, resp.Diagnostics.HasError())
}

// Test that only certificate-based auth (without basic auth) is accepted.
func TestSCCProvider_CertAuthOnly(t *testing.T) {
	var resp provider.ConfigureResponse
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
	ok := validateConfig("https://example.com", "", "", dummyPEM, dummyPEM, dummyPEM, &resp)

	assert.True(t, ok)
	assert.False(t, resp.Diagnostics.HasError())
}

// Test that empty auth results in error.
func TestSCCProvider_NoAuth(t *testing.T) {
	var resp provider.ConfigureResponse
	ok := validateConfig("https://example.com", "", "", "", "", "", &resp)

	assert.False(t, ok)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_InvalidPEM(t *testing.T) {
	diags := validatePEM("not-a-valid-pem")
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

	diags := validatePEM(dummyPEM)
	assert.False(t, diags.HasError(), "expected no error diagnostics for valid PEM")
	assert.Len(t, diags, 0)
}

func TestSCCProvider_ClientCreationFails(t *testing.T) {
	var resp provider.ConfigureResponse

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

	ok := validateConfig(instanceURL, username, password, dummyPEM, dummyPEM, dummyPEM, &resp)

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

	diags := testProviderConnection(client)
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

	diags := testProviderConnection(client)

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
	var resp provider.ConfigureResponse
	urlStr := "https://valid.example.com"

	parsed := parseInstanceURL(urlStr, &resp)

	assert.NotNil(t, parsed)
	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "valid.example.com", parsed.Host)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_ParseInstanceURL_Invalid(t *testing.T) {
	var resp provider.ConfigureResponse
	invalidURL := "ht!tp://bad-url"

	parsed := parseInstanceURL(invalidURL, &resp)

	assert.Nil(t, parsed)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid Cloud Connector Instance URL")
}

func TestSCCProvider_CreateClient_Success(t *testing.T) {
	var resp provider.ConfigureResponse
	httpClient := &http.Client{}
	parsedURL := mustParseURL(t, "https://example.com")

	client := createClient(httpClient, parsedURL, "user", "pass", "", "", "", &resp)

	assert.NotNil(t, client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSCCProvider_CreateClient_Failure_InvalidCert(t *testing.T) {
	var resp provider.ConfigureResponse
	httpClient := &http.Client{}
	parsedURL := mustParseURL(t, "https://example.com")

	invalidCert := "-----BEGIN BAD-----"

	client := createClient(httpClient, parsedURL, "", "", invalidCert, invalidCert, invalidCert, &resp)

	assert.Nil(t, client)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Client Creation Failed")
}
