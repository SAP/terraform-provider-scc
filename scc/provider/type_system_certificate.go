package provider

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemCertificateConfig struct {
	SubjectDN       types.String `tfsdk:"subject_dn"`
	Issuer          types.String `tfsdk:"issuer"`
	ValidFrom       types.String `tfsdk:"valid_from"`
	ValidTo         types.String `tfsdk:"valid_to"`
	SerialNumber    types.String `tfsdk:"serial_number"`
	SubjectAltNames types.String `tfsdk:"subject_alternative_names"`
	CertificatePEM  types.String `tfsdk:"certificate_pem"`
}

func SystemCertificateDataSourceValueFrom(ctx context.Context, value apiobjects.SystemCertificate, pemBytes []byte) (SystemCertificateConfig, diag.Diagnostics) {
	subjectAltNames := types.StringNull()

	if value.SubjectAltNames != "" {
		subjectAltNames = types.StringValue(value.SubjectAltNames)
	}

	model := &SystemCertificateConfig{
		ValidTo:         ConvertMillisToTimes(value.NotAfterTimeStamp).WithTimezone,
		ValidFrom:       ConvertMillisToTimes(value.NotBeforeTimeStamp).WithTimezone,
		SubjectDN:       types.StringValue(value.SubjectDN),
		Issuer:          types.StringValue(value.Issuer),
		SerialNumber:    types.StringValue(value.SerialNumber),
		SubjectAltNames: subjectAltNames,
		CertificatePEM:  types.StringValue(string(pemBytes)),
	}
	return *model, diag.Diagnostics{}
}
