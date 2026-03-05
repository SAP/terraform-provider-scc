package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Wrappers for testing purposes (allows mocking in tests)
var pkcs12CertificateResourceValueFromFunc = SystemCertificatePKCS12CertificateResourceValueFrom

type SystemCertificateConfig struct {
	SubjectDN       types.Object `tfsdk:"subject_dn"`
	Issuer          types.String `tfsdk:"issuer"`
	ValidFrom       types.String `tfsdk:"valid_from"`
	ValidTo         types.String `tfsdk:"valid_to"`
	SerialNumber    types.String `tfsdk:"serial_number"`
	SubjectAltNames types.String `tfsdk:"subject_alternative_names"`
	CertificatePEM  types.String `tfsdk:"certificate_pem"`
}

type SystemCertificateSelfSignedResourceConfig struct {
	KeySize        types.Int64  `tfsdk:"key_size"`
	SubjectDN      types.Object `tfsdk:"subject_dn"`
	Issuer         types.String `tfsdk:"issuer"`
	ValidFrom      types.String `tfsdk:"valid_from"`
	ValidTo        types.String `tfsdk:"valid_to"`
	SerialNumber   types.String `tfsdk:"serial_number"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
}

type SystemCertificateSignedChainResourceConfig struct {
	SignedChain     types.String `tfsdk:"signed_chain"`
	SubjectDN       types.Object `tfsdk:"subject_dn"`
	Issuer          types.String `tfsdk:"issuer"`
	ValidFrom       types.String `tfsdk:"valid_from"`
	ValidTo         types.String `tfsdk:"valid_to"`
	SerialNumber    types.String `tfsdk:"serial_number"`
	SubjectAltNames types.String `tfsdk:"subject_alternative_names"`
	CertificatePEM  types.String `tfsdk:"certificate_pem"`
}

type SystemCertificatePKCS12CertificateResourceConfig struct {
	PKCS12Certificate types.String `tfsdk:"pkcs12_certificate"`
	Password          types.String `tfsdk:"password"`
	KeyPassword       types.String `tfsdk:"key_password"`
	SubjectDN         types.Object `tfsdk:"subject_dn"`
	Issuer            types.String `tfsdk:"issuer"`
	ValidFrom         types.String `tfsdk:"valid_from"`
	ValidTo           types.String `tfsdk:"valid_to"`
	SerialNumber      types.String `tfsdk:"serial_number"`
	SubjectAltNames   types.String `tfsdk:"subject_alternative_names"`
	CertificatePEM    types.String `tfsdk:"certificate_pem"`
type CSRActionConfig struct {
	Type                    types.String `tfsdk:"type"`
	KeySize                 types.Int64  `tfsdk:"key_size"`
	SubjectDN               types.Object `tfsdk:"subject_dn"`
	SubjectAlternativeNames types.List   `tfsdk:"subject_alternative_names"`
}

func SystemCertificateDataSourceValueFrom(ctx context.Context, value apiobjects.SystemCertificate, pemBytes []byte) (SystemCertificateConfig, diag.Diagnostics) {
	subjectAltNames := types.StringNull()

	if value.SubjectAltNames != "" {
		subjectAltNames = types.StringValue(value.SubjectAltNames)
	}

	model := &SystemCertificateConfig{
		ValidTo:         ConvertMillisToTimes(value.NotAfterTimeStamp).WithTimezone,
		ValidFrom:       ConvertMillisToTimes(value.NotBeforeTimeStamp).WithTimezone,
		Issuer:          types.StringValue(value.Issuer),
		SerialNumber:    types.StringValue(value.SerialNumber),
		SubjectAltNames: subjectAltNames,
		CertificatePEM:  types.StringValue(string(pemBytes)),
		SubjectDN:       types.ObjectNull(subjectDNAttrTypes.AttrTypes),
	}

	if value.SubjectDN != "" {
		dn := parseSubjectDN(value.SubjectDN)
		model.SubjectDN = buildSubjectDNObject(dn)
	} else {
		model.SubjectDN = types.ObjectNull(subjectDNAttrTypes.AttrTypes)
	}
	return *model, diag.Diagnostics{}
}

func SystemCertificateSelfSignedResourceValueFrom(ctx context.Context, value apiobjects.SystemCertificate, pemBytes []byte, existingDN *CertificateSubjectDNConfig) (SystemCertificateSelfSignedResourceConfig, diag.Diagnostics) {
	model := &SystemCertificateSelfSignedResourceConfig{
		ValidTo:        ConvertMillisToTimes(value.NotAfterTimeStamp).WithTimezone,
		ValidFrom:      ConvertMillisToTimes(value.NotBeforeTimeStamp).WithTimezone,
		Issuer:         types.StringValue(value.Issuer),
		SerialNumber:   types.StringValue(value.SerialNumber),
		CertificatePEM: types.StringValue(string(pemBytes)),
		SubjectDN:      types.ObjectNull(subjectDNAttrTypes.AttrTypes),
	}

	if existingDN != nil {
		model.SubjectDN = buildSubjectDNObject(existingDN)
	} else if value.SubjectDN != "" {
		dn := parseSubjectDN(value.SubjectDN)
		model.SubjectDN = buildSubjectDNObject(dn)
	}
	return *model, diag.Diagnostics{}
}

func SystemCertificateSignedChainResourceValueFrom(ctx context.Context, value apiobjects.SystemCertificate, pemBytes []byte) (SystemCertificateSignedChainResourceConfig, diag.Diagnostics) {
	subjectAltNames := types.StringNull()

	if value.SubjectAltNames != "" {
		subjectAltNames = types.StringValue(value.SubjectAltNames)
	}

	model := &SystemCertificateSignedChainResourceConfig{
		ValidTo:         ConvertMillisToTimes(value.NotAfterTimeStamp).WithTimezone,
		ValidFrom:       ConvertMillisToTimes(value.NotBeforeTimeStamp).WithTimezone,
		Issuer:          types.StringValue(value.Issuer),
		SerialNumber:    types.StringValue(value.SerialNumber),
		SubjectAltNames: subjectAltNames,
		SubjectDN:       types.ObjectNull(subjectDNAttrTypes.AttrTypes),
		CertificatePEM:  types.StringValue(string(pemBytes)),
	}

	if value.SubjectDN != "" {
		dn := parseSubjectDN(value.SubjectDN)
		model.SubjectDN = buildSubjectDNObject(dn)
	} else {
		model.SubjectDN = types.ObjectNull(subjectDNAttrTypes.AttrTypes)
	}

	return *model, diag.Diagnostics{}
}

func SystemCertificatePKCS12CertificateResourceValueFrom(ctx context.Context, value apiobjects.SystemCertificate, pemBytes []byte) (SystemCertificatePKCS12CertificateResourceConfig, diag.Diagnostics) {
	subjectAltNames := types.StringNull()

	if value.SubjectAltNames != "" {
		subjectAltNames = types.StringValue(value.SubjectAltNames)
	}

	model := &SystemCertificatePKCS12CertificateResourceConfig{
		ValidTo:         ConvertMillisToTimes(value.NotAfterTimeStamp).WithTimezone,
		ValidFrom:       ConvertMillisToTimes(value.NotBeforeTimeStamp).WithTimezone,
		Issuer:          types.StringValue(value.Issuer),
		SerialNumber:    types.StringValue(value.SerialNumber),
		SubjectAltNames: subjectAltNames,
		SubjectDN:       types.ObjectNull(subjectDNAttrTypes.AttrTypes),
		CertificatePEM:  types.StringValue(string(pemBytes)),
	}

	if value.SubjectDN != "" {
		dn := parseSubjectDN(value.SubjectDN)
		model.SubjectDN = buildSubjectDNObject(dn)
	} else {
		model.SubjectDN = types.ObjectNull(subjectDNAttrTypes.AttrTypes)
	}

	return *model, diag.Diagnostics{}
}

func buildSubjectDNObject(dn *CertificateSubjectDNConfig) types.Object {
	if dn == nil {
		return types.ObjectNull(subjectDNAttrTypes.AttrTypes)
	}

	attrs := map[string]attr.Value{
		"cn":    dn.CommonName,
		"email": dn.Email,
		"l":     dn.Locality,
		"ou":    dn.OrganizationalUnit,
		"o":     dn.Organization,
		"st":    dn.State,
		"c":     dn.Country,
	}

	obj, diags := types.ObjectValue(subjectDNAttrTypes.AttrTypes, attrs)
	if diags.HasError() {
		panic("failed to build subject_dn object")
	}
	return obj
}
