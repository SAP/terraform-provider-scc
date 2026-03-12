package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Wrappers for CA Certificate testing purposes (allows mocking in tests)
var selfSignedCACertificateResourceValueFromFunc = SelfSignedCACertificateResourceValueFrom
var signedChainCACertificateResourceValueFromFunc = SignedChainCACertificateResourceValueFrom
var pkcs12CACertificateResourceValueFromFunc = PKCS12CACertificateResourceValueFrom

// Wrappers for System Certificate testing purposes (allows mocking in tests)
var selfSignedSystemCertificateResourceValueFromFunc = SelfSignedSystemCertificateResourceValueFrom
var signedChainSystemCertificateResourceValueFromFunc = SignedChainSystemCertificateResourceValueFrom
var pkcs12SystemCertificateResourceValueFromFunc = PKCS12SystemCertificateResourceValueFrom

type CACertificateDataSourceConfig struct {
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	CertificateWithSANConfig
}

type SystemCertificateDataSourceConfig struct {
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	CertificateConfig
}

type SelfSignedSystemCertificateResourceConfig struct {
	KeySize        types.Int64  `tfsdk:"key_size"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	CertificateConfig
}

type SignedChainSystemCertificateResourceConfig struct {
	SignedChain    types.String `tfsdk:"signed_chain"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	CertificateConfig
}

type PKCS12SystemCertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	CertificatePEM    types.String `tfsdk:"certificate_pem"`
	CertificateConfig
}

type SelfSignedCACertificateResourceConfig struct {
	KeySize        types.Int64  `tfsdk:"key_size"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	CertificateWithSANConfig
}

type SignedChainCACertificateResourceConfig struct {
	SignedChain    types.String `tfsdk:"signed_chain"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	CertificateWithSANConfig
}

type PKCS12CACertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	CertificatePEM    types.String `tfsdk:"certificate_pem"`
	CertificateWithSANConfig
}

// type SelfSignedUICertificateResourceConfig struct {
// 	KeySize types.Int64 `tfsdk:"key_size"`
// 	CertificateWithSANConfig
// }

// type SignedChainUICertificateResourceConfig struct {
// 	SignedChain types.String `tfsdk:"signed_chain"`
// 	CertificateWithSANConfig
// }

// type PKCS12UICertificateResourceConfig struct {
// 	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
// 	Password          types.String `tfsdk:"password"`
// 	KeyPassword       types.String `tfsdk:"key_password"`
// 	CertificateWithSANConfig
// }

type CSRActionConfig struct {
	Type                    types.String `tfsdk:"type"`
	KeySize                 types.Int64  `tfsdk:"key_size"`
	SubjectDN               types.Object `tfsdk:"subject_dn"`
	SubjectAlternativeNames types.List   `tfsdk:"subject_alternative_names"`
}

func buildCertificateConfig(ctx context.Context, value apiobjects.Certificate, existingDN *certificateSubjectDNConfig) (CertificateConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModel(ctx, value)
	if diags.HasError() {
		return CertificateConfig{}, diags
	}

	if existingDN != nil {
		certificateConfig.SubjectDN = buildSubjectDNObject(existingDN)
	}

	return certificateConfig, diag.Diagnostics{}
}

func buildCertificateWithSANConfig(ctx context.Context, value apiobjects.Certificate, existingDN *certificateSubjectDNConfig) (CertificateWithSANConfig, diag.Diagnostics) {
	certificateConfig, diags := buildCertificateModelWithSAN(ctx, value)
	if diags.HasError() {
		return CertificateWithSANConfig{}, diags
	}

	if existingDN != nil {
		certificateConfig.SubjectDN = buildSubjectDNObject(existingDN)
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
func SelfSignedCACertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, existingDN *certificateSubjectDNConfig) (SelfSignedCACertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, existingDN)
	if diags.HasError() {
		return SelfSignedCACertificateResourceConfig{}, diags
	}

	return SelfSignedCACertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

func SignedChainCACertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (SignedChainCACertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, nil)
	if diags.HasError() {
		return SignedChainCACertificateResourceConfig{}, diags
	}

	return SignedChainCACertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

func PKCS12CACertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (PKCS12CACertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateWithSANConfig(ctx, value, nil)
	if diags.HasError() {
		return PKCS12CACertificateResourceConfig{}, diags
	}

	return PKCS12CACertificateResourceConfig{
		CertificateWithSANConfig: config,
	}, diag.Diagnostics{}
}

// System Certificates
func SelfSignedSystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate, existingDN *certificateSubjectDNConfig) (SelfSignedSystemCertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateConfig(ctx, value, existingDN)
	if diags.HasError() {
		return SelfSignedSystemCertificateResourceConfig{}, diags
	}

	return SelfSignedSystemCertificateResourceConfig{
		CertificateConfig: config,
	}, diag.Diagnostics{}
}

func SignedChainSystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (SignedChainSystemCertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateConfig(ctx, value, nil)
	if diags.HasError() {
		return SignedChainSystemCertificateResourceConfig{}, diags
	}

	return SignedChainSystemCertificateResourceConfig{
		CertificateConfig: config,
	}, diag.Diagnostics{}
}

func PKCS12SystemCertificateResourceValueFrom(ctx context.Context, value apiobjects.Certificate) (PKCS12SystemCertificateResourceConfig, diag.Diagnostics) {
	config, diags := buildCertificateConfig(ctx, value, nil)
	if diags.HasError() {
		return PKCS12SystemCertificateResourceConfig{}, diags
	}

	return PKCS12SystemCertificateResourceConfig{
		CertificateConfig: config,
	}, diag.Diagnostics{}
}
