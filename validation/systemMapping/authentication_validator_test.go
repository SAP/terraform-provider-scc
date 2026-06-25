package systemMapping

import "testing"

func TestValidateAuthenticationMode(t *testing.T) {
	tests := []struct {
		name        string
		protocol    string
		authMode    string
		expectError bool
	}{
		// HTTP
		{"HTTP valid NONE", "HTTP", "NONE", false},
		{"HTTP valid KERBEROS", "HTTP", "KERBEROS", false},
		{"HTTP invalid X509_GENERAL", "HTTP", "X509_GENERAL", true},
		{"HTTP invalid X509_RESTRICTED", "HTTP", "X509_RESTRICTED", true},

		// HTTPS
		{"HTTPS valid NONE", "HTTPS", "NONE", false},
		{"HTTPS valid NONE_RESTRICTED", "HTTPS", "NONE_RESTRICTED", false},
		{"HTTPS valid X509_GENERAL", "HTTPS", "X509_GENERAL", false},
		{"HTTPS valid X509_RESTRICTED", "HTTPS", "X509_RESTRICTED", false},
		{"HTTPS valid KERBEROS", "HTTPS", "KERBEROS", false},
		{"HTTPS invalid RFC", "HTTPS", "RFC", true},

		// RFC
		{"RFC valid NONE", "RFC", "NONE", false},
		{"RFC invalid KERBEROS", "RFC", "KERBEROS", true},
		{"RFC invalid X509_GENERAL", "RFC", "X509_GENERAL", true},

		// RFCS
		{"RFCS valid NONE", "RFCS", "NONE", false},
		{"RFCS valid X509_GENERAL", "RFCS", "X509_GENERAL", false},
		{"RFCS valid X509_RESTRICTED", "RFCS", "X509_RESTRICTED", false},
		{"RFCS invalid KERBEROS", "RFCS", "KERBEROS", true},
		{"RFCS invalid NONE_RESTRICTED", "RFCS", "NONE_RESTRICTED", true},

		// RFCWS
		{"RFCWS valid NONE", "RFCWS", "NONE", false},
		{"RFCWS valid X509_GENERAL", "RFCWS", "X509_GENERAL", false},
		{"RFCWS invalid KERBEROS", "RFCWS", "KERBEROS", true},
		{"RFCWS invalid X509_RESTRICTED", "RFCWS", "X509_RESTRICTED", true},

		// LDAP
		{"LDAP valid NONE", "LDAP", "NONE", false},
		{"LDAP invalid KERBEROS", "LDAP", "KERBEROS", true},
		{"LDAP invalid X509_GENERAL", "LDAP", "X509_GENERAL", true},

		// LDAPS
		{"LDAPS valid NONE", "LDAPS", "NONE", false},
		{"LDAPS invalid KERBEROS", "LDAPS", "KERBEROS", true},
		{"LDAPS invalid X509_GENERAL", "LDAPS", "X509_GENERAL", true},

		// TCP
		{"TCP valid NONE", "TCP", "NONE", false},
		{"TCP invalid KERBEROS", "TCP", "KERBEROS", true},
		{"TCP invalid X509_GENERAL", "TCP", "X509_GENERAL", true},

		// TCPS
		{"TCPS valid NONE", "TCPS", "NONE", false},
		{"TCPS invalid KERBEROS", "TCPS", "KERBEROS", true},
		{"TCPS invalid X509_GENERAL", "TCPS", "X509_GENERAL", true},

		// Unknown protocol
		{"Unknown protocol", "FOO", "NONE", false},
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
