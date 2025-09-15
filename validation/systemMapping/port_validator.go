package systemMapping

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Wrapper ---
type PortValidator struct{}

func (v PortValidator) Description(_ context.Context) string {
	return "Validates internal_port/virtual_port based on protocol"
}

func (v PortValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v PortValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var protocol types.String
	diags := req.Config.GetAttribute(ctx, path.Root("protocol"), &protocol)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	portValue := req.ConfigValue.ValueString()
	protocolValue := protocol.ValueString()

	switch protocolValue {
	case "HTTP", "HTTPS", "TCP", "TCPS", "LDAP", "LDAPS":
		// Only numeric 1–65535
		if !isNumericInRange(portValue, 1, 65535) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Port",
				fmt.Sprintf("Value %q is not valid for protocol %q. Must be numeric between 1 and 65535.", portValue, protocolValue),
			)
		}

	case "RFC":
		// With LB: port 1–65535 OR sapmsSID
		// Without LB: sapgwXX OR port 100–65535
		if !ValidateRFCValue(portValue) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid RFC Value",
				fmt.Sprintf("Value %q is not valid for RFC. Allowed: numeric 1–65535, sapmsSID, or sapgwXX (or numeric 100–65535 without LB).", portValue),
			)
		}

	case "RFCS":
		// With LB: port 1–65535 OR sapmsSID
		// Without LB: sapgwXXs OR port 100–65535
		if !ValidateRFCSValue(portValue) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid RFCS Value",
				fmt.Sprintf("Value %q is not valid for RFCS. Allowed: numeric 1–65535, sapmsSID, or sapgwXXs (or numeric 100–65535 without LB).", portValue),
			)
		}
	}
}

func isNumericInRange(s string, min, max int) bool {
	if !regexp.MustCompile(`^[0-9]{1,5}$`).MatchString(s) {
		return false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return n >= min && n <= max
}

func ValidateHTTPPort(value string) bool {
	return isNumericInRange(value, 1, 65535)
}

func ValidateRFCValue(value string) bool {
	return isNumericInRange(value, 1, 65535) ||
		regexp.MustCompile(`^sapms[A-Z0-9]{3}$`).MatchString(value) ||
		regexp.MustCompile(`^sapgw[0-9]{2}$`).MatchString(value)
}

func ValidateRFCSValue(value string) bool {
	return isNumericInRange(value, 1, 65535) ||
		regexp.MustCompile(`^sapms[A-Z0-9]{3}$`).MatchString(value) ||
		regexp.MustCompile(`^sapgw[0-9]{2}s$`).MatchString(value)
}

func ValidatePort() validator.String {
	return PortValidator{}
}
