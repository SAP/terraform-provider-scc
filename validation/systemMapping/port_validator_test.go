package systemMapping

import "testing"

func TestValidateHTTPPort(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"80", true},
		{"65535", true},
		{"0", false},
		{"70000", false},
		{"abc", false},
	}

	for _, tt := range tests {
		if got := ValidateHTTPPort(tt.value); got != tt.expected {
			t.Errorf("HTTPPort(%q) expected %v, got %v", tt.value, tt.expected, got)
		}
	}
}

func TestValidateRFCValue(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"sapmsABC", true},
		{"sapgw00", true},
		{"8080", true},
		{"sapgw100", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		if got := ValidateRFCValue(tt.value); got != tt.expected {
			t.Errorf("RFCValue(%q) expected %v, got %v", tt.value, tt.expected, got)
		}
	}
}

func TestValidateRFCSValue(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"sapmsXYZ", true},
		{"sapgw01s", true},
		{"443", true},
		{"sapgw01", false}, // missing "s"
		{"notvalid", false},
	}

	for _, tt := range tests {
		if got := ValidateRFCSValue(tt.value); got != tt.expected {
			t.Errorf("RFCSValue(%q) expected %v, got %v", tt.value, tt.expected, got)
		}
	}
}
