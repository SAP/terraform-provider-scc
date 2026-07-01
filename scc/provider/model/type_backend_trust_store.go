package model

import (
	"context"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BackendTrustStoreActionConfig struct {
	TrustAllBackends types.Bool `tfsdk:"trust_all_backends"`
}

type BackendTrustStoreDataSourceConfig struct {
	TrustAllBackends types.Bool `tfsdk:"trust_all_backends"`
	TrustedBackends  types.List `tfsdk:"trusted_backends"`
}

type TrustedBackendsData struct {
	Alias     types.String `tfsdk:"alias"`
	SubjectDN types.Object `tfsdk:"subject_dn"`
	Issuer    types.String `tfsdk:"issuer"`
	ValidTo   types.String `tfsdk:"valid_to"`
}

type BackendTrustStoreResourceConfig struct {
	Certificate types.String `tfsdk:"certificate"`
	Alias       types.String `tfsdk:"alias"`
	SubjectDN   types.Object `tfsdk:"subject_dn"`
	Issuer      types.String `tfsdk:"issuer"`
	ValidTo     types.String `tfsdk:"valid_to"`
}

var TrustedBackendsDataAttrTypes = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"alias":      types.StringType,
		"subject_dn": helpers.SubjectDNAttrTypes,
		"issuer":     types.StringType,
		"valid_to":   types.StringType,
	},
}

func BackendTrustStoreDataSourceValueFrom(ctx context.Context, value apiobjects.BackendTrustStoreConfiguration) (BackendTrustStoreDataSourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	trustedBackendsValue := []TrustedBackendsData{}
	for _, trustedBackend := range value.TrustedBackends {
		dn := helpers.ParseSubjectDNFunc(trustedBackend.SubjectDN)

		trustedBackendsValue = append(trustedBackendsValue, TrustedBackendsData{
			Alias:     types.StringValue(trustedBackend.Alias),
			SubjectDN: helpers.BuildSubjectDNObjectFunc(dn),
			Issuer:    types.StringValue(trustedBackend.Issuer),
			ValidTo:   helpers.ConvertMillisToTimes(trustedBackend.ValidTo).WithTimezone,
		})
	}

	trustBackends, diags := types.ListValueFrom(ctx, TrustedBackendsDataAttrTypes, trustedBackendsValue)
	if diags.HasError() {
		return BackendTrustStoreDataSourceConfig{}, diags
	}

	return BackendTrustStoreDataSourceConfig{
		TrustAllBackends: types.BoolValue(value.TrustAllBackends),
		TrustedBackends:  trustBackends,
	}, diags
}

func BackendTrustStoreResourceValueFrom(certificate string, value apiobjects.TrustedBackends) (BackendTrustStoreResourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	dn := helpers.ParseSubjectDNFunc(value.SubjectDN)

	return BackendTrustStoreResourceConfig{
		Certificate: types.StringValue(certificate),
		Alias:       types.StringValue(value.Alias),
		SubjectDN:   helpers.BuildSubjectDNObjectFunc(dn),
		Issuer:      types.StringValue(value.Issuer),
		ValidTo:     helpers.ConvertMillisToTimes(value.ValidTo).WithTimezone,
	}, diags
}
