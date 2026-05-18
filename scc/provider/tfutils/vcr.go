package tfutils

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

var (
	RegexpValidUUID         = uuidvalidator.UuidRegexp
	RegexpValidTimeStamp    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?: \+\d{4})?$`)
	RegexpValidSerialNumber = regexp.MustCompile(`^(?:[0-9a-fA-F]{2}:){14,}[0-9a-fA-F]{1,2}$`)
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

func ProviderConfig(testUser User) string {
	return fmt.Sprintf(`
	provider "scc" {
	instance_url= "%s"
	username= "%s"
	password= "%s"
	}
	`, testUser.InstanceURL, testUser.InstanceUsername, testUser.InstancePassword)
}

func GetTestProviders(httpClient *http.Client) map[string]func() (tfprotov6.ProviderServer, error) {
	cloudconnectorProvider := provider.NewWithClient(httpClient).(*provider.CloudConnectorProvider)

	return map[string]func() (tfprotov6.ProviderServer, error){
		"scc": providerserver.NewProtocol6WithError(cloudconnectorProvider),
	}
}

func SetupVCR(t *testing.T, cassetteName string) (*recorder.Recorder, User) {
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
		t.Fatal(err)
	}

	rec.SetMatcher(requestMatcher(t))
	rec.AddHook(hookRedactSensitiveCredentials(), recorder.BeforeSaveHook)
	rec.AddHook(hookRedactBodyLinks(), recorder.BeforeSaveHook)
	rec.AddHook(hookRedactSensitiveBody(), recorder.BeforeSaveHook)
	rec.AddHook(hookRedactBinaryCertificate(), recorder.BeforeSaveHook)

	return rec, user
}

func StopQuietly(rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		panic(err)
	}
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

		if strings.Contains(i.Response.Body, "subjectDN") {
			reSubjectDN := regexp.MustCompile(`"subjectDN"\s*:\s*".*?"`)
			i.Response.Body = reSubjectDN.ReplaceAllString(i.Response.Body, `"subjectDN": "CN=redacted,L=redacted,OU=redacted,OU=redacted,O=redacted,C=redacted"`)
		}

		if strings.Contains(i.Response.Body, "issuer") {
			reIssuer := regexp.MustCompile(`"issuer"\s*:\s*".*?"`)
			i.Response.Body = reIssuer.ReplaceAllString(i.Response.Body, `"issuer": "CN=redacted,OU=redacted,O=redacted,L=redacted,C=redacted"`)
		}

		if strings.Contains(i.Response.Body, "serialNumber") {
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

func hookRedactBinaryCertificate() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		// Detect SCC certificate binary response
		if strings.Contains(strings.ToLower(i.Response.Headers.Get("Content-Type")), "octet-stream") ||
			strings.Contains(strings.ToLower(i.Response.Headers.Get("Content-Disposition")), "ca_certificate.der") {

			i.Response.Body = "REDACTED_BINARY_CERTIFICATE"
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
