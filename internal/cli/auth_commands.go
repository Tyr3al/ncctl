package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyr3al/ncctl/internal/config"
	"github.com/tyr3al/ncctl/pkg/netcup"
)

func newLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with netcup SCP",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			a, err := newApp(opts)
			if err != nil {
				return err
			}
			ctx, cancel := contextWithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			auth, err := a.authClient()
			if err != nil {
				return err
			}
			device, err := auth.StartDeviceAuthorization(ctx)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Open: %s\n", device.VerificationURIComplete)
			fmt.Fprintf(out, "Code: %s\n", device.UserCode)
			token, err := pollDeviceToken(cmd.Context(), auth, device, out)
			if err != nil {
				return err
			}
			info, err := auth.UserInfo(cmd.Context(), token.AccessToken)
			if err != nil {
				return err
			}
			userID, err := strconv.Atoi(info.ID)
			if err != nil {
				return fmt.Errorf("unexpected user ID %q in server response", info.ID)
			}
			a.cfg.APIBaseURL = opts.APIBaseURL
			a.cfg.AuthBaseURL = opts.AuthBaseURL
			a.cfg.UserID = userID
			a.cfg.Refresh = token.RefreshToken
			if err := a.save(); err != nil {
				return err
			}
			fmt.Fprintln(out, "Login successful")
			return nil
		},
	}
}

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored ncctl credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			if err := config.Remove(opts.ConfigPath); err != nil {
				return err
			}
			_, err := io.WriteString(cmd.OutOrStdout(), "Logged out\n")
			return err
		},
	}
}

func newWhoamiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Print authenticated SCP user information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			a, err := newApp(opts)
			if err != nil {
				return err
			}
			if a.cfg.Refresh == "" {
				return fmt.Errorf("not logged in; run ncctl login")
			}
			ctx, cancel := contextWithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			auth, err := a.authClient()
			if err != nil {
				return err
			}
			source := netcup.NewRefreshTokenSource(auth, a.cfg.Refresh)
			token, err := source.Token(ctx)
			if err != nil {
				return err
			}
			info, err := auth.UserInfo(ctx, token)
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), info)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ID\tUSERNAME\tEMAIL\n%s\t%s\t%s\n", info.ID, info.Username, info.Email)
			return nil
		},
	}
}

func pollDeviceToken(ctx context.Context, auth *netcup.AuthClient, device *netcup.DeviceAuthorization, out io.Writer) (*netcup.TokenResponse, error) {
	deadline := time.Now().Add(device.ExpiresAfter())
	if device.ExpiresAfter() <= 0 {
		deadline = time.Now().Add(10 * time.Minute)
	}
	ticker := time.NewTicker(device.PollInterval())
	defer ticker.Stop()
	for {
		token, err := auth.ExchangeDeviceCode(ctx, device.DeviceCode)
		if err == nil {
			return token, nil
		}
		if errors.Is(err, netcup.ErrAuthorizationDenied) || errors.Is(err, netcup.ErrAuthorizationExpired) {
			return nil, err
		}
		if !errors.Is(err, netcup.ErrAuthorizationPending) {
			return nil, err
		}
		fmt.Fprintln(out, "Waiting for browser authorization...")
		if time.Now().After(deadline) {
			return nil, netcup.ErrAuthorizationExpired
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
