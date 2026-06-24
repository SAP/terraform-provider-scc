package systemMapping

import "testing"

func TestValidateProtocolBackend(t *testing.T) {
	tests := []struct {
		name        string
		protocol    string
		backendType string
		expectError bool
	}{
		// HTTP
		{"HTTP with abapSys", "HTTP", "abapSys", false},
		{"HTTP with BC", "HTTP", "BC", false},
		{"HTTP with nonSAPsys", "HTTP", "nonSAPsys", false},

		// HTTPS
		{"HTTPS with hana", "HTTPS", "hana", false},
		{"HTTPS with PI", "HTTPS", "PI", false},

		// RFC family
		{"RFC with abapSys", "RFC", "abapSys", false},
		{"RFC with netweaverGW", "RFC", "netweaverGW", false},
		{"RFC with hana", "RFC", "hana", true},
		{"RFC with nonSAPsys", "RFC", "nonSAPsys", true},

		{"RFCS with abapSys", "RFCS", "abapSys", false},
		{"RFCS with netweaverGW", "RFCS", "netweaverGW", false},
		{"RFCS with BC", "RFCS", "BC", true},

		{"RFCWS with abapSys", "RFCWS", "abapSys", false},
		{"RFCWS with netweaverGW", "RFCWS", "netweaverGW", false},
		{"RFCWS with PI", "RFCWS", "PI", true},

		// LDAP family
		{"LDAP with nonSAPsys", "LDAP", "nonSAPsys", false},
		{"LDAP with abapSys", "LDAP", "abapSys", true},

		{"LDAPS with nonSAPsys", "LDAPS", "nonSAPsys", false},
		{"LDAPS with hana", "LDAPS", "hana", true},

		// TCP family
		{"TCP with hana", "TCP", "hana", false},
		{"TCP with otherSAPsys", "TCP", "otherSAPsys", false},
		{"TCP with applServerJava", "TCP", "applServerJava", true},

		{"TCPS with abapSys", "TCPS", "abapSys", false},
		{"TCPS with nonSAPsys", "TCPS", "nonSAPsys", false},
		{"TCPS with BC", "TCPS", "BC", true},
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
