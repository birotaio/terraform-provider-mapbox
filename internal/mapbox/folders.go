package mapbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BreadcrumbType string

const (
	BreadcrumbTypeStyle  BreadcrumbType = "style"
	BreadcrumbTypeFolder BreadcrumbType = "folder"
)

type FolderBreadcrumbResponse struct {
	Path []struct {
		ID   string         `json:"id"`
		Type BreadcrumbType `json:"type"`
	} `json:"path"`
}

func (c *Client) GetStyleFolderId(ctx context.Context, styleId string) (string, error) {
	path := fmt.Sprintf("/folders/v1/%s/breadcrumb/%s?type=style&_=%d", c.Username, styleId, time.Now().UnixMilli())

	body, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("getting style folder ID: %w", err)
	}

	var resp FolderBreadcrumbResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("decoding style folder ID response: %w", err)
	}

	if len(resp.Path) < 2 {
		return "", fmt.Errorf("unexpected breadcrumb response: %v", resp)
	} else if resp.Path[len(resp.Path)-1].Type != BreadcrumbTypeStyle {
		return "", fmt.Errorf("unexpected breadcrumb response, second-to-last item is not style: %v", resp)
	}

	// The folder ID is the second-to-last item in the path (the last item is the style itself)
	folderId := resp.Path[len(resp.Path)-2].ID
	return folderId, nil
}

type FolderFile struct {
	ID   string         `json:"id"`
	Type BreadcrumbType `json:"type"`
}

type SetStyleFolderIdRequest struct {
	Parent string       `json:"parent"`
	Files  []FolderFile `json:"files"`
}

func (c *Client) SetStyleFolderId(ctx context.Context, styleId, folderId string) error {
	path := fmt.Sprintf("/folders/v1/%s", c.Username)

	req := SetStyleFolderIdRequest{
		Parent: folderId,
		Files: []FolderFile{
			{
				ID:   styleId,
				Type: BreadcrumbTypeStyle,
			},
		},
	}

	_, err := c.doRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return fmt.Errorf("setting style folder ID: %w", err)
	}

	return nil
}
