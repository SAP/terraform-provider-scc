package helpers

import (
	"fmt"
	"strings"
)

func GetFormattedValueAsTableRow(val string, description string) string {
	return fmt.Sprintf("\n  | %s | %s | ", strings.ReplaceAll(val, "|", "\\|"), strings.ReplaceAll(description, "|", "\\|"))
}
