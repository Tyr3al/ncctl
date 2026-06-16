package netcup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultClientID = "scp"

var (
	ErrAuthorizationPending = errors.New("authorization pending")
	ErrAuthorizationDenied  = errors.New("authorization denied")
	ErrAuthorizationExpired = errors.New("authorization expired")
)

type AuthClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	clientID   string
}

type AuthOption func(*AuthClient)

func WithAuthHTTPClient(httpClient *http.Client) AuthOption {
	return func(c *AuthClient) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithClientID(clientID string) AuthOption {
	return func(c *AuthClient) {
		if clientID != "" {
			c.clientID = clientID
		}
	}
}

func NewAuthClient(baseURL string, opts ...AuthOption) (*AuthClient, error) {
	if baseURL == "" {
		baseURL = DefaultAuthBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse auth base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("auth base URL must be absolute: %q", baseURL)
	}
	client := &AuthClient{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		clientID: defaultClientID,
	}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}

type DeviceAuthorization struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in,omitempty"`
	Expires                 int    `json:"expires,omitempty"`
	Interval                int    `json:"interval,omitempty"`
}

func (d DeviceAuthorization) ExpiresAfter() time.Duration {
	if d.ExpiresIn > 0 {
		return time.Duration(d.ExpiresIn) * time.Second
	}
	return time.Duration(d.Expires) * time.Second
}

func (d DeviceAuthorization) PollInterval() time.Duration {
	if d.Interval <= 0 {
		return 5 * time.Second
	}
	return time.Duration(d.Interval) * time.Second
}

func (c *AuthClient) StartDeviceAuthorization(ctx context.Context) (*DeviceAuthorization, error) {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("scope", "offline_access openid")
	var out DeviceAuthorization
	if err := c.postForm(ctx, "/realms/scp/protocol/openid-connect/auth/device", values, &out); err != nil {
		return nil, err
	}
	out.VerificationURI = c.absoluteMaybe(out.VerificationURI)
	out.VerificationURIComplete = c.absoluteMaybe(out.VerificationURIComplete)
	return &out, nil
}

func (c *AuthClient) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	values.Set("device_code", deviceCode)
	var out TokenResponse
	if err := c.postForm(ctx, "/realms/scp/protocol/openid-connect/token", values, &out); err != nil {
		if tokenErr, ok := asTokenError(err); ok {
			switch tokenErr.ErrorCode {
			case "authorization_pending", "slow_down":
				return nil, ErrAuthorizationPending
			case "access_denied":
				return nil, ErrAuthorizationDenied
			case "expired_token":
				return nil, ErrAuthorizationExpired
			}
		}
		return nil, err
	}
	return &out, nil
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	var out TokenResponse
	if err := c.postForm(ctx, "/realms/scp/protocol/openid-connect/token", values, &out); err != nil {
		return nil, err
	}
	if out.RefreshToken == "" {
		out.RefreshToken = refreshToken
	}
	return &out, nil
}

func (c *AuthClient) RevokeToken(ctx context.Context, token string) error {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("token", token)
	values.Set("token_type_hint", "refresh_token")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.resolve("/realms/scp/protocol/openid-connect/revoke").String(),
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	data, _ := io.ReadAll(resp.Body)
	var tokenErr tokenError
	if len(data) > 0 {
		_ = json.Unmarshal(data, &tokenErr)
	}
	tokenErr.StatusCode = resp.StatusCode
	tokenErr.Status = resp.Status
	return &tokenErr
}

func (c *AuthClient) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resolve("/realms/scp/protocol/openid-connect/userinfo").String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out UserInfo
	if err := decodeResponse(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type TokenResponse struct {
	AccessToken     string `json:"access_token"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	Expires         int    `json:"expires,omitempty"`
	RefreshExpires  int    `json:"refresh_expires,omitempty"`
	RefreshToken    string `json:"refresh_token"`
	TokenType       string `json:"token_type"`
	NotBeforePolicy int    `json:"not-before-policy,omitempty"`
	SessionState    string `json:"session_state,omitempty"`
	Scope           string `json:"scope,omitempty"`
}

func (t TokenResponse) ExpiresAfter() time.Duration {
	if t.ExpiresIn > 0 {
		return time.Duration(t.ExpiresIn) * time.Second
	}
	return time.Duration(t.Expires) * time.Second
}

type RefreshTokenSource struct {
	auth         *AuthClient
	refreshToken string
	mu           sync.Mutex
	accessToken  string
	expiresAt    time.Time
}

func NewRefreshTokenSource(auth *AuthClient, refreshToken string) *RefreshTokenSource {
	return &RefreshTokenSource{auth: auth, refreshToken: refreshToken}
}

func (s *RefreshTokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.accessToken != "" && time.Now().Before(s.expiresAt.Add(-30*time.Second)) {
		return s.accessToken, nil
	}
	resp, err := s.auth.RefreshToken(ctx, s.refreshToken)
	if err != nil {
		return "", err
	}
	s.accessToken = resp.AccessToken
	if resp.RefreshToken != "" {
		s.refreshToken = resp.RefreshToken
	}
	expiresAfter := resp.ExpiresAfter()
	if expiresAfter <= 0 {
		expiresAfter = 5 * time.Minute
	}
	s.expiresAt = time.Now().Add(expiresAfter)
	return s.accessToken, nil
}

func (s *RefreshTokenSource) RefreshToken() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.refreshToken
}

func (c *AuthClient) postForm(ctx context.Context, path string, values url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.resolve(path).String(), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return decodeResponse(resp, out)
	}
	data, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	var tokenErr tokenError
	if len(data) > 0 {
		_ = json.Unmarshal(data, &tokenErr)
	}
	tokenErr.StatusCode = resp.StatusCode
	tokenErr.Status = resp.Status
	if tokenErr.ErrorCode == "" {
		var responseErr ResponseError
		_ = json.Unmarshal(data, &responseErr)
		return &APIError{StatusCode: resp.StatusCode, Status: resp.Status, ResponseError: responseErr}
	}
	return &tokenErr
}

func (c *AuthClient) resolve(path string) *url.URL {
	return c.baseURL.ResolveReference(&url.URL{Path: joinURLPath(c.baseURL.Path, path)})
}

func (c *AuthClient) absoluteMaybe(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err == nil {
		if parsed.IsAbs() {
			return raw
		}
		return c.baseURL.ResolveReference(parsed).String()
	}
	return c.resolve(raw).String()
}

type tokenError struct {
	StatusCode       int
	Status           string
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *tokenError) Error() string {
	if e.ErrorDescription != "" {
		return e.ErrorCode + ": " + e.ErrorDescription
	}
	if e.ErrorCode != "" {
		return e.ErrorCode
	}
	return "token request failed: " + strconv.Itoa(e.StatusCode)
}

func asTokenError(err error) (*tokenError, bool) {
	var tokenErr *tokenError
	if errors.As(err, &tokenErr) {
		return tokenErr, true
	}
	return nil, false
}
