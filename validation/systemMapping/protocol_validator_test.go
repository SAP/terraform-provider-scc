package systemMapping

import "testing"

func TestValidateStringProtocol(t *testing.T) {
	tests := []struct {
		name        string
		allowed     []string
		protocol    string
		expectError bool
	}{
		// HTTP family
		{"valid HTTP protocol", []string{"HTTP", "HTTPS"}, "HTTP", false},
		{"valid HTTPS protocol", []string{"HTTP", "HTTPS"}, "HTTPS", false},
		{"invalid RFC protocol", []string{"HTTP", "HTTPS"}, "RFC", true},

		// RFC family
		{"valid RFC protocol", []string{"RFC", "RFCS"}, "RFC", false},
		{"valid RFCS protocol", []string{"RFC", "RFCS"}, "RFCS", false},
		{"invalid HTTP for RFC family", []string{"RFC", "RFCS"}, "HTTP", true},

		// Single protocol
		{"valid LDAP protocol", []string{"LDAP"}, "LDAP", false},
		{"invalid LDAPS protocol", []string{"LDAP"}, "LDAPS", true},

		// Empty allowed list
		{"empty allowed list", []string{}, "HTTP", true},

		// Unknown protocol
		{"unknown protocol", []string{"HTTP", "HTTPS"}, "FOO", true},
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
