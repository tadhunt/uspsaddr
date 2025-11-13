package uspsaddr

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tadhunt/logger"

	"github.com/tadhunt/uspsaddr/uspsinternal"
)

// Client provides address validation using the USPS API
type Client struct {
	config       Config
	tokenManager *tokenManager
	client       *uspsinternal.ClientWithResponses
	httpClient   *http.Client
	log          logger.CompatLogWriter
}

// NewClient creates a new USPS address validation client
// The config must contain ClientID and ClientSecret from the USPS developer portal
func NewClient(config Config) (*Client, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}

	level := logger.LogLevel_INFO
	if config.Debug {
		level = logger.LogLevel_DEBUG
	}
	log := logger.NewCompatLogWriter(level)

	// Set defaults
	config.setDefaults()

	c := &Client{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		log:        log,
	}

	// Create token manager
	c.tokenManager = newTokenManager(
		config.ClientID,
		config.ClientSecret,
		config.TokenURL,
		c.httpClient,
	)

	// Create the USPS client with token injection
	client, err := uspsinternal.NewClientWithResponses(
		config.ServerURL,
		uspsinternal.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			token, err := c.tokenManager.getToken()
			if err != nil {
				return fmt.Errorf("failed to get access token: %w", err)
			}
			req.Header.Set("Authorization", "Bearer "+token)
			// Debug: log the full request URL
			log.Debugf("USPS API Request URL: %s\n", req.URL.String())
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create USPS client: %w", err)
	}

	c.client = client

	return c, nil
}

// ValidateAddress validates and canonicalizes an address
// Returns an array of validation results (typically one, but may be multiple for ambiguous addresses)
func (c *Client) ValidateAddress(ctx context.Context, address *Address) ([]ValidationResult, error) {
	if address == nil {
		return nil, fmt.Errorf("address cannot be nil")
	}

	// Validate required fields
	if address.StreetAddress == "" {
		return nil, fmt.Errorf("street address is required")
	}

	params := &uspsinternal.GetAddressParams{
		StreetAddress: address.StreetAddress,
	}

	if address.State == "" {
		return nil, fmt.Errorf("2 letter state abbreviation is required")
	}

	if len(address.State) != 2 {
		return nil, fmt.Errorf("2 letter state abbreviation is required")
	}

	params.State = strings.ToUpper(address.State)

	if address.SecondaryAddress != "" {
		params.SecondaryAddress = &address.SecondaryAddress
	}
	if address.City != "" {
		params.City = &address.City
	}
	if address.ZIPCode != "" {
		params.ZIPCode = &address.ZIPCode
	}
	if address.Firm != "" {
		params.Firm = &address.Firm
	}
	if address.Urbanization != "" {
		params.Urbanization = &address.Urbanization
	}

	c.log.Debugf("Calling USPS API with params:\n")
	c.log.Debugf("  StreetAddress: %q\n", params.StreetAddress)
	c.log.Debugf("  State: %q\n", params.State)
	if params.SecondaryAddress != nil {
		c.log.Debugf("  SecondaryAddress: %q\n", *params.SecondaryAddress)
	} else {
		c.log.Debugf("  SecondaryAddress: nil\n")
	}
	if params.City != nil {
		c.log.Debugf("  City: %q\n", *params.City)
	}
	if params.ZIPCode != nil {
		c.log.Debugf("  ZIPCode: %q\n", *params.ZIPCode)
	}

	// Call USPS API
	resp, err := c.client.GetAddressWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("USPS API request failed: %w", err)
	}

	// Debug: dump the response
	c.log.Debugf("USPS API Response Status: %d\n", resp.StatusCode())
	if resp.Body != nil {
		c.log.Debugf("USPS API Response Body: %s\n", string(resp.Body))
	}

	// Handle error responses
	if resp.StatusCode() != http.StatusOK {
		if resp.JSON400 != nil {
			return nil, convertError(resp.JSON400)
		}
		if resp.JSON401 != nil {
			return nil, convertError(resp.JSON401)
		}
		if resp.JSON403 != nil {
			return nil, convertError(resp.JSON403)
		}
		if resp.JSON404 != nil {
			return nil, convertError(resp.JSON404)
		}
		if resp.JSON429 != nil {
			return nil, convertError(resp.JSON429)
		}
		if resp.JSON503 != nil {
			return nil, convertError(resp.JSON503)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	// Convert response to our types
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response")
	}

	result := convertResponse(resp.JSON200)
	return []ValidationResult{result}, nil
}
