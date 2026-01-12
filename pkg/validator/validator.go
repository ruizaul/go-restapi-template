// Package validator provides validation utilities for email, phone, UUID, and business logic rules
package validator

import (
	"regexp"

	"github.com/google/uuid"
)

var (
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex      = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`) // E.164 format
	phoneDigitsOnly = regexp.MustCompile(`[^0-9]`)             // For cleaning phone numbers
)

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidPhone validates phone number (E.164 format)
func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// IsValidUUID validates UUID format
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// IsValidRole validates user role
func IsValidRole(role string) bool {
	validRoles := map[string]bool{
		"customer": true,
		"merchant": true,
		"driver":   true,
		"admin":    true,
	}
	return validRoles[role]
}

// IsValidOrderStatus validates order status
func IsValidOrderStatus(status string) bool {
	validStatuses := map[string]bool{
		"pending":   true,
		"assigned":  true,
		"picked_up": true,
		"delivered": true,
		"canceled":  true,
	}
	return validStatuses[status]
}

// IsValidDriverStatus validates driver status
func IsValidDriverStatus(status string) bool {
	validStatuses := map[string]bool{
		"available": true,
		"busy":      true,
		"offline":   true,
	}
	return validStatuses[status]
}

// NormalizePhone converts a phone number to E.164 format
// Accepts formats like: 526621816014, +526621816014, (52) 662-181-6014
// Returns format: +526621816014
func NormalizePhone(phone string) string {
	// Remove all non-digit characters
	digitsOnly := phoneDigitsOnly.ReplaceAllString(phone, "")

	// If already starts with +, just clean it
	if len(phone) > 0 && phone[0] == '+' {
		return "+" + digitsOnly
	}

	// Add + prefix if not present
	if len(digitsOnly) > 0 {
		return "+" + digitsOnly
	}

	return phone // Return original if something went wrong
}
