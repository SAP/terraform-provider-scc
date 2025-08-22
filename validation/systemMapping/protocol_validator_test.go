package systemMapping

import (
	"testing"
)

func TestValidateProtocol(t *testing.T) {
	tests := []struct {
		name        string
		allowed     []string
		protocol    string
		expectError bool
	}{
		{"valid protocol", []string{"HTTP", "HTTPS"}, "HTTP", false},
		{"another valid protocol", []string{"HTTP", "HTTPS"}, "HTTPS", false},
		{"invalid protocol", []string{"HTTP", "HTTPS"}, "RFC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateStringProtocol(tt.protocol, tt.allowed)

			if tt.expectError && !diags.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && diags.HasError() {
				t.Errorf("expected no error but got: %v", diags)
			}
		})
	}
}
