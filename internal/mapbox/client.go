package mapbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const defaultBaseURL = "https://api.mapbox.com"

// Client is an HTTP client for the Mapbox API.
type Client struct {
	HTTPClient  *http.Client
	BaseURL     string
	AccessToken string
	Fresh       bool
	Username    string
}

// NewClient creates a new Mapbox API client.
func NewClient(accessToken, username string, fresh bool) *Client {
	return &Client{
		HTTPClient:  http.DefaultClient,
		BaseURL:     defaultBaseURL,
		AccessToken: accessToken,
		Username:    username,
		Fresh:       fresh,
	}
}

// doRequest executes an authenticated HTTP request against the Mapbox API.
func (c *Client) doRequest(ctx context.Context, method, path string, body any) ([]byte, error) {
	ctx = tflog.NewSubsystem(ctx, "http_client")

	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	q := u.Query()
	q.Set("access_token", c.AccessToken)
	if c.Fresh {
		q.Set("fresh", "true")
	}

	u.RawQuery = q.Encode()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	tflog.Trace(ctx, fmt.Sprintf("Making %s request to %s", method, u.String()))
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	tflog.Trace(ctx, fmt.Sprintf("Performing %s request to %s", method, u.String()))

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		tflog.Trace(ctx, fmt.Sprintf("Request body: %s", reqBody))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	tflog.Trace(ctx, fmt.Sprintf("Received response with status %d", resp.StatusCode))
	tflog.Trace(ctx, fmt.Sprintf("Response body: %s", respBody))

	if err := checkResponse(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *Client) doRequestRawText(ctx context.Context, method, path string, body string) (string, error) {
	ctx = tflog.NewSubsystem(ctx, "http_client")

	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return "", fmt.Errorf("parsing URL: %w", err)
	}

	q := u.Query()
	q.Set("access_token", c.AccessToken)
	if c.Fresh {
		q.Set("fresh", "true")
	}

	u.RawQuery = q.Encode()

	reqBody := bytes.NewReader([]byte(body))

	tflog.Trace(ctx, fmt.Sprintf("Making %s request to %s", method, u.String()))
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	tflog.Trace(ctx, fmt.Sprintf("Performing %s request to %s", method, u.String()))

	req.Header.Set("Content-Type", "text/plain")
	tflog.Trace(ctx, fmt.Sprintf("Request body: %s", reqBody))
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	tflog.Trace(ctx, fmt.Sprintf("Received response with status %d", resp.StatusCode))
	tflog.Trace(ctx, fmt.Sprintf("Response body: %s", respBody))

	if err := checkResponse(resp.StatusCode, respBody); err != nil {
		return "", err
	}

	return string(respBody), nil
}

// doRequestNoContent executes an authenticated HTTP request that expects no response body.
func (c *Client) doRequestNoContent(ctx context.Context, method, path string) error {
	ctx = tflog.NewSubsystem(ctx, "http_client")
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	q := u.Query()
	q.Set("access_token", c.AccessToken)
	if c.Fresh {
		q.Set("fresh", "true")
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	tflog.Trace(ctx, fmt.Sprintf("Performing %s request to %s", method, u.String()))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	tflog.Trace(ctx, fmt.Sprintf("Received response with status %d", resp.StatusCode))
	tflog.Trace(ctx, fmt.Sprintf("Response body: %s", respBody))

	return checkResponse(resp.StatusCode, respBody)
}
