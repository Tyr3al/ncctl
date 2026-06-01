package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tyr3al/ncctl/internal/config"
	"github.com/tyr3al/ncctl/pkg/netcup"
)

type app struct {
	opts *options
	cfg  *config.Config
}

func newApp(opts *options) (*app, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, err
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = opts.APIBaseURL
	}
	if cfg.AuthBaseURL == "" {
		cfg.AuthBaseURL = opts.AuthBaseURL
	}
	for _, rawURL := range []string{cfg.APIBaseURL, cfg.AuthBaseURL} {
		if rawURL == "" {
			continue
		}
		u, err := url.Parse(rawURL)
		if err != nil || u.Scheme != "https" {
			return nil, fmt.Errorf("URL %q must use https", rawURL)
		}
	}
	return &app{opts: opts, cfg: cfg}, nil
}

func (a *app) save() error {
	return config.Save(a.opts.ConfigPath, a.cfg)
}

// requestTimeout caps individual HTTP requests. It is intentionally separate
// from --timeout, which controls the overall operation deadline (e.g. a wait loop).
const requestTimeout = 30 * time.Second

func (a *app) authClient() (*netcup.AuthClient, error) {
	return netcup.NewAuthClient(a.cfg.AuthBaseURL, netcup.WithAuthHTTPClient(&http.Client{Timeout: requestTimeout}))
}

func (a *app) apiClient() (*netcup.Client, *netcup.RefreshTokenSource, error) {
	if a.cfg.Refresh == "" {
		return nil, nil, fmt.Errorf("not logged in; run ncctl login")
	}
	auth, err := a.authClient()
	if err != nil {
		return nil, nil, err
	}
	source := netcup.NewRefreshTokenSource(auth, a.cfg.Refresh)
	client, err := netcup.NewClient(a.cfg.APIBaseURL, netcup.WithHTTPClient(&http.Client{Timeout: requestTimeout}), netcup.WithTokenSource(source))
	if err != nil {
		return nil, nil, err
	}
	return client, source, nil
}

// contextWithTimeout returns a context bounded by timeout. A timeout of 0 means
// no deadline — useful when --timeout is not set and a wait loop must run freely.
func contextWithTimeout(cmdCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if cmdCtx == nil {
		cmdCtx = context.Background()
	}
	if timeout <= 0 {
		return context.WithCancel(cmdCtx)
	}
	return context.WithTimeout(cmdCtx, timeout)
}
