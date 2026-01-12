package config

import (
	"os"
)

// TwilioConfig holds Twilio configuration
type TwilioConfig struct {
	AccountSID string
	APIKey     string
	APISecret  string
	FromPhone  string
	Enabled    bool
}

// LoadTwilioConfig loads Twilio configuration from environment variables
// Returns config with Enabled=false if credentials are missing (mock mode)
func LoadTwilioConfig() *TwilioConfig {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")
	fromPhone := os.Getenv("TWILIO_FROM_PHONE_NUMBER")

	// Check if all credentials are present
	if accountSID == "" || apiKey == "" || apiSecret == "" || fromPhone == "" {
		return &TwilioConfig{
			Enabled: false,
		}
	}
	return &TwilioConfig{
		AccountSID: accountSID,
		APIKey:     apiKey,
		APISecret:  apiSecret,
		FromPhone:  fromPhone,
		Enabled:    true,
	}
}
