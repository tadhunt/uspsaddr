package uspsaddr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// tokenManager handles OAuth2 token acquisition and automatic refresh
type tokenManager struct {
	clientID     string
	clientSecret string
	tokenURL     string
	httpClient   *http.Client

	mu            sync.RWMutex
	accessToken   string
	expiresAt     time.Time
	refreshBuffer time.Duration // Refresh token this much before expiry
}

// tokenResponse is the OAuth2 token response from USPS
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

// newTokenManager creates a new token manager
func newTokenManager(clientID, clientSecret, tokenURL string, httpClient *http.Client) *tokenManager {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &tokenManager{
		clientID:      clientID,
		clientSecret:  clientSecret,
		tokenURL:      tokenURL,
		httpClient:    httpClient,
		refreshBuffer: 5 * time.Minute, // Refresh 5 minutes before expiry
	}
}

// getToken returns a valid access token, refreshing if necessary
func (tm *tokenManager) getToken() (string, error) {
	tm.mu.RLock()
	// Check if we have a valid token
	if tm.accessToken != "" && time.Now().Before(tm.expiresAt) {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// Need to acquire or refresh token
	return tm.refreshToken()
}

// refreshToken acquires a new access token from USPS
func (tm *tokenManager) refreshToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check in case another goroutine already refreshed
	if tm.accessToken != "" && time.Now().Before(tm.expiresAt) {
		return tm.accessToken, nil
	}

	// Build OAuth2 token request
	reqBody := map[string]string{
		"client_id":     tm.clientID,
		"client_secret": tm.clientSecret,
		"grant_type":    "client_credentials",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	// Make token request
	req, err := http.NewRequest("POST", tm.tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("received empty access token")
	}

	// Store token and calculate expiry
	tm.accessToken = tokenResp.AccessToken
	expiresIn := time.Duration(tokenResp.ExpiresIn) * time.Second
	if expiresIn == 0 {
		expiresIn = 1 * time.Hour // Default to 1 hour if not specified
	}
	tm.expiresAt = time.Now().Add(expiresIn - tm.refreshBuffer)

	return tm.accessToken, nil
}
