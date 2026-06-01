package netcup

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDoJSONBuildsRequestAndDecodesResponse(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/scp-core/api/v1/servers" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got, want := r.URL.Query().Get("limit"), "10"; got != want {
			t.Fatalf("limit = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}
		return jsonResponse(http.StatusOK, []ServerListMinimal{{ID: 7, Name: "vps"}}), nil
	})

	client, err := NewClient("https://example.test/scp-core", WithHTTPClient(&http.Client{Transport: transport}), WithTokenSource(StaticToken("test-token")))
	if err != nil {
		t.Fatal(err)
	}
	servers, err := client.ListServers(context.Background(), ListServersOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListServers() error = %v", err)
	}
	if len(servers) != 1 || servers[0].ID != 7 {
		t.Fatalf("servers = %#v", servers)
	}
}

func TestDoJSONReturnsAPIError(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusBadRequest, ResponseError{Code: "bad_request", Message: "nope"}), nil
	})

	client, err := NewClient("https://example.test", WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	err = client.Ping(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Code != "bad_request" || apiErr.Message != "nope" {
		t.Fatalf("apiErr = %#v", apiErr)
	}
}

func TestReadOnlyEndpointPaths(t *testing.T) {
	seen := map[string]bool{}
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seen[r.Method+" "+r.URL.Path] = true
		switch r.URL.Path {
		case "/scp-core/api/v1/servers/42":
			return jsonResponse(http.StatusOK, Server{ID: 42}), nil
		case "/scp-core/api/v1/servers/42/interfaces":
			return jsonResponse(http.StatusOK, []Interface{{MAC: "aa:bb"}}), nil
		case "/scp-core/api/v1/users/9/failoverips/v4":
			return jsonResponse(http.StatusOK, []FailoverIPv4{{ID: 1, IP: "192.0.2.1"}}), nil
		case "/scp-core/api/v1/users/9/failoverips/v6":
			return jsonResponse(http.StatusOK, []FailoverIPv6{{ID: 2, NetworkPrefix: "2001:db8::"}}), nil
		case "/scp-core/api/v1/tasks/task-1":
			return jsonResponse(http.StatusOK, TaskInfo{TaskInfoMinimal: TaskInfoMinimal{UUID: "task-1"}}), nil
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
			return nil, nil
		}
	})

	client, err := NewClient("https://example.test/scp-core", WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := client.GetServer(ctx, 42, true); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListInterfaces(ctx, 42, true); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListFailoverIPv4(ctx, 9, ListFailoverOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListFailoverIPv6(ctx, 9, ListFailoverOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetTask(ctx, "task-1"); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{
		"GET /scp-core/api/v1/servers/42",
		"GET /scp-core/api/v1/servers/42/interfaces",
		"GET /scp-core/api/v1/users/9/failoverips/v4",
		"GET /scp-core/api/v1/users/9/failoverips/v6",
		"GET /scp-core/api/v1/tasks/task-1",
	} {
		if !seen[key] {
			t.Fatalf("did not see %s", key)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonResponse(status int, body any) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(string(data))),
	}
}
