package uuidvalidator

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// UuidRegexp matches:
// - Standard UUID format: 8-4-4-4-12 (hex with hyphens)
// - SAP subaccount ID format: [a-z] + 8 hex characters (e.g. xf014edd7)
var UuidRegexp = regexp.MustCompile(
	`^(?:` +
		`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}` +
		`|` +
		`[a-z][0-9a-fA-F]{8}` +
	`)$`,
)

// ValidUUID validates that the string attribute value is either
// a standard UUID or a SAP subaccount ID
func ValidUUID() validator.String {
	return stringvalidator.RegexMatches(
		UuidRegexp,
		"value must be a valid UUID or SAP subaccount id ([a-z]########)",
	)
}
