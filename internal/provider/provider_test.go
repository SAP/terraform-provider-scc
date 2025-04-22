package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-cloudconnector/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
	InstanceUsername: "testuser@sap.com",
	InstancePassword: "testpassword",
	InstanceURL:      "https://127.0.0.1:8443",
	CloudUsername:    "testuser@sap.com",
	CloudPassword:    "testpassword",
}

func providerConfig(testUser User) string {
	return fmt.Sprintf(`
	provider "cloudconnector" {
	instance_url= "%s"
	username= "%s"
	password= "%s"
	}
	`, testUser.InstanceURL, testUser.InstanceUsername, testUser.InstancePassword)
}

func getTestProviders(httpClient *http.Client) map[string]func() (tfprotov6.ProviderServer, error) {
	cloudconnectorProvider := NewWithClient(httpClient).(*cloudConnectorProvider)

	return map[string]func() (tfprotov6.ProviderServer, error){
		"cloudconnector": providerserver.NewProtocol6WithError(cloudconnectorProvider),
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
		user.InstanceUsername = os.Getenv("CC_USERNAME")
		user.InstancePassword = os.Getenv("CC_PASSWORD")
		user.InstanceURL = os.Getenv("CC_INSTANCE_URL")

		user.CloudUsername = os.Getenv("TF_CLOUD_USER")
		user.CloudPassword = os.Getenv("TF_CLOUD_PASSWORD")
		if len(user.InstanceUsername) == 0 || len(user.InstancePassword) == 0 || len(user.InstanceURL) == 0 {
			t.Fatal("Env vars CC_USERNAME, CC_PASSWORD and CC_INSTANCE_URL are required when recording test fixtures")
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

		ipPattern := regexp.MustCompile(`https://(?:\d{1,3}\.){3}\d{1,3}(?::\d+)?`)
		hostPattern := regexp.MustCompile(`^(?:\d{1,3}\.){3}\d{1,3}(?::\d+)?$`)
		i.Request.URL = ipPattern.ReplaceAllString(i.Request.URL, redactedTestUser.InstanceURL)
		i.Request.Host = hostPattern.ReplaceAllString(i.Request.Host, redactedTestUser.InstanceURL)

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

func TestCCProvider_AllResources(t *testing.T) {

	expectedResources := []string{
		"cloudconnector_domain_mapping",
		"cloudconnector_subaccount",
		"cloudconnector_system_mapping_resource",
		"cloudconnector_system_mapping",
		"cloudconnector_subaccount_k8s_service_channel",
	}

	ctx := context.Background()
	registeredResources := []string{}

	for _, resourceFunc := range New().Resources(ctx) {
		var resp resource.MetadataResponse

		resourceFunc().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cloudconnector"}, &resp)

		registeredResources = append(registeredResources, resp.TypeName)
	}

	assert.ElementsMatch(t, expectedResources, registeredResources)
}

func TestCCProvider_AllDataSources(t *testing.T) {

	expectedDataSources := []string{
		"cloudconnector_domain_mapping",
		"cloudconnector_domain_mappings",
		"cloudconnector_subaccount_configuration",
		"cloudconnector_subaccounts",
		"cloudconnector_system_mapping_resource",
		"cloudconnector_system_mapping_resources",
		"cloudconnector_system_mapping",
		"cloudconnector_system_mappings",
		"cloudconnector_subaccount_k8s_service_channel",
		"cloudconnector_subaccount_k8s_service_channels",
	}

	ctx := context.Background()
	registeredDataSources := []string{}

	for _, datasourceFunc := range New().DataSources(ctx) {
		var resp datasource.MetadataResponse

		datasourceFunc().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "cloudconnector"}, &resp)

		registeredDataSources = append(registeredDataSources, resp.TypeName)
	}

	assert.ElementsMatch(t, expectedDataSources, registeredDataSources)
}
