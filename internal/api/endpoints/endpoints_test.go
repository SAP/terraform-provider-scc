package endpoints

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Certificate endpoints
// ---------------------------------------------------------------------------

func TestGetCACertificateEndpoint(t *testing.T) {
	ep := GetCACertificateEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"), "should be a relative path")
}

func TestGetSystemCertificateEndpoint(t *testing.T) {
	ep := GetSystemCertificateEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

func TestGetUICertificateEndpoint(t *testing.T) {
	ep := GetUICertificateEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

// All three certificate endpoints must be distinct
func TestCertificateEndpoints_Distinct(t *testing.T) {
	ca := GetCACertificateEndpoint()
	sys := GetSystemCertificateEndpoint()
	ui := GetUICertificateEndpoint()
	assert.NotEqual(t, ca, sys)
	assert.NotEqual(t, ca, ui)
	assert.NotEqual(t, sys, ui)
}

// ---------------------------------------------------------------------------
// Backup endpoint
// ---------------------------------------------------------------------------

func TestGetBackupEndpoint(t *testing.T) {
	ep := GetBackupEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

// ---------------------------------------------------------------------------
// Proxy settings endpoint
// ---------------------------------------------------------------------------

func TestGetProxySettingsEndpoint(t *testing.T) {
	ep := GetProxySettingsEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

// ---------------------------------------------------------------------------
// Master instance endpoint
// ---------------------------------------------------------------------------

func TestGetMasterInstanceBaseEndpoint(t *testing.T) {
	ep := GetMasterInstanceBaseEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

// ---------------------------------------------------------------------------
// Backend trust store endpoints
// ---------------------------------------------------------------------------

func TestGetBackendTrustStoreBaseEndpoint(t *testing.T) {
	ep := GetBackendTrustStoreBaseEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

func TestGetBackendTrustStoreCertificateEndpoint(t *testing.T) {
	ep := GetBackendTrustStoreCertificateEndpoint()
	base := GetBackendTrustStoreBaseEndpoint()
	assert.True(t, strings.HasPrefix(ep, base), "certificate endpoint should extend base")
}

// ---------------------------------------------------------------------------
// Subject pattern rules endpoints
// ---------------------------------------------------------------------------

func TestGetSubjectPatternRulesBaseEndpoint(t *testing.T) {
	ep := GetSubjectPatternRulesBaseEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

func TestGetSubjectPatternRuleByIndexEndpoint(t *testing.T) {
	base := GetSubjectPatternRulesBaseEndpoint()
	ep0 := GetSubjectPatternRuleByIndexEndpoint(0)
	ep5 := GetSubjectPatternRuleByIndexEndpoint(5)
	assert.True(t, strings.HasPrefix(ep0, base))
	assert.Contains(t, ep0, "0")
	assert.Contains(t, ep5, "5")
	assert.NotEqual(t, ep0, ep5)
}

// ---------------------------------------------------------------------------
// Subaccount endpoints
// ---------------------------------------------------------------------------

func TestGetSubaccountBaseEndpoint(t *testing.T) {
	ep := GetSubaccountBaseEndpoint()
	assert.NotEmpty(t, ep)
	assert.True(t, strings.HasPrefix(ep, "/"))
}

func TestGetSubaccountEndpoint(t *testing.T) {
	base := GetSubaccountBaseEndpoint()
	ep := GetSubaccountEndpoint("eu12.hana.ondemand.com", "my-subaccount")
	assert.True(t, strings.HasPrefix(ep, base))
	assert.Contains(t, ep, "eu12.hana.ondemand.com")
	assert.Contains(t, ep, "my-subaccount")
}

// ---------------------------------------------------------------------------
// Domain mapping endpoints
// ---------------------------------------------------------------------------

func TestGetDomainMappingBaseEndpoint(t *testing.T) {
	ep := GetDomainMappingBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount")
	assert.Contains(t, ep, "eu12.hana.ondemand.com")
	assert.Contains(t, ep, "my-subaccount")
	assert.Contains(t, ep, "domainMappings")
}

func TestGetDomainMappingEndpoint(t *testing.T) {
	base := GetDomainMappingBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount")
	ep := GetDomainMappingEndpoint("eu12.hana.ondemand.com", "my-subaccount", "internal.corp")
	assert.True(t, strings.HasPrefix(ep, base))
	assert.Contains(t, ep, "internal.corp")
}

// ---------------------------------------------------------------------------
// Subaccount service channel endpoints
// ---------------------------------------------------------------------------

func TestGetSubaccountServiceChannelBaseEndpoint(t *testing.T) {
	ep := GetSubaccountServiceChannelBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount", "kubernetes")
	assert.Contains(t, ep, "eu12.hana.ondemand.com")
	assert.Contains(t, ep, "my-subaccount")
	assert.Contains(t, ep, "kubernetes")
}

func TestGetSubaccountServiceChannelEndpoint(t *testing.T) {
	base := GetSubaccountServiceChannelBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount", "kubernetes")
	ep := GetSubaccountServiceChannelEndpoint("eu12.hana.ondemand.com", "my-subaccount", "kubernetes", 42)
	assert.True(t, strings.HasPrefix(ep, base))
	assert.Contains(t, ep, "42")
}

// ---------------------------------------------------------------------------
// System mapping endpoints
// ---------------------------------------------------------------------------

func TestGetSystemMappingBaseEndpoint(t *testing.T) {
	ep := GetSystemMappingBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount")
	assert.Contains(t, ep, "eu12.hana.ondemand.com")
	assert.Contains(t, ep, "my-subaccount")
	assert.Contains(t, ep, "systemMappings")
}

func TestGetSystemMappingEndpoint(t *testing.T) {
	base := GetSystemMappingBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount")
	ep := GetSystemMappingEndpoint("eu12.hana.ondemand.com", "my-subaccount", "vhost", "8080")
	assert.True(t, strings.HasPrefix(ep, base))
	assert.Contains(t, ep, "vhost")
	assert.Contains(t, ep, "8080")
}

// ---------------------------------------------------------------------------
// System mapping resource endpoints
// ---------------------------------------------------------------------------

func TestGetSystemMappingResourceBaseEndpoint(t *testing.T) {
	ep := GetSystemMappingResourceBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount", "vhost", "8080")
	assert.Contains(t, ep, "resources")
	assert.Contains(t, ep, "vhost")
}

func TestGetSystemMappingResourceEndpoint(t *testing.T) {
	base := GetSystemMappingResourceBaseEndpoint("eu12.hana.ondemand.com", "my-subaccount", "vhost", "8080")
	ep := GetSystemMappingResourceEndpoint("eu12.hana.ondemand.com", "my-subaccount", "vhost", "8080", "res-1")
	assert.True(t, strings.HasPrefix(ep, base))
	assert.Contains(t, ep, "res-1")
}
