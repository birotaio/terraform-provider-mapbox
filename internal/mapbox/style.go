package mapbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Style represents a Mapbox style.
type Style struct {
	ID         string          `json:"id,omitempty"`
	Version    int             `json:"version"`
	Name       string          `json:"name"`
	Owner      string          `json:"owner,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	Sources    json.RawMessage `json:"sources"`
	Layers     json.RawMessage `json:"layers"`
	Sprite     string          `json:"sprite,omitempty"`
	Glyphs     string          `json:"glyphs,omitempty"`
	Visibility string          `json:"visibility,omitempty"`
	Protected  bool            `json:"protected,omitempty"`
	Draft      bool            `json:"draft,omitempty"`
	Created    string          `json:"created,omitempty"`
	Modified   string          `json:"modified,omitempty"`
}

// CreateStyleRequest is the payload for creating a new style.
type CreateStyleRequest struct {
	Version    int             `json:"version"`
	Name       string          `json:"name,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	Sources    json.RawMessage `json:"sources"`
	Layers     json.RawMessage `json:"layers"`
	Glyphs     string          `json:"glyphs,omitempty"`
	Visibility string          `json:"visibility,omitempty"`
}

// UpdateStyleRequest is the payload for updating an existing style.
type UpdateStyleRequest struct {
	Version    int             `json:"version"`
	Name       string          `json:"name"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	Sources    json.RawMessage `json:"sources"`
	Layers     json.RawMessage `json:"layers"`
	Sprite     string          `json:"sprite,omitempty"`
	Glyphs     string          `json:"glyphs,omitempty"`
	Owner      string          `json:"owner,omitempty"`
	Visibility string          `json:"visibility,omitempty"`
}

// CreateStyle creates a new style.
func (c *Client) CreateStyle(ctx context.Context, req CreateStyleRequest) (*Style, error) {
	path := fmt.Sprintf("/styles/v1/%s", c.Username)

	body, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("creating style: %w", err)
	}

	var style Style
	if err := json.Unmarshal(body, &style); err != nil {
		return nil, fmt.Errorf("decoding create style response: %w", err)
	}

	return &style, nil
}

// GetStyle retrieves a style by its ID.
func (c *Client) GetStyle(ctx context.Context, styleID string) (*Style, error) {
	path := fmt.Sprintf("/styles/v1/%s/%s", c.Username, styleID)

	body, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting style: %w", err)
	}

	var style Style
	if err := json.Unmarshal(body, &style); err != nil {
		return nil, fmt.Errorf("decoding get style response: %w", err)
	}

	return &style, nil
}

// UpdateStyle updates an existing style by its ID.
func (c *Client) UpdateStyle(ctx context.Context, styleID string, req UpdateStyleRequest) (*Style, error) {
	path := fmt.Sprintf("/styles/v1/%s/%s", c.Username, styleID)

	body, err := c.doRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("updating style: %w", err)
	}

	var style Style
	if err := json.Unmarshal(body, &style); err != nil {
		return nil, fmt.Errorf("decoding update style response: %w", err)
	}

	return &style, nil
}

// DeleteStyle deletes a style by its ID.
func (c *Client) DeleteStyle(ctx context.Context, styleID string) error {
	path := fmt.Sprintf("/styles/v1/%s/%s", c.Username, styleID)
	return c.doRequestNoContent(ctx, http.MethodDelete, path)
}
