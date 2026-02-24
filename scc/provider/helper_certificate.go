package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type CertificateSubjectDNConfig struct {
	CommonName         types.String `tfsdk:"cn"`
	Email              types.String `tfsdk:"email"`
	Locality           types.String `tfsdk:"l"`
	OrganizationalUnit types.String `tfsdk:"ou"`
	Organization       types.String `tfsdk:"o"`
	State              types.String `tfsdk:"st"`
	Country            types.String `tfsdk:"c"`
}

var subjectDNAttrTypes = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"cn":    types.StringType,
		"email": types.StringType,
		"l":     types.StringType,
		"ou":    types.StringType,
		"o":     types.StringType,
		"st":    types.StringType,
		"c":     types.StringType,
	},
}

func ExpandSubjectDN(ctx context.Context, subjectDN types.Object) (*CertificateSubjectDNConfig, diag.Diagnostics) {
	if subjectDN.IsNull() || subjectDN.IsUnknown() {
		return nil, diag.Diagnostics{}
	}

	var result CertificateSubjectDNConfig
	diags := subjectDN.As(ctx, &result, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &result, diags
}

func BuildSubjectDN(subjectDN *CertificateSubjectDNConfig) string {
	if subjectDN == nil || subjectDN.CommonName.IsNull() {
		return ""
	}

	clean := func(value string) string {
		return strings.TrimSpace(value)
	}
	parts := []string{
		fmt.Sprintf("CN=%s", clean(subjectDN.CommonName.ValueString())),
	}

	if !subjectDN.Email.IsNull() && strings.TrimSpace(subjectDN.Email.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("EMAIL=%s", clean(subjectDN.Email.ValueString())))
	}
	if !subjectDN.Locality.IsNull() && strings.TrimSpace(subjectDN.Locality.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("L=%s", clean(subjectDN.Locality.ValueString())))
	}
	if !subjectDN.OrganizationalUnit.IsNull() && strings.TrimSpace(subjectDN.OrganizationalUnit.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("OU=%s", clean(subjectDN.OrganizationalUnit.ValueString())))
	}
	if !subjectDN.Organization.IsNull() && strings.TrimSpace(subjectDN.Organization.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("O=%s", clean(subjectDN.Organization.ValueString())))
	}
	if !subjectDN.State.IsNull() && strings.TrimSpace(subjectDN.State.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("ST=%s", clean(subjectDN.State.ValueString())))
	}
	if !subjectDN.Country.IsNull() && strings.TrimSpace(subjectDN.Country.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("C=%s", clean(subjectDN.Country.ValueString())))
	}
	return strings.Join(parts, ",")
}

func parseSubjectDN(dn string) *CertificateSubjectDNConfig {
	result := &CertificateSubjectDNConfig{
		CommonName:         types.StringNull(),
		Email:              types.StringNull(),
		Locality:           types.StringNull(),
		OrganizationalUnit: types.StringNull(),
		Organization:       types.StringNull(),
		State:              types.StringNull(),
		Country:            types.StringNull(),
	}

	for part := range strings.SplitSeq(dn, ",") {
		part = strings.TrimSpace(part)

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.ToUpper(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])

		switch key {
		case "CN":
			result.CommonName = types.StringValue(val)
		case "EMAIL":
			result.Email = types.StringValue(val)
		case "L":
			result.Locality = types.StringValue(val)
		case "OU":
			result.OrganizationalUnit = types.StringValue(val)
		case "O":
			result.Organization = types.StringValue(val)
		case "ST":
			result.State = types.StringValue(val)
		case "C":
			result.Country = types.StringValue(val)
		}
	}

	return result
}
