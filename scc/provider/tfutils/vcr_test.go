package tfutils

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

func TestRegexpValidUUID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, uuidvalidator.UuidRegexp, RegexpValidUUID)
	assert.True(t, RegexpValidUUID.MatchString("550e8400-e29b-41d4-a716-446655440000"))
	assert.False(t, RegexpValidUUID.MatchString("not-a-uuid"))
}

func TestRegexpValidTimeStamp(t *testing.T) {
	t.Parallel()

	assert.True(t, RegexpValidTimeStamp.MatchString("2024-01-15 08:30:00"))
	assert.True(t, RegexpValidTimeStamp.MatchString("2024-01-15 08:30:00 +0000"))
	assert.False(t, RegexpValidTimeStamp.MatchString("2024-01-15"))
	assert.False(t, RegexpValidTimeStamp.MatchString("not-a-timestamp"))
}

func TestRegexpValidSerialNumber(t *testing.T) {
	t.Parallel()

	assert.True(t, RegexpValidSerialNumber.MatchString("aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"))
	assert.False(t, RegexpValidSerialNumber.MatchString("aa:bb"))
	assert.False(t, RegexpValidSerialNumber.MatchString("not-a-serial"))
}

func TestProviderConfig(t *testing.T) {
	t.Parallel()

	user := User{
		InstanceURL:      "https://example.com",
		InstanceUsername: "admin",
		InstancePassword: "secret",
	}

	config := ProviderConfig(user)

	assert.Contains(t, config, `"https://example.com"`)
	assert.Contains(t, config, `"admin"`)
	assert.Contains(t, config, `"secret"`)
}

func TestGetTestProviders(t *testing.T) {
	t.Parallel()

	providers := GetTestProviders(&http.Client{})

	require.Contains(t, providers, "scc")
	assert.NotNil(t, providers["scc"])
}

func TestSetupVCR_ReplayMode(t *testing.T) {
	t.Parallel()

	// Uses an existing fixture cassette from the datasources package
	rec, user := SetupVCR(t, "../datasources/fixtures/datasource_proxy_settings")
	defer StopQuietly(rec)

	assert.NotNil(t, rec)
	assert.Equal(t, redactedTestUser, user)
	assert.False(t, rec.IsRecording())
}

func TestHookRedactSensitiveCredentials(t *testing.T) {
	t.Parallel()

	hook := hookRedactSensitiveCredentials()

	i := &cassette.Interaction{
		Request: cassette.Request{
			URL:  "https://real-host.example.com:8443/api",
			Host: "real-host.example.com:8443",
			Headers: http.Header{
				"Authorization": []string{"Basic abc123"},
				"X-Csrf-Token":  []string{"token-value"},
				"Content-Type":  []string{"application/json"},
			},
		},
		Response: cassette.Response{
			Headers: http.Header{
				"Set-Cookie": []string{"session=abc"},
				"Location":   []string{"https://somewhere.else"},
				"X-Custom":   []string{"keep-me"},
			},
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.Equal(t, []string{"redacted"}, i.Request.Headers["Authorization"])
	assert.Equal(t, []string{"redacted"}, i.Request.Headers["X-Csrf-Token"])
	assert.Equal(t, []string{"application/json"}, i.Request.Headers["Content-Type"])
	assert.Equal(t, []string{"redacted"}, i.Response.Headers["Set-Cookie"])
	assert.Equal(t, []string{"redacted"}, i.Response.Headers["Location"])
	assert.Equal(t, []string{"keep-me"}, i.Response.Headers["X-Custom"])
	assert.Contains(t, i.Request.URL, redactedTestUser.InstanceURL)
}

func TestHookRedactSensitiveBody(t *testing.T) {
	t.Parallel()

	hook := hookRedactSensitiveBody()

	i := &cassette.Interaction{
		Request: cassette.Request{
			Body: `{"cloudPassword":"mySecret","cloudUser":"me@example.com","authenticationData":"authdata123","k8sCluster":"cluster.host","k8sService":"svc-id","abapCloudTenantHost":"tenant.host"}`,
		},
		Response: cassette.Response{
			Body: `{"k8sCluster":"cluster.host","k8sService":"svc-id","abapCloudTenantHost":"tenant.host","subjectDN":"CN=real,O=real","issuer":"CN=real","serialNumber":"aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99","tunnel":{"user":"real@example.com"}}`,
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.Contains(t, i.Request.Body, redactedTestUser.CloudPassword)
	assert.Contains(t, i.Request.Body, redactedTestUser.CloudUsername)
	assert.Contains(t, i.Request.Body, redactedTestUser.CloudAuthenticationData)
	assert.Contains(t, i.Request.Body, redactedTestUser.K8SCluster)
	assert.Contains(t, i.Request.Body, redactedTestUser.K8SService)
	assert.Contains(t, i.Request.Body, redactedTestUser.ABAPCloudTenantHost)
	assert.Contains(t, i.Response.Body, redactedTestUser.K8SCluster)
	assert.Contains(t, i.Response.Body, redactedTestUser.K8SService)
	assert.Contains(t, i.Response.Body, redactedTestUser.ABAPCloudTenantHost)
	assert.Contains(t, i.Response.Body, redactedTestUser.CloudUsername)
}

func TestHookRedactBinaryCertificate_OctetStream(t *testing.T) {
	t.Parallel()

	hook := hookRedactBinaryCertificate()

	i := &cassette.Interaction{
		Response: cassette.Response{
			Body: "binary-cert-data",
			Headers: http.Header{
				"Content-Type": []string{"application/octet-stream"},
			},
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.Equal(t, "REDACTED_BINARY_CERTIFICATE", i.Response.Body)
}

func TestHookRedactBinaryCertificate_DERDisposition(t *testing.T) {
	t.Parallel()

	hook := hookRedactBinaryCertificate()

	i := &cassette.Interaction{
		Response: cassette.Response{
			Body: "binary-cert-data",
			Headers: http.Header{
				"Content-Disposition": []string{"attachment; filename=ca_certificate.der"},
			},
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.Equal(t, "REDACTED_BINARY_CERTIFICATE", i.Response.Body)
}

func TestHookRedactBinaryCertificate_NoMatch(t *testing.T) {
	t.Parallel()

	hook := hookRedactBinaryCertificate()

	i := &cassette.Interaction{
		Response: cassette.Response{
			Body: "normal response",
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.Equal(t, "normal response", i.Response.Body)
}

func TestHookRedactBodyLinks(t *testing.T) {
	t.Parallel()

	hook := hookRedactBodyLinks()

	i := &cassette.Interaction{
		Response: cassette.Response{
			Body: `{"_links":{"self":{"href":"https://real-host.example.com/api/resource"}}}`,
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.NotContains(t, i.Response.Body, "real-host.example.com")
	assert.Contains(t, i.Response.Body, "redacted.url")
}

func TestHookRedactBodyLinks_NoLinks(t *testing.T) {
	t.Parallel()

	hook := hookRedactBodyLinks()

	original := `{"name":"value"}`
	i := &cassette.Interaction{
		Response: cassette.Response{
			Body: original,
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.Equal(t, original, i.Response.Body)
}

// ---------------------------------------------------------------------------
// requestMatcher
// ---------------------------------------------------------------------------

func TestRequestMatcher_Match(t *testing.T) {
	t.Parallel()

	matcher := requestMatcher(t)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/api", strings.NewReader("body"))
	recorded := cassette.Request{
		Method: http.MethodGet,
		URL:    "https://example.com/api",
		Body:   "body",
	}

	assert.True(t, matcher(req, recorded))
}

func TestRequestMatcher_MethodMismatch(t *testing.T) {
	t.Parallel()

	matcher := requestMatcher(t)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/api", strings.NewReader("body"))
	recorded := cassette.Request{
		Method: http.MethodGet,
		URL:    "https://example.com/api",
		Body:   "body",
	}

	assert.False(t, matcher(req, recorded))
}

func TestRequestMatcher_URLMismatch(t *testing.T) {
	t.Parallel()

	matcher := requestMatcher(t)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/other", strings.NewReader("body"))
	recorded := cassette.Request{
		Method: http.MethodGet,
		URL:    "https://example.com/api",
		Body:   "body",
	}

	assert.False(t, matcher(req, recorded))
}

func TestRequestMatcher_BodyMismatch(t *testing.T) {
	t.Parallel()

	matcher := requestMatcher(t)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/api", strings.NewReader("different-body"))
	recorded := cassette.Request{
		Method: http.MethodGet,
		URL:    "https://example.com/api",
		Body:   "body",
	}

	assert.False(t, matcher(req, recorded))
}

// ---------------------------------------------------------------------------
// hookRedactSensitiveBody — subaccountCertificate branch
// ---------------------------------------------------------------------------

func TestHookRedactSensitiveBody_SubaccountCertificate(t *testing.T) {
	t.Parallel()

	hook := hookRedactSensitiveBody()

	i := &cassette.Interaction{
		Request: cassette.Request{Body: ""},
		Response: cassette.Response{
			Body: `{"subaccountCertificate":true,"subjectDN":"CN=real,O=SAP","issuer":"CN=real-issuer,O=SAP","serialNumber":"aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"}`,
		},
	}

	err := hook(i)

	require.NoError(t, err)
	assert.Contains(t, i.Response.Body, "CN=redacted")
	assert.NotContains(t, i.Response.Body, "CN=real,O=SAP")
	assert.NotContains(t, i.Response.Body, "CN=real-issuer")
	assert.NotContains(t, i.Response.Body, "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99")
}

// ---------------------------------------------------------------------------
// SetupVCR record mode — exercises the TEST_RECORD=true branch
// ---------------------------------------------------------------------------

func TestSetupVCR_RecordMode(t *testing.T) {
	// Not parallel: mutates env var
	t.Setenv("TEST_RECORD", "true")
	t.Setenv("SCC_USERNAME", "record-user@example.com")
	t.Setenv("SCC_PASSWORD", "record-password")
	t.Setenv("SCC_INSTANCE_URL", "https://record.instance.url")
	t.Setenv("TF_VAR_cloud_user", "cloud@example.com")
	t.Setenv("TF_VAR_cloud_password", "cloud-pass")
	t.Setenv("TF_VAR_authentication_data", "auth-data")
	t.Setenv("TF_VAR_k8s_cluster_host", "k8s.cluster.host")
	t.Setenv("TF_VAR_k8s_service_id", "k8s-svc-id")
	t.Setenv("TF_VAR_abap_cloud_tenant_host", "abap.tenant.host")

	cassettePath := filepath.Join(t.TempDir(), "record_mode_cassette")

	rec, user := SetupVCR(t, cassettePath)
	defer StopQuietly(rec)

	assert.True(t, rec.IsRecording())
	assert.Equal(t, "record-user@example.com", user.InstanceUsername)
	assert.Equal(t, "record-password", user.InstancePassword)
	assert.Equal(t, "https://record.instance.url", user.InstanceURL)
	assert.Equal(t, "cloud@example.com", user.CloudUsername)
	assert.Equal(t, "cloud-pass", user.CloudPassword)
	assert.Equal(t, "auth-data", user.CloudAuthenticationData)
	assert.Equal(t, "k8s.cluster.host", user.K8SCluster)
	assert.Equal(t, "k8s-svc-id", user.K8SService)
	assert.Equal(t, "abap.tenant.host", user.ABAPCloudTenantHost)
}
