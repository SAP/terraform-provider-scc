package systemMapping

import "testing"

func TestValidateHTTPPort(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"minimum valid", "1", true},
		{"common port", "80", true},
		{"maximum valid", "65535", true},

		{"zero", "0", false},
		{"above maximum", "65536", false},
		{"way above maximum", "70000", false},
		{"negative", "-1", false},
		{"alpha", "abc", false},
		{"mixed", "80abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateHTTPPort(tt.value); got != tt.expected {
				t.Errorf("ValidateHTTPPort(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestValidateRFCValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		// Numeric
		{"numeric valid", "8080", true},
		{"numeric max", "65535", true},
		{"numeric min", "1", true},

		// sapgw
		{"sapgw00", "sapgw00", true},
		{"sapgw99", "sapgw99", true},

		// sapms
		{"sapmsABC", "sapmsABC", true},
		{"sapmsA01", "sapmsA01", true},

		// Invalid sapgw
		{"sapgw100", "sapgw100", false},
		{"sapgw", "sapgw", false},

		// Invalid sapms
		{"reserved sid SAP", "sapmsSAP", false},
		{"reserved sid SYS", "sapmsSYS", false},
		{"lowercase sid", "sapmsabc", false},
		{"starts with digit", "sapms1AB", false},
		{"too short", "sapmsAB", false},

		// Invalid numeric
		{"zero", "0", false},
		{"too large", "70000", false},

		// Garbage
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateRFCValue(tt.value); got != tt.expected {
				t.Errorf("ValidateRFCValue(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestValidateRFCSValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		// Numeric
		{"numeric valid", "443", true},
		{"numeric max", "65535", true},

		// sapgwXXs
		{"sapgw00s", "sapgw00s", true},
		{"sapgw99s", "sapgw99s", true},

		// sapms
		{"sapmsXYZ", "sapmsXYZ", true},
		{"sapmsA01", "sapmsA01", true},

		// Invalid sapgw
		{"missing s suffix", "sapgw01", false},
		{"sapgw100s", "sapgw100s", false},

		// Invalid sapms
		{"reserved sid SAP", "sapmsSAP", false},
		{"lowercase sid", "sapmsxyz", false},

		// Invalid numeric
		{"zero", "0", false},
		{"too large", "70000", false},

		// Garbage
		{"invalid", "notvalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateRFCSValue(tt.value); got != tt.expected {
				t.Errorf("ValidateRFCSValue(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestIsValidSID(t *testing.T) {
	tests := []struct {
		name     string
		sid      string
		expected bool
	}{
		{"valid SID", "ABC", true},
		{"valid alphanumeric", "A01", true},

		{"reserved SAP", "SAP", false},
		{"reserved SYS", "SYS", false},

		{"starts with digit", "1AB", false},
		{"lowercase", "abc", false},
		{"too short", "AB", false},
		{"too long", "ABCD", false},
		{"special chars", "A-C", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidSID(tt.sid); got != tt.expected {
				t.Errorf("isValidSID(%q) = %v, want %v", tt.sid, got, tt.expected)
			}
		})
	}
}
