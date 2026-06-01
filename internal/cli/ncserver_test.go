package cli

import (
	"context"
	"net/http"
	"testing"

	"github.com/tyr3al/ncctl/pkg/netcup"
)

func TestIdentifyServerByIPFindsMatchingServer(t *testing.T) {
	client, err := netcup.NewClient("https://example.test/scp-core", netcup.WithHTTPClient(&http.Client{
		Transport: cliRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			ip := r.URL.Query().Get("ip")
			// Only return a result for the IP we expect the function to try.
			// The exact IP depends on the test machine's interfaces, so we
			// accept any non-empty IP query and return a match.
			if ip == "" {
				return cliJSONResponse(http.StatusOK, []netcup.ServerListMinimal{}), nil
			}
			return cliJSONResponse(http.StatusOK, []netcup.ServerListMinimal{
				{ID: 42, Name: "v220000000000000000"},
			}), nil
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}

	id, err := identifyServerByIP(context.Background(), client)
	if err != nil {
		t.Fatalf("identifyServerByIP() error = %v", err)
	}
	if id != 42 {
		t.Fatalf("id = %d, want 42", id)
	}
}

func TestIdentifyServerByIPReturnsErrorWhenNoMatch(t *testing.T) {
	client, err := netcup.NewClient("https://example.test/scp-core", netcup.WithHTTPClient(&http.Client{
		Transport: cliRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			return cliJSONResponse(http.StatusOK, []netcup.ServerListMinimal{}), nil
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = identifyServerByIP(context.Background(), client)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
