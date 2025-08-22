package systemMapping

import (
	"testing"
)

func TestValidateProtocolBackend(t *testing.T) {
	tests := []struct {
		name        string
		protocol    string
		backendType string
		expectError bool
	}{
		{"LDAP with nonSAPsys", "LDAP", "nonSAPsys", false},
		{"LDAPS with nonSAPsys", "LDAPS", "nonSAPsys", false},
		{"LDAP with SAPsys", "LDAP", "SAPsys", true},
		{"LDAPS with SAPsys", "LDAPS", "SAPsys", true},
		{"HTTPS with SAPsys", "HTTPS", "SAPsys", false},
		{"RFC with nonSAPsys", "RFC", "nonSAPsys", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateProtocolBackend(tt.protocol, tt.backendType)

			if tt.expectError && !diags.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && diags.HasError() {
				t.Errorf("expected no error but got: %v", diags)
			}
		})
	}
}
