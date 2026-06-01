package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tyr3al/netcup-api/internal/config"
	"github.com/tyr3al/netcup-api/pkg/netcup"
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
	return &app{opts: opts, cfg: cfg}, nil
}

func (a *app) save() error {
	return config.Save(a.opts.ConfigPath, a.cfg)
}

func (a *app) authClient() (*netcup.AuthClient, error) {
	return netcup.NewAuthClient(a.cfg.AuthBaseURL, netcup.WithAuthHTTPClient(&http.Client{Timeout: a.opts.Timeout}))
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
	client, err := netcup.NewClient(a.cfg.APIBaseURL, netcup.WithHTTPClient(&http.Client{Timeout: a.opts.Timeout}), netcup.WithTokenSource(source))
	if err != nil {
		return nil, nil, err
	}
	return client, source, nil
}

func contextWithTimeout(cmdCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if cmdCtx == nil {
		cmdCtx = context.Background()
	}
	return context.WithTimeout(cmdCtx, timeout)
}
