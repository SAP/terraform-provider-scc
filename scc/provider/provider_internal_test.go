package provider

// White-box tests for unexported helpers in provider.go.
// These live in package provider (not provider_test) to access unexported symbols.

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// getBoolAttribute
// ---------------------------------------------------------------------------

func TestGetBoolAttribute_HCLTrue(t *testing.T) {
	result := getBoolAttribute(types.BoolValue(true), "SCC_SKIP_SSL_VALIDATION")
	assert.True(t, result)
}

func TestGetBoolAttribute_HCLFalse(t *testing.T) {
	result := getBoolAttribute(types.BoolValue(false), "SCC_SKIP_SSL_VALIDATION")
	assert.False(t, result)
}

func TestGetBoolAttribute_NullFallsBackToEnv_True(t *testing.T) {
	t.Setenv("SCC_TEST_BOOL", "true")
	result := getBoolAttribute(types.BoolNull(), "SCC_TEST_BOOL")
	assert.True(t, result)
}

func TestGetBoolAttribute_NullFallsBackToEnv_Missing(t *testing.T) {
	t.Setenv("SCC_TEST_BOOL_MISSING", "")
	result := getBoolAttribute(types.BoolNull(), "SCC_TEST_BOOL_MISSING")
	assert.False(t, result)
}

// ---------------------------------------------------------------------------
// getNonEmptyAttribute
// ---------------------------------------------------------------------------

func TestGetNonEmptyAttribute_HCLValue(t *testing.T) {
	result := getNonEmptyAttribute(types.StringValue("from-hcl"), "SCC_TEST_ATTR")
	assert.Equal(t, "from-hcl", result)
}

func TestGetNonEmptyAttribute_EmptyHCLFallsBackToEnv(t *testing.T) {
	t.Setenv("SCC_TEST_ATTR", "from-env")
	result := getNonEmptyAttribute(types.StringValue(""), "SCC_TEST_ATTR")
	assert.Equal(t, "from-env", result)
}

func TestGetNonEmptyAttribute_NullFallsBackToEnv(t *testing.T) {
	t.Setenv("SCC_TEST_ATTR2", "env-value")
	result := getNonEmptyAttribute(types.StringNull(), "SCC_TEST_ATTR2")
	assert.Equal(t, "env-value", result)
}

func TestGetNonEmptyAttribute_NullNoEnv(t *testing.T) {
	t.Setenv("SCC_TEST_ATTR_MISSING", "")
	result := getNonEmptyAttribute(types.StringNull(), "SCC_TEST_ATTR_MISSING")
	assert.Equal(t, "", result)
}

// ---------------------------------------------------------------------------
// resolveAttributes
// ---------------------------------------------------------------------------

func TestResolveAttributes_FromHCL(t *testing.T) {
	cfg := CloudConnectorProviderData{
		InstanceURL:       types.StringValue("https://hcl.example.com"),
		Username:          types.StringValue("hcl-user"),
		Password:          types.StringValue("hcl-pass"),
		CaCertificate:     types.StringValue("hcl-ca"),
		ClientCertificate: types.StringValue("hcl-cert"),
		ClientKey:         types.StringValue("hcl-key"),
		SkipSSLValidation: types.BoolValue(true),
	}

	url, user, pass, ca, cert, key, skip := resolveAttributes(cfg)

	assert.Equal(t, "https://hcl.example.com", url)
	assert.Equal(t, "hcl-user", user)
	assert.Equal(t, "hcl-pass", pass)
	assert.Equal(t, "hcl-ca", ca)
	assert.Equal(t, "hcl-cert", cert)
	assert.Equal(t, "hcl-key", key)
	assert.True(t, skip)
}

func TestResolveAttributes_FromEnv(t *testing.T) {
	t.Setenv("SCC_INSTANCE_URL", "https://env.example.com")
	t.Setenv("SCC_USERNAME", "env-user")
	t.Setenv("SCC_PASSWORD", "env-pass")
	t.Setenv("SCC_CA_CERTIFICATE", "env-ca")
	t.Setenv("SCC_CLIENT_CERTIFICATE", "env-cert")
	t.Setenv("SCC_CLIENT_KEY", "env-key")
	t.Setenv("SCC_SKIP_SSL_VALIDATION", "true")

	cfg := CloudConnectorProviderData{
		InstanceURL:       types.StringNull(),
		Username:          types.StringNull(),
		Password:          types.StringNull(),
		CaCertificate:     types.StringNull(),
		ClientCertificate: types.StringNull(),
		ClientKey:         types.StringNull(),
		SkipSSLValidation: types.BoolNull(),
	}

	url, user, pass, ca, cert, key, skip := resolveAttributes(cfg)

	assert.Equal(t, "https://env.example.com", url)
	assert.Equal(t, "env-user", user)
	assert.Equal(t, "env-pass", pass)
	assert.Equal(t, "env-ca", ca)
	assert.Equal(t, "env-cert", cert)
	assert.Equal(t, "env-key", key)
	assert.True(t, skip)
}
