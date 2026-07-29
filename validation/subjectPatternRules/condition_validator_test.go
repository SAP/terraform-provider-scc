package subjectpatternrules

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// conditionAttrTypes mirrors the condition object shape expected by the validator.
var conditionAttrTypes = map[string]attr.Type{
	"variable": types.StringType,
	"operator": types.StringType,
	"value":    types.StringType,
}

// makeConditionObject builds a types.Object for a condition, omitting value when empty.
func makeConditionObject(t *testing.T, variable, operator, value string) types.Object {
	t.Helper()
	attrs := map[string]attr.Value{
		"variable": types.StringValue(variable),
		"operator": types.StringValue(operator),
		"value":    types.StringNull(),
	}
	if value != "" {
		attrs["value"] = types.StringValue(value)
	}
	obj, diags := types.ObjectValue(conditionAttrTypes, attrs)
	if diags.HasError() {
		t.Fatalf("failed to build condition object: %v", diags)
	}
	return obj
}

func runValidator(t *testing.T, obj types.Object) validator.ObjectResponse {
	t.Helper()
	req := validator.ObjectRequest{ConfigValue: obj}
	resp := &validator.ObjectResponse{}
	SubjectPatternRuleCondition().ValidateObject(context.Background(), req, resp)
	return *resp
}

func TestConditionValidator(t *testing.T) {
	tests := []struct {
		name         string
		variable     string
		operator     string
		value        string
		expectError  bool
		errorSummary string
	}{
		// ---------------------------------------------------------------
		// Valid: is / is_not with value
		// ---------------------------------------------------------------
		{
			name:        "name is value - valid",
			variable:    "name",
			operator:    "is",
			value:       "john",
			expectError: false,
		},
		{
			name:        "name is_not value - valid",
			variable:    "name",
			operator:    "is_not",
			value:       "john",
			expectError: false,
		},
		{
			name:        "mail is value - valid",
			variable:    "mail",
			operator:    "is",
			value:       "user@example.com",
			expectError: false,
		},
		{
			name:        "email is_not value - valid",
			variable:    "email",
			operator:    "is_not",
			value:       "admin@example.com",
			expectError: false,
		},
		{
			name:        "display_name is value - valid",
			variable:    "display_name",
			operator:    "is",
			value:       "John Doe",
			expectError: false,
		},
		{
			name:        "login_name is_not value - valid",
			variable:    "login_name",
			operator:    "is_not",
			value:       "jdoe",
			expectError: false,
		},
		{
			name:        "user_type is Business - valid",
			variable:    "user_type",
			operator:    "is",
			value:       "Business",
			expectError: false,
		},
		{
			name:        "user_type is Technical - valid",
			variable:    "user_type",
			operator:    "is",
			value:       "Technical",
			expectError: false,
		},
		{
			name:        "user_type is_not Business - valid",
			variable:    "user_type",
			operator:    "is_not",
			value:       "Business",
			expectError: false,
		},
		{
			name:        "user_type is_not Technical - valid",
			variable:    "user_type",
			operator:    "is_not",
			value:       "Technical",
			expectError: false,
		},
		// ---------------------------------------------------------------
		// Valid: exist / exist_not without value
		// ---------------------------------------------------------------
		{
			name:        "name exist - valid",
			variable:    "name",
			operator:    "exist",
			value:       "",
			expectError: false,
		},
		{
			name:        "name exist_not - valid",
			variable:    "name",
			operator:    "exist_not",
			value:       "",
			expectError: false,
		},
		{
			name:        "mail exist - valid",
			variable:    "mail",
			operator:    "exist",
			value:       "",
			expectError: false,
		},
		{
			name:        "email exist_not - valid",
			variable:    "email",
			operator:    "exist_not",
			value:       "",
			expectError: false,
		},
		{
			name:        "display_name exist - valid",
			variable:    "display_name",
			operator:    "exist",
			value:       "",
			expectError: false,
		},
		{
			name:        "login_name exist_not - valid",
			variable:    "login_name",
			operator:    "exist_not",
			value:       "",
			expectError: false,
		},
		// ---------------------------------------------------------------
		// Invalid: user_type with exist / exist_not
		// ---------------------------------------------------------------
		{
			name:         "user_type exist - invalid operator",
			variable:     "user_type",
			operator:     "exist",
			value:        "",
			expectError:  true,
			errorSummary: "Invalid operator",
		},
		{
			name:         "user_type exist_not - invalid operator",
			variable:     "user_type",
			operator:     "exist_not",
			value:        "",
			expectError:  true,
			errorSummary: "Invalid operator",
		},
		// ---------------------------------------------------------------
		// Invalid: is / is_not without value
		// ---------------------------------------------------------------
		{
			name:         "name is without value",
			variable:     "name",
			operator:     "is",
			value:        "",
			expectError:  true,
			errorSummary: "Missing value",
		},
		{
			name:         "name is_not without value",
			variable:     "name",
			operator:     "is_not",
			value:        "",
			expectError:  true,
			errorSummary: "Missing value",
		},
		{
			name:         "user_type is without value",
			variable:     "user_type",
			operator:     "is",
			value:        "",
			expectError:  true,
			errorSummary: "Missing value",
		},
		// ---------------------------------------------------------------
		// Invalid: exist / exist_not with value
		// ---------------------------------------------------------------
		{
			name:         "name exist with value",
			variable:     "name",
			operator:     "exist",
			value:        "unexpected",
			expectError:  true,
			errorSummary: "Unexpected value",
		},
		{
			name:         "name exist_not with value",
			variable:     "name",
			operator:     "exist_not",
			value:        "unexpected",
			expectError:  true,
			errorSummary: "Unexpected value",
		},
		// ---------------------------------------------------------------
		// Invalid: user_type is / is_not with wrong value
		// ---------------------------------------------------------------
		{
			name:         "user_type is Guest",
			variable:     "user_type",
			operator:     "is",
			value:        "Guest",
			expectError:  true,
			errorSummary: "Invalid value",
		},
		{
			name:         "user_type is Admin",
			variable:     "user_type",
			operator:     "is",
			value:        "Admin",
			expectError:  true,
			errorSummary: "Invalid value",
		},
		{
			name:         "user_type is_not Guest",
			variable:     "user_type",
			operator:     "is_not",
			value:        "Guest",
			expectError:  true,
			errorSummary: "Invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := makeConditionObject(t, tt.variable, tt.operator, tt.value)
			resp := runValidator(t, obj)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error but got none")
				return
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("expected no error but got: %v", resp.Diagnostics)
				return
			}
			if tt.expectError && tt.errorSummary != "" {
				found := false
				for _, d := range resp.Diagnostics {
					if d.Summary() == tt.errorSummary {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected diagnostic with summary %q, got: %v", tt.errorSummary, resp.Diagnostics)
				}
			}
		})
	}
}

func TestConditionValidator_NullValue(t *testing.T) {
	obj := types.ObjectNull(conditionAttrTypes)
	resp := runValidator(t, obj)
	if resp.Diagnostics.HasError() {
		t.Errorf("null object should be skipped, got: %v", resp.Diagnostics)
	}
}

func TestConditionValidator_Description(t *testing.T) {
	v := SubjectPatternRuleCondition()
	if v.Description(context.Background()) == "" {
		t.Error("Description should not be empty")
	}
	if v.MarkdownDescription(context.Background()) == "" {
		t.Error("MarkdownDescription should not be empty")
	}
}
