package systemMapping

import (
	"testing"
)

func TestValidateProtocolList(t *testing.T) {
	tests := []struct {
		name        string
		allowed     []string
		protocol    string
		expectError bool
	}{
		{"valid RFC protocol", []string{"RFC", "RFCS"}, "RFC", false},
		{"valid RFCS protocol", []string{"RFC", "RFCS"}, "RFCS", false},
		{"invalid HTTP protocol", []string{"RFC", "RFCS"}, "HTTP", true},
		{"invalid HTTPS protocol", []string{"RFC", "RFCS"}, "HTTPS", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateListProtocol(tt.protocol, tt.allowed)

			if tt.expectError && !diags.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && diags.HasError() {
				t.Errorf("expected no error but got: %v", diags)
			}
		})
	}
}
