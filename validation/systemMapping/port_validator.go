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

var reservedSIDs = map[string]struct{}{
	"ADD": {}, "ALL": {}, "AND": {}, "ANY": {}, "ASC": {}, "AUX": {}, "COM": {}, "CON": {}, "DBA": {}, "END": {},
	"EPS": {}, "FOR": {}, "GID": {}, "IBM": {}, "INT": {}, "KEY": {}, "LOG": {}, "LPT": {}, "MON": {}, "NIX": {},
	"NOT": {}, "NUL": {}, "OFF": {}, "OMS": {}, "PRN": {}, "RAW": {}, "ROW": {}, "SAP": {}, "SET": {}, "SGA": {},
	"SHG": {}, "SID": {}, "SQL": {}, "SYS": {}, "TMP": {}, "TOP": {}, "UID": {}, "USE": {}, "USR": {}, "VAR": {},
}

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
				fmt.Sprintf("Value %q is not valid for RFC. Allowed: numeric 1–65535, sapmsSID (valid 3-char SID), or sapgwXX (00–99) (or numeric 100–65535 without LB)."+
					"Ensure the SID is not one of the reserved SIDs (e.g., SAP, SYS, ALL, COM, etc.).", portValue),
			)
		}

	case "RFCS":
		// With LB: port 1–65535 OR sapmsSID
		// Without LB: sapgwXXs OR port 100–65535
		if !ValidateRFCSValue(portValue) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid RFCS Value",
				fmt.Sprintf("Value %q is not valid for RFCS. Allowed: numeric 1–65535, sapmsSID (valid 3-char SID), or sapgwXXs (00–99) (or numeric 100–65535 without LB)."+
					"Ensure the SID is not one of the reserved SIDs (e.g., SAP, SYS, ALL, COM, etc.).", portValue),
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

func isValidSID(sid string) bool {
	// Must be exactly 3 chars
	if len(sid) != 3 {
		return false
	}
	// Only uppercase alphanumeric
	if !regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(sid) {
		return false
	}
	// First char must be a letter
	if !regexp.MustCompile(`^[A-Z]`).MatchString(string(sid[0])) {
		return false
	}
	// Reserved check
	if _, reserved := reservedSIDs[sid]; reserved {
		return false
	}
	return true
}

func isValidSapms(value string) bool {
	re := regexp.MustCompile(`^sapms([A-Z0-9]{3})$`)
	m := re.FindStringSubmatch(value)
	if m == nil {
		return false
	}
	return isValidSID(m[1])
}

func ValidateHTTPPort(value string) bool {
	return isNumericInRange(value, 1, 65535)
}

func ValidateRFCValue(value string) bool {
	return isNumericInRange(value, 1, 65535) ||
		regexp.MustCompile(`^sapgw[0-9]{2}$`).MatchString(value) ||
		isValidSapms(value)
}

func ValidateRFCSValue(value string) bool {
	return isNumericInRange(value, 1, 65535) ||
		regexp.MustCompile(`^sapgw[0-9]{2}s$`).MatchString(value) ||
		isValidSapms(value)
}

func ValidatePort() validator.String {
	return PortValidator{}
}
