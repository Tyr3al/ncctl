package netcup

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDeviceAuthorizationAndPollingErrors(t *testing.T) {
	var requests []string
	auth, err := NewAuthClient("https://auth.example.test", WithAuthHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		data, _ := io.ReadAll(r.Body)
		requests = append(requests, r.URL.Path+" "+string(data))
		switch r.URL.Path {
		case "/realms/scp/protocol/openid-connect/auth/device":
			return jsonResponse(http.StatusOK, DeviceAuthorization{
				DeviceCode:              "device",
				UserCode:                "USER",
				VerificationURI:         "/realms/scp/device",
				VerificationURIComplete: "/realms/scp/device?user_code=USER",
				Expires:                 600,
				Interval:                5,
			}), nil
		case "/realms/scp/protocol/openid-connect/token":
			return jsonResponse(http.StatusBadRequest, map[string]string{"error": "authorization_pending"}), nil
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
			return nil, nil
		}
	})}))
	if err != nil {
		t.Fatal(err)
	}
	device, err := auth.StartDeviceAuthorization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if device.VerificationURIComplete != "https://auth.example.test/realms/scp/device?user_code=USER" {
		t.Fatalf("verification URI = %q", device.VerificationURIComplete)
	}
	_, err = auth.ExchangeDeviceCode(context.Background(), "device")
	if err != ErrAuthorizationPending {
		t.Fatalf("ExchangeDeviceCode() error = %v, want pending", err)
	}
	if !strings.Contains(requests[0], "client_id=scp") || !strings.Contains(requests[0], "scope=offline_access+openid") {
		t.Fatalf("device request body = %q", requests[0])
	}
}

func TestRevokeToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		auth, err := NewAuthClient("https://auth.example.test", WithAuthHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/realms/scp/protocol/openid-connect/revoke" {
				t.Fatalf("unexpected path %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			bs := string(body)
			if !strings.Contains(bs, "token=mytoken") || !strings.Contains(bs, "client_id=scp") {
				t.Fatalf("unexpected body %q", bs)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		})}))
		if err != nil {
			t.Fatal(err)
		}
		if err := auth.RevokeToken(context.Background(), "mytoken"); err != nil {
			t.Fatalf("RevokeToken() error = %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		auth, err := NewAuthClient("https://auth.example.test", WithAuthHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusBadRequest, map[string]string{"error": "invalid_token", "error_description": "Token not active"}), nil
		})}))
		if err != nil {
			t.Fatal(err)
		}
		if err := auth.RevokeToken(context.Background(), "badtoken"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestRefreshTokenSourceCachesAccessToken(t *testing.T) {
	var refreshes int
	auth, err := NewAuthClient("https://auth.example.test", WithAuthHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		refreshes++
		return jsonResponse(http.StatusOK, TokenResponse{AccessToken: "access", RefreshToken: "refresh2", ExpiresIn: 300}), nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	source := NewRefreshTokenSource(auth, "refresh1")
	first, err := source.Token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := source.Token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first != "access" || second != "access" || refreshes != 1 {
		t.Fatalf("tokens = %q/%q refreshes=%d", first, second, refreshes)
	}
	if source.RefreshToken() != "refresh2" {
		t.Fatalf("refresh token = %q", source.RefreshToken())
	}
}
