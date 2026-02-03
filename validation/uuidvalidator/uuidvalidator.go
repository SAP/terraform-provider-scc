package uuidvalidator

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	uuidRegexp = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`

	uuidCompiledRegexp = regexp.MustCompile(`^` + uuidRegexp + `$`)

	// SAP BTP short subaccount ID
	// Exactly 8 lowercase alphanumeric characters, starting with a letter
	subaccountShortRegexp = `[a-z][a-z0-9]{7}`

	// Accepts either a UUID or a SAP BTP short subaccount ID
	validSubaccountRegexp = regexp.MustCompile(
		`^(` + uuidRegexp + `|` + subaccountShortRegexp + `)$`,
	)
)

// ValidUUID validates that the value is a UUID
func ValidUUID() validator.String {
	return stringvalidator.RegexMatches(
		uuidCompiledRegexp,
		"value must be a valid UUID",
	)
}

// ValidSubaccountID validates that the value is either a UUID or a SAP BTP short subaccount ID
func ValidSubaccountID() validator.String {
	return stringvalidator.RegexMatches(
		validSubaccountRegexp,
		"value must be a valid UUID or a valid SAP BTP short subaccount ID",
	)
}
