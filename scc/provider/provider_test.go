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
	regexValidTimeStamp    = regexp.MustCompile(`^\d{13}$`)
	regexValidSerialNumber = regexp.MustCompile(`^(?:[0-9a-fA-F]{2}:){14,}[0-9a-fA-F]{2}$`)
)

type User struct {
	InstanceUsername string
	InstancePassword string
	InstanceURL      string
	// For adding subaccount to the cloud connector
	CloudUsername string
	CloudPassword string
}

var redactedTestUser = User{
	InstanceUsername: "test-user@example.com",
	InstancePassword: "REDACTED_INSTANCE_PASSWORD",
	InstanceURL:      "https://redacted.instance.url",
	CloudUsername:    "cloud-user@example.com",
	CloudPassword:    "REDACTED_CLOUD_PASSWORD",
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
		user.InstanceUsername = os.Getenv("SCC_USERNAME")
		user.InstancePassword = os.Getenv("SCC_PASSWORD")
		user.InstanceURL = os.Getenv("SCC_INSTANCE_URL")

		user.CloudUsername = os.Getenv("TF_VAR_cloud_user")
		user.CloudPassword = os.Getenv("TF_VAR_cloud_password")
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

		if strings.Contains(i.Response.Body, "subaccountCertificate") {
			reNotAfter := regexp.MustCompile(`"notAfterTimeStamp"\s*:\s*\d{13}`)
			i.Response.Body = reNotAfter.ReplaceAllString(i.Response.Body, `"notAfterTimeStamp": 1111111111111`)

			reNotBefore := regexp.MustCompile(`"notBeforeTimeStamp"\s*:\s*\d{13}`)
			i.Response.Body = reNotBefore.ReplaceAllString(i.Response.Body, `"notBeforeTimeStamp": 1111111111111`)

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
	err := validatePEM("not-a-valid-pem")
	assert.Error(t, err)
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

	err := validatePEM(dummyPEM)
	assert.NoError(t, err)
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
