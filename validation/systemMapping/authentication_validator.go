package systemMapping

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func validateAuthenticationMode(protocol, authMode string) diag.Diagnostics {
	var diags diag.Diagnostics

	allowed := map[string][]string{
		"HTTP":  {"NONE", "KERBEROS"},
		"HTTPS": {"NONE", "X509_GENERAL", "X509_RESTRICTED", "KERBEROS"},
		"RFC":   {"NONE"},
		"RFCS":  {"NONE", "X509_GENERAL", "X509_RESTRICTED"},
		"LDAP":  {"NONE"},
		"LDAPS": {"NONE"},
		"TCP":   {"NONE"},
		"TCPS":  {"NONE"},
	}

	validValues, ok := allowed[protocol]
	if !ok {
		return diags
	}

	isValid := slices.Contains(validValues, authMode)

	if !isValid {
		diags.AddAttributeError(
			path.Empty(),
			"Invalid authentication_mode for protocol",
			fmt.Sprintf("authentication_mode %q is not valid for protocol %q. Allowed values: %v", authMode, protocol, validValues),
		)
	}

	return diags
}

type ProtocolAuthenticationModeValidator struct{}

func (v ProtocolAuthenticationModeValidator) Description(_ context.Context) string {
	return "Ensures authentication_mode is valid for the chosen protocol."
}
func (v ProtocolAuthenticationModeValidator) MarkdownDescription(_ context.Context) string {
	return "Validates that **authentication_mode** matches the allowed values for the configured **protocol**."
}
func (v ProtocolAuthenticationModeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var protocol string
	diags := req.Config.GetAttribute(ctx, path.Root("protocol"), &protocol)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateAuthenticationMode(protocol, req.ConfigValue.ValueString())...)
}

func ValidateAuthenticationMode() validator.String {
	return ProtocolAuthenticationModeValidator{}
}
