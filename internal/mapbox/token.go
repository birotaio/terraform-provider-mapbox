package mapbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Token represents a Mapbox access token.
type Token struct {
	ID          string   `json:"id"`
	Usage       string   `json:"usage"`
	Client      string   `json:"client"`
	Default     bool     `json:"default"`
	Note        string   `json:"note"`
	Scopes      []string `json:"scopes"`
	AllowedUrls []string `json:"allowedUrls,omitempty"`
	TokenString string   `json:"token"`
	Created     string   `json:"created"`
	Modified    string   `json:"modified"`
}

// CreateTokenRequest is the payload for creating a new token.
type CreateTokenRequest struct {
	Note        string   `json:"note,omitempty"`
	Scopes      []string `json:"scopes"`
	AllowedUrls []string `json:"allowedUrls,omitempty"`
}

// UpdateTokenRequest is the payload for updating an existing token.
type UpdateTokenRequest struct {
	Note        *string  `json:"note,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	AllowedUrls []string `json:"allowedUrls,omitempty"`
}

// ListTokens retrieves all tokens for the configured username.
func (c *Client) ListTokens(ctx context.Context) ([]Token, error) {
	path := fmt.Sprintf("/tokens/v2/%s", c.Username)

	body, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("listing tokens: %w", err)
	}

	var tokens []Token
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("decoding tokens response: %w", err)
	}

	return tokens, nil
}

// CreateToken creates a new access token.
func (c *Client) CreateToken(ctx context.Context, req CreateTokenRequest) (*Token, error) {
	path := fmt.Sprintf("/tokens/v2/%s", c.Username)

	body, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("creating token: %w", err)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decoding create token response: %w", err)
	}

	return &token, nil
}

// GetToken retrieves a token by listing all tokens and filtering by ID.
func (c *Client) GetToken(ctx context.Context, tokenID string) (*Token, error) {
	tokens, err := c.ListTokens(ctx)
	if err != nil {
		return nil, err
	}

	for _, t := range tokens {
		if t.ID == tokenID {
			return &t, nil
		}
	}

	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("token with ID %q not found", tokenID),
	}
}

// UpdateToken updates an existing token by its ID.
func (c *Client) UpdateToken(ctx context.Context, tokenID string, req UpdateTokenRequest) (*Token, error) {
	path := fmt.Sprintf("/tokens/v2/%s/%s", c.Username, tokenID)

	body, err := c.doRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("updating token: %w", err)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decoding update token response: %w", err)
	}

	return &token, nil
}

// DeleteToken deletes a token by its ID.
func (c *Client) DeleteToken(ctx context.Context, tokenID string) error {
	path := fmt.Sprintf("/tokens/v2/%s/%s", c.Username, tokenID)
	return c.doRequestNoContent(ctx, http.MethodDelete, path)
}
