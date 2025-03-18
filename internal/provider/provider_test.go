package provider

import (
	"context"
	// "crypto/tls"
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
	regexValidSerialNumber = regexp.MustCompile(`^(?:[0-9a-fA-F]{2}:){15}[0-9a-fA-F]{2}$`)
)

type User struct {
	Username string
	Password string
}

var redactedTestUser = User{
	Username: "Administrator",
	Password: "Terraform",
}

func providerConfig(_ string, testUser User) string {
	instance_url := "https://10.52.109.11:8443"
	return fmt.Sprintf(`
	provider "cloudconnector" {
	instance_url= "%s"
	username= "%s"
	password= "%s"
	}
	`, instance_url, testUser.Username, testUser.Password)
}

func getTestProviders(httpClient *http.Client) map[string]func() (tfprotov6.ProviderServer, error) {
	// httpClient.Transport = &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
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

	rec, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassetteName,
		Mode:               mode,
		SkipRequestLatency: true,
		RealTransport:      http.DefaultTransport,
	})

	user := redactedTestUser
	if rec.IsRecording() {
		t.Logf("ATTENTION: Recording '%s'", cassetteName)
		user.Username = os.Getenv("CC_USERNAME")
		user.Password = os.Getenv("CC_PASSWORD")
		if len(user.Username) == 0 || len(user.Password) == 0 {
			t.Fatal("Env vars CC_USERNAME and CC_PASSWORD are required when recording test fixtures")
		}
	} else {
		t.Logf("Replaying '%s'", cassetteName)
	}

	if err != nil {
		t.Fatal()
	}

	rec.SetMatcher(requestMatcher(t))
	rec.AddHook(redactAuthorizationToken(), recorder.BeforeSaveHook)

	return rec, user
}

func requestMatcher(t *testing.T) cassette.MatcherFunc {
	return func(r *http.Request, i cassette.Request) bool {
		t.Logf("Request: %s %s", r.Method, r.URL.String())
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Logf("Error reading request body: %v", err)
			return false
		}
		t.Logf("Request Body: %s", string(bytes))
		return r.Method == i.Method && r.URL.String() == i.URL
	}
}

func redactAuthorizationToken() recorder.HookFunc {
	return func(i *cassette.Interaction) error {

		redact := func(headers map[string][]string) {
			for key := range headers {
				if strings.Contains(strings.ToLower(key), "authorization") {
					headers[key] = []string{"redacted"}
				}
			}
		}

		redact(i.Request.Headers)
		redact(i.Response.Headers)

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
		"cloudconnector_subaccount_service_channel_k8s",
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
		"cloudconnector_domain_mappings",
		"cloudconnector_subaccount",
		"cloudconnector_subaccounts",
		"cloudconnector_system_mapping_resource",
		"cloudconnector_system_mapping_resources",
		"cloudconnector_system_mapping",
		"cloudconnector_system_mappings",
		"cloudconnector_subaccount_service_channel_k8s",
		"cloudconnector_subaccount_service_channels_k8s",
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
