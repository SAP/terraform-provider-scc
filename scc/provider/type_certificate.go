package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Wrappers for testing purposes (allows mocking in tests)
var selfSignedCertificateResourceValueFromFunc = SelfSignedCertificateResourceValueFrom
var signedChainCertificateResourceValueFromFunc = SignedChainCertificateResourceValueFrom
var pkcs12CertificateResourceValueFromFunc = PKCS12CertificateResourceValueFrom

var selfSignedSystemCertificateResourceValueFromFunc = SelfSignedSystemCertificateResourceValueFrom
var signedChainSystemCertificateResourceValueFromFunc = SignedChainSystemCertificateResourceValueFrom
var pkcs12SystemCertificateResourceValueFromFunc = PKCS12SystemCertificateResourceValueFrom

type SelfSignedSystemCertificateResourceConfig struct {
	KeySize types.Int64 `tfsdk:"key_size"`
	CertificateConfig
}

type SignedChainSystemCertificateResourceConfig struct {
	SignedChain types.String `tfsdk:"signed_chain"`
	CertificateConfig
}

type PKCS12SystemCertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	CertificateConfig
}

type SelfSignedCertificateResourceConfig struct {
	KeySize types.Int64 `tfsdk:"key_size"`
	CertificateWithSANConfig
}

type SignedChainCertificateResourceConfig struct {
	SignedChain types.String `tfsdk:"signed_chain"`
	CertificateWithSANConfig
}

type PKCS12CertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	CertificateWithSANConfig
}

type CSRActionConfig struct {
	Type                    types.String `tfsdk:"type"`
	KeySize                 types.Int64  `tfsdk:"key_size"`
	SubjectDN               types.Object `tfsdk:"subject_dn"`
	SubjectAlternativeNames types.List   `tfsdk:"subject_alternative_names"`
}

func CertificateDataSourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (CertificateWithSANConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModelWithSAN(ctx, value, pemBytes)
	if diags.HasError() {
		return CertificateWithSANConfig{}, diags
	}

	return certificateConfig, diag.Diagnostics{}
}

func SelfSignedCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte, existingDN *certificateSubjectDNConfig) (SelfSignedCertificateResourceConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModelWithSAN(ctx, value, pemBytes)
	if diags.HasError() {
		return SelfSignedCertificateResourceConfig{}, diags
	}

	if existingDN != nil {
		certificateConfig.SubjectDN = buildSubjectDNObject(existingDN)
	}

	model := &SelfSignedCertificateResourceConfig{
		CertificateWithSANConfig: certificateConfig,
	}

	return *model, diag.Diagnostics{}
}

func SignedChainCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (SignedChainCertificateResourceConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModelWithSAN(ctx, value, pemBytes)
	if diags.HasError() {
		return SignedChainCertificateResourceConfig{}, diags
	}

	model := &SignedChainCertificateResourceConfig{
		CertificateWithSANConfig: certificateConfig,
	}

	return *model, diag.Diagnostics{}
}

func PKCS12CertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (PKCS12CertificateResourceConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModelWithSAN(ctx, value, pemBytes)
	if diags.HasError() {
		return PKCS12CertificateResourceConfig{}, diags
	}

	model := &PKCS12CertificateResourceConfig{
		CertificateWithSANConfig: certificateConfig,
	}

	return *model, diag.Diagnostics{}
}

func SystemCertificateDataSourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (CertificateConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModel(ctx, value, pemBytes)
	if diags.HasError() {
		return CertificateConfig{}, diags
	}

	return certificateConfig, diag.Diagnostics{}
}

func SelfSignedSystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte, existingDN *certificateSubjectDNConfig) (SelfSignedSystemCertificateResourceConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModel(ctx, value, pemBytes)
	if diags.HasError() {
		return SelfSignedSystemCertificateResourceConfig{}, diags
	}

	if existingDN != nil {
		certificateConfig.SubjectDN = buildSubjectDNObject(existingDN)
	}

	model := &SelfSignedSystemCertificateResourceConfig{
		CertificateConfig: certificateConfig,
	}

	return *model, diag.Diagnostics{}
}

func SignedChainSystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (SignedChainSystemCertificateResourceConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModel(ctx, value, pemBytes)
	if diags.HasError() {
		return SignedChainSystemCertificateResourceConfig{}, diags
	}

	model := &SignedChainSystemCertificateResourceConfig{
		CertificateConfig: certificateConfig,
	}

	return *model, diag.Diagnostics{}
}

func PKCS12SystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (PKCS12SystemCertificateResourceConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModel(ctx, value, pemBytes)
	if diags.HasError() {
		return PKCS12SystemCertificateResourceConfig{}, diags
	}

	model := &PKCS12SystemCertificateResourceConfig{
		CertificateConfig: certificateConfig,
	}

	return *model, diag.Diagnostics{}
}
