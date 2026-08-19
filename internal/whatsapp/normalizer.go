package whatsapp

import (
	"regexp"
	"strings"
)

var nonDigitRegex = regexp.MustCompile(`[^\d+]`)

// NormalizePhoneNumber formats phone numbers into standard E.164 (e.g. +5562999999999)
func NormalizePhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}

	hasPlus := strings.HasPrefix(phone, "+")

	// Remove all non-digits
	cleaned := nonDigitRegex.ReplaceAllString(phone, "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	if cleaned == "" {
		return ""
	}

	// Always return with leading +
	_ = hasPlus
	return "+" + cleaned
}
