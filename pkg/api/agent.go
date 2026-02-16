package api

import (
	"fmt"
)

// AgentToken represents an agent token for an instance
type AgentToken struct {
	ID           int    `json:"id"`
	InstanceID   string `json:"instance_id"`
	Hostname     string `json:"hostname"`
	Status       string `json:"token_status"` // "active" or "revoked"
	CreatedAt    string `json:"created_at"`
	LastUsedAt   string `json:"last_used_at"`
}

// AgentSettings represents account-level agent settings
type AgentSettings struct {
	Enabled      bool   `json:"enabled"`
	UseOwnKey    bool   `json:"use_own_key"`
	HasOwnKey    bool   `json:"has_own_key"`
	MonthlyLimit int    `json:"monthly_token_limit"`
	TokensUsed   int    `json:"tokens_used_this_month"`
}

// GetAgentTokens retrieves all agent tokens for the authenticated user
func (c *Client) GetAgentTokens() ([]AgentToken, error) {
	path := "/v1/client/agent-tokens/"

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var tokens []AgentToken
	if err := parseResponse(resp, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// GetAgentToken retrieves a specific agent token by instance ID
func (c *Client) GetAgentToken(instanceID string) (*AgentToken, error) {
	path := fmt.Sprintf("/v1/client/agent-tokens/%s/", instanceID)

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var token AgentToken
	if err := parseResponse(resp, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// GetAgentSettings retrieves account-level agent settings
func (c *Client) GetAgentSettings() (*AgentSettings, error) {
	path := "/v1/client/agent-settings/"

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var settings AgentSettings
	if err := parseResponse(resp, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// RevokeResponse represents the response from a revoke operation
type RevokeResponse struct {
	Message string `json:"message"`
	Count   int    `json:"count"` // Used for revoke-all
}

// RegenerateTokenResponse represents the response from regenerate operation
type RegenerateTokenResponse struct {
	Token      string `json:"token"`       // Plaintext token (one-time display)
	InstanceID string `json:"instance_id"`
	CreatedAt  string `json:"created_at"`
}

// RevokeAgentToken revokes the agent token for a specific instance
func (c *Client) RevokeAgentToken(instanceID string) error {
	path := fmt.Sprintf("/v1/client/agent-tokens/%s/revoke/", instanceID)

	resp, err := c.Post(path, nil)
	if err != nil {
		return err
	}

	if err := parseResponse(resp, nil); err != nil {
		return err
	}

	return nil
}

// RevokeAllAgentTokens revokes agent tokens for all instances
func (c *Client) RevokeAllAgentTokens() (*RevokeResponse, error) {
	path := "/v1/client/agent-tokens/revoke-all/"

	resp, err := c.Post(path, nil)
	if err != nil {
		return nil, err
	}

	var revokeResp RevokeResponse
	if err := parseResponse(resp, &revokeResp); err != nil {
		return nil, err
	}

	return &revokeResp, nil
}

// RegenerateAgentToken regenerates an agent token, returns plaintext token
func (c *Client) RegenerateAgentToken(instanceID string) (*RegenerateTokenResponse, error) {
	path := fmt.Sprintf("/v1/client/agent-tokens/%s/regenerate/", instanceID)

	resp, err := c.Post(path, nil)
	if err != nil {
		return nil, err
	}

	var tokenResp RegenerateTokenResponse
	if err := parseResponse(resp, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
