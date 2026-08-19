package whatsapp

import "testing"

func TestNormalizePhoneNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+55 62 99999-9999", "+5562999999999"},
		{"5562999999999", "+5562999999999"},
		{"+1 (555) 123-4567", "+15551234567"},
		{"  +55-62-98888.7777  ", "+5562988887777"},
		{"", ""},
		{"+", ""},
	}

	for _, tt := range tests {
		got := NormalizePhoneNumber(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizePhoneNumber(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
