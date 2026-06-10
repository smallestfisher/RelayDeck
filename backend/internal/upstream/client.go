package upstream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type Client struct {
	httpClient *http.Client
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type Error struct {
	StatusCode int
	Body       []byte
}

func (e *Error) Error() string {
	return fmt.Sprintf("upstream returned status %d", e.StatusCode)
}

func NewClient(timeout time.Duration) *Client {
	return &Client{httpClient: &http.Client{Timeout: timeout}}
}

func NewClientWithHTTPClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

func (c *Client) DoJSON(ctx context.Context, site domain.UpstreamSite, path string, body []byte) (Response, error) {
	baseURL := strings.TrimRight(site.BaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if site.Credential != "" {
		req.Header.Set("Authorization", "Bearer "+site.Credential)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	if resp.StatusCode >= 500 {
		return Response{}, &Error{StatusCode: resp.StatusCode, Body: respBody}
	}
	return Response{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: respBody}, nil
}
