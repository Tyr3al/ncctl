package netcup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultAPIBaseURL  = "https://servercontrolpanel.de/scp-core"
	DefaultAuthBaseURL = "https://servercontrolpanel.de"
)

type TokenSource interface {
	Token(context.Context) (string, error)
}

type StaticToken string

func (t StaticToken) Token(context.Context) (string, error) {
	return string(t), nil
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      TokenSource
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithTokenSource(token TokenSource) Option {
	return func(c *Client) {
		c.token = token
	}
}

func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("base URL must be absolute: %q", baseURL)
	}
	client := &Client{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}

func (c *Client) DoJSON(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	return c.doJSON(ctx, method, path, query, body, out, "application/json")
}

func (c *Client) DoMergePatch(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	return c.doJSON(ctx, method, path, query, body, out, "application/merge-patch+json")
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body any, out any, contentType string) error {
	req, err := c.newRequest(ctx, method, path, query, body, contentType)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, out)
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body any, contentType string) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	target := c.baseURL.ResolveReference(&url.URL{Path: joinURLPath(c.baseURL.Path, path)})
	if len(query) > 0 {
		target.RawQuery = query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, target.String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	if c.token != nil {
		token, err := c.token.Token(ctx)
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	return req, nil
}

func decodeResponse(resp *http.Response, out any) error {
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Status: resp.Status}
		if len(data) > 0 {
			_ = json.Unmarshal(data, &apiErr.ResponseError)
		}
		return apiErr
	}
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

type APIError struct {
	StatusCode int
	Status     string
	ResponseError
}

func (e *APIError) Error() string {
	if e.Code != "" || e.Message != "" {
		return fmt.Sprintf("netcup API error: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("netcup API error: %s", e.Status)
}

func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

func joinURLPath(basePath, requestPath string) string {
	basePath = strings.TrimRight(basePath, "/")
	requestPath = "/" + strings.TrimLeft(requestPath, "/")
	if basePath == "" {
		return requestPath
	}
	return basePath + requestPath
}
