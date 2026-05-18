package model

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Wrappers for CA Certificate testing purposes (allows mocking in tests)
var SelfSignedCACertificateResourceValueFromFunc = selfSignedCACertificateResourceValueFrom
var SignedChainCACertificateResourceValueFromFunc = signedChainCACertificateResourceValueFrom
var PKCS12CACertificateResourceValueFromFunc = pkcs12CACertificateResourceValueFrom

// Wrappers for System Certificate testing purposes (allows mocking in tests)
var SelfSignedSystemCertificateResourceValueFromFunc = selfSignedSystemCertificateResourceValueFrom
var SignedChainSystemCertificateResourceValueFromFunc = signedChainSystemCertificateResourceValueFrom
var PKCS12SystemCertificateResourceValueFromFunc = pkcs12SystemCertificateResourceValueFrom

// Wrappers for UI Certificate testing purposes (allows mocking in tests)
var SelfSignedUICertificateResourceValueFromFunc = selfSignedUICertificateResourceValueFrom
var SignedChainUICertificateResourceValueFromFunc = signedChainUICertificateResourceValueFrom
var PKCS12UICertificateResourceValueFromFunc = pkcs12UICertificateResourceValueFrom

type CACertificateDataSourceConfig struct {
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	helpers.CertificateWithSANConfig
}

type SystemCertificateDataSourceConfig struct {
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	helpers.CertificateConfig
}

type SelfSignedSystemCertificateResourceConfig struct {
	KeySize        types.Int64  `tfsdk:"key_size"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	helpers.CertificateConfig
}

type SignedChainSystemCertificateResourceConfig struct {
	SignedChain    types.String `tfsdk:"signed_chain"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	helpers.CertificateConfig
}

type PKCS12SystemCertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	CertificatePEM    types.String `tfsdk:"certificate_pem"`
	helpers.CertificateConfig
}

type SelfSignedCACertificateResourceConfig struct {
	KeySize        types.Int64  `tfsdk:"key_size"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	helpers.CertificateWithSANConfig
}

type SignedChainCACertificateResourceConfig struct {
	SignedChain    types.String `tfsdk:"signed_chain"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	helpers.CertificateWithSANConfig
}

type PKCS12CACertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	CertificatePEM    types.String `tfsdk:"certificate_pem"`
	helpers.CertificateWithSANConfig
}

type SelfSignedUICertificateResourceConfig struct {
	KeySize types.Int64 `tfsdk:"key_size"`
	helpers.CertificateWithSANConfig
}

type SignedChainUICertificateResourceConfig struct {
	SignedChain types.String `tfsdk:"signed_chain"`
	helpers.CertificateWithSANConfig
}

type PKCS12UICertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	helpers.CertificateWithSANConfig
}

type CSRActionConfig struct {
	Type                    types.String `tfsdk:"type"`
	KeySize                 types.Int64  `tfsdk:"key_size"`
	SubjectDN               types.Object `tfsdk:"subject_dn"`
	SubjectAlternativeNames types.List   `tfsdk:"subject_alternative_names"`
}

func buildCertificateConfig(ctx context.Context, value apiobjects.Certificate, existingDN *helpers.CertificateSubjectDNConfig) (helpers.CertificateConfig, diag.Diagnostics) {
	certificateConfig, diags := helpers.BuildCertificateModelFunc(ctx, value)
	if diags.HasError() {
		return helpers.CertificateConfig{}, diags
	}

	if existingDN != nil {
		certificateConfig.SubjectDN = helpers.BuildSubjectDNObjectFunc(existingDN)
	}

	return certificateConfig, diag.Diagnostics{}
}

func buildCertificateWithSANConfig(ctx context.Context, value apiobjects.Certificate, existingDN *helpers.CertificateSubjectDNConfig) (helpers.CertificateWithSANConfig, diag.Diagnostics) {
	certificateConfig, diags := helpers.BuildCertificateModelWithSANFunc(ctx, value)
	if diags.HasError() {
		return helpers.CertificateWithSANConfig{}, diags
	}

	if existingDN != nil {
		certificateConfig.SubjectDN = helpers.BuildSubjectDNObjectFunc(existingDN)
	}

	return certificateConfig, diag.Diagnostics{}
}

func CACertificateDataSourceValueFrom(ctx context.Context, value apiobjects.Certificate) (CACertificateDataSourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, nil)
	if diags.HasError() {
		return CACertificateDataSourceConfig{}, diags
	}

	return CACertificateDataSourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}

}

func SystemCertificateDataSourceValueFrom(ctx context.Context, value apiobjects.Certificate) (SystemCertificateDataSourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateConfig(ctx, value, nil)
	if diags.HasError() {
		return SystemCertificateDataSourceConfig{}, diags
	}

	return SystemCertificateDataSourceConfig{
		CertificateConfig: config,
	}, diag.Diagnostics{}
}

// CA Certificates
func selfSignedCACertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, existingDN *helpers.CertificateSubjectDNConfig) (SelfSignedCACertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, existingDN)
	if diags.HasError() {
		return SelfSignedCACertificateResourceConfig{}, diags
	}

	return SelfSignedCACertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

func signedChainCACertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (SignedChainCACertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, nil)
	if diags.HasError() {
		return SignedChainCACertificateResourceConfig{}, diags
	}

	return SignedChainCACertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

func pkcs12CACertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (PKCS12CACertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, nil)
	if diags.HasError() {
		return PKCS12CACertificateResourceConfig{}, diags
	}

	return PKCS12CACertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

// System Certificates
func selfSignedSystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, existingDN *helpers.CertificateSubjectDNConfig) (SelfSignedSystemCertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateConfig(ctx, value, existingDN)
	if diags.HasError() {
		return SelfSignedSystemCertificateResourceConfig{}, diags
	}

	return SelfSignedSystemCertificateResourceConfig{
		CertificateConfig: config,
	}, diag.Diagnostics{}
}

func signedChainSystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (SignedChainSystemCertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateConfig(ctx, value, nil)
	if diags.HasError() {
		return SignedChainSystemCertificateResourceConfig{}, diags
	}

	return SignedChainSystemCertificateResourceConfig{
		CertificateConfig: config,
	}, diag.Diagnostics{}
}

func pkcs12SystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (PKCS12SystemCertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateConfig(ctx, value, nil)
	if diags.HasError() {
		return PKCS12SystemCertificateResourceConfig{}, diags
	}

	return PKCS12SystemCertificateResourceConfig{
		CertificateConfig: config,
	}, diag.Diagnostics{}
}

// UI Certificates
func selfSignedUICertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, existingDN *helpers.CertificateSubjectDNConfig) (SelfSignedUICertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, existingDN)
	if diags.HasError() {
		return SelfSignedUICertificateResourceConfig{}, diags
	}

	return SelfSignedUICertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

func signedChainUICertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (SignedChainUICertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, nil)
	if diags.HasError() {
		return SignedChainUICertificateResourceConfig{}, diags
	}

	return SignedChainUICertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

func pkcs12UICertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (PKCS12UICertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, nil)
	if diags.HasError() {
		return PKCS12UICertificateResourceConfig{}, diags
	}

	return PKCS12UICertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}
