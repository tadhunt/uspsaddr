package uspsaddr

// Config contains the USPS API credentials
type Config struct {
	// ClientID is the OAuth2 client ID from USPS developer portal
	ClientID string

	// ClientSecret is the OAuth2 client secret from USPS developer portal
	ClientSecret string

	// ServerURL is the USPS API server URL (optional, defaults to production)
	// Production: https://apis.usps.com/addresses/v3
	// Testing: https://apis-tem.usps.com/addresses/v3
	ServerURL string

	// TokenURL is the OAuth2 token endpoint (optional, defaults to production)
	// Production: https://apis.usps.com/oauth2/v3/token
	// Testing: https://apis-tem.usps.com/oauth2/v3/token
	TokenURL string
}

// Validate checks if the config is valid
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return &Error{
			Title:  "Invalid configuration",
			Detail: "ClientID is required",
		}
	}
	if c.ClientSecret == "" {
		return &Error{
			Title:  "Invalid configuration",
			Detail: "ClientSecret is required",
		}
	}
	return nil
}

// setDefaults sets default values for optional fields
func (c *Config) setDefaults() {
	if c.ServerURL == "" {
		c.ServerURL = "https://apis.usps.com/addresses/v3"
	}
	if c.TokenURL == "" {
		c.TokenURL = "https://apis.usps.com/oauth2/v3/token"
	}
}
