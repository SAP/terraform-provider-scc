package systemMapping

import (
	"testing"
)

func TestValidateAuthenticationMode(t *testing.T) {
	tests := []struct {
		name        string
		protocol    string
		authMode    string
		expectError bool
	}{
		{"HTTP valid NONE", "HTTP", "NONE", false},
		{"HTTP valid KERBEROS", "HTTP", "KERBEROS", false},
		{"HTTP invalid X509", "HTTP", "X509_GENERAL", true},
		{"HTTPS valid NONE", "HTTPS", "NONE", false},
		{"HTTPS valid X509", "HTTPS", "X509_RESTRICTED", false},
		{"HTTPS invalid RFC", "HTTPS", "RFC", true},
		{"RFC valid NONE", "RFC", "NONE", false},
		{"RFC invalid KERBEROS", "RFC", "KERBEROS", true},
		{"Unknown protocol", "FOO", "NONE", false}, // skip validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateAuthenticationMode(tt.protocol, tt.authMode)

			if tt.expectError && !diags.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && diags.HasError() {
				t.Errorf("expected no error but got: %v", diags)
			}
		})
	}
}
