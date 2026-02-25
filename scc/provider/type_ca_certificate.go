package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CACertificateConfig struct {
	SubjectDN       types.Object `tfsdk:"subject_dn"`
	Issuer          types.String `tfsdk:"issuer"`
	ValidFrom       types.String `tfsdk:"valid_from"`
	ValidTo         types.String `tfsdk:"valid_to"`
	SerialNumber    types.String `tfsdk:"serial_number"`
	SubjectAltNames types.String `tfsdk:"subject_alternative_names"`
	CertificatePEM  types.String `tfsdk:"certificate_pem"`
}

func CACertificateDataSourceValueFrom(ctx context.Context, value apiobjects.CACertificate, pemBytes []byte) (CACertificateConfig, diag.Diagnostics) {
	subjectAltNames := types.StringNull()

	if value.SubjectAltNames != "" {
		subjectAltNames = types.StringValue(value.SubjectAltNames)
	}

	model := &CACertificateConfig{
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
