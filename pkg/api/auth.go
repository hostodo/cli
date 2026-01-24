package api

import (
	"fmt"
)

// NOTE: Login now uses OAuth device flow - see pkg/auth/oauth.go
// The old JWT-based login is deprecated and removed

// RevokeSession revokes the current CLI session on the server
// Note: This is a best-effort call - we still clear local token even if server call fails
func (c *Client) RevokeSession() error {
	// Call logout endpoint to revoke server-side session
	// The backend will invalidate the token hash
	logoutReq := map[string]bool{"logout": true}
	resp, err := c.Post("/v1/auth/", logoutReq)
	if err != nil {
		// Log but don't fail - local cleanup will still happen
		return fmt.Errorf("server revocation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned error %d", resp.StatusCode)
	}

	return nil
}

// GetCurrentUser retrieves the authenticated user's information
func (c *Client) GetCurrentUser() (*User, error) {
	resp, err := c.Get("/client/user/")
	if err != nil {
		return nil, err
	}

	var user User
	if err := parseResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ValidateSession checks if the current session is valid
func (c *Client) ValidateSession() (*User, error) {
	resp, err := c.Get("/v1/auth/")
	if err != nil {
		return nil, err
	}

	var user User
	if err := parseResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
