package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tyr3al/ncctl/pkg/netcup"
)

func TestResolveServerIDAcceptsNumericID(t *testing.T) {
	id, err := resolveServerID(context.Background(), nil, "12345")
	if err != nil {
		t.Fatal(err)
	}
	if id != 12345 {
		t.Fatalf("id = %d, want 12345", id)
	}
}

func TestResolveServerIDByExactName(t *testing.T) {
	client, err := netcup.NewClient("https://example.test/scp-core", netcup.WithHTTPClient(&http.Client{Transport: cliRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/scp-core/api/v1/servers" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		return cliJSONResponse(http.StatusOK, []netcup.ServerListMinimal{
			{ID: 11, Name: "other"},
			{ID: 42, Name: "v220000000000000000"},
		}), nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	id, err := resolveServerID(context.Background(), client, "v220000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("id = %d, want 42", id)
	}
}

func TestResolveServerIDByNickname(t *testing.T) {
	nickname := "main-leviathan"
	client, err := netcup.NewClient("https://example.test/scp-core", netcup.WithHTTPClient(&http.Client{Transport: cliRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return cliJSONResponse(http.StatusOK, []netcup.ServerListMinimal{
			{ID: 11, Name: "v100000000000000000"},
			{ID: 42, Name: "v220000000000000000", Nickname: &nickname},
		}), nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	id, err := resolveServerID(context.Background(), client, "main-leviathan")
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("id = %d, want 42", id)
	}
}

type cliRoundTripFunc func(*http.Request) (*http.Response, error)

func (f cliRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func cliJSONResponse(status int, body any) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(string(data))),
	}
}
