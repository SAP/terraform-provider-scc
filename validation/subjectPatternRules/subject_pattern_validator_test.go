package subjectpatternrules

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var subjectPatternAttrTypes = map[string]attr.Type{
	"cn":    types.StringType,
	"email": types.StringType,
	"l":     types.StringType,
	"ou":    types.StringType,
	"o":     types.StringType,
	"st":    types.StringType,
	"c":     types.StringType,
}

func makeSubjectPatternObject(t *testing.T, cn, email, l, ou, o, st, c string) types.Object {
	t.Helper()
	toVal := func(s string) attr.Value {
		if s == "" {
			return types.StringNull()
		}
		return types.StringValue(s)
	}
	obj, diags := types.ObjectValue(subjectPatternAttrTypes, map[string]attr.Value{
		"cn":    toVal(cn),
		"email": toVal(email),
		"l":     toVal(l),
		"ou":    toVal(ou),
		"o":     toVal(o),
		"st":    toVal(st),
		"c":     toVal(c),
	})
	if diags.HasError() {
		t.Fatalf("failed to build subject pattern object: %v", diags)
	}
	return obj
}

func runSubjectPatternValidator(t *testing.T, obj types.Object) validator.ObjectResponse {
	t.Helper()
	req := validator.ObjectRequest{ConfigValue: obj}
	resp := &validator.ObjectResponse{}
	SubjectPatternAtLeastOneField().ValidateObject(context.Background(), req, resp)
	return *resp
}

func TestSubjectPatternValidator(t *testing.T) {
	tests := []struct {
		name                       string
		cn, email, l, ou, o, st, c string
		expectError                bool
	}{
		// Valid: single field set
		{"only cn", "John", "", "", "", "", "", "", false},
		{"only email", "", "user@example.com", "", "", "", "", "", false},
		{"only l", "", "", "Berlin", "", "", "", "", false},
		{"only ou", "", "", "", "Engineering", "", "", "", false},
		{"only o", "", "", "", "", "ACME", "", "", false},
		{"only st", "", "", "", "", "", "BE", "", false},
		{"only c", "", "", "", "", "", "", "DE", false},

		// Valid: multiple fields set
		{"cn and o", "John", "", "", "", "ACME", "", "", false},
		{"cn, o, c", "John", "", "", "", "ACME", "", "DE", false},
		{"all fields", "John", "j@x.com", "Berlin", "Eng", "ACME", "BE", "DE", false},

		// Invalid: all fields empty
		{"all empty - invalid", "", "", "", "", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := makeSubjectPatternObject(t, tt.cn, tt.email, tt.l, tt.ou, tt.o, tt.st, tt.c)
			resp := runSubjectPatternValidator(t, obj)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("expected no error but got: %v", resp.Diagnostics)
			}
			if tt.expectError && resp.Diagnostics.HasError() {
				found := false
				for _, d := range resp.Diagnostics {
					if d.Summary() == "Empty subject pattern" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected 'Empty subject pattern' diagnostic, got: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestSubjectPatternValidator_NullValue(t *testing.T) {
	obj := types.ObjectNull(subjectPatternAttrTypes)
	resp := runSubjectPatternValidator(t, obj)
	if resp.Diagnostics.HasError() {
		t.Errorf("null object should be skipped, got: %v", resp.Diagnostics)
	}
}

func TestSubjectPatternValidator_Description(t *testing.T) {
	v := SubjectPatternAtLeastOneField()
	if v.Description(context.Background()) == "" {
		t.Error("Description should not be empty")
	}
	if v.MarkdownDescription(context.Background()) == "" {
		t.Error("MarkdownDescription should not be empty")
	}
}
