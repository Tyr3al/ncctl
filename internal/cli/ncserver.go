package cli

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyr3al/ncctl/pkg/netcup"
)

// NewServerRootCommand creates the ncserver command tree.
// It exposes a reduced set of commands scoped to the server it runs on.
func NewServerRootCommand() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "ncserver",
		Short: "Manage this server's netcup SCP resources",
		Long: `ncserver is a minimal SCP client intended to run on a netcup server.

It identifies itself via the SCP API and exposes only the commands relevant
to managing the server it runs on. Use ncctl for full administrative access.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "config file path")
	cmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", 0, "overall operation timeout; 0 means no limit")
	cmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "write JSON output")
	cmd.PersistentFlags().BoolVarP(&opts.Yes, "yes", "y", false, "confirm risky operations")
	// Base URLs are intentionally hidden; defaults cover the vast majority of cases.
	cmd.PersistentFlags().StringVar(&opts.APIBaseURL, "api-base-url", defaultAPIBaseURL, "SCP API base URL")
	cmd.PersistentFlags().StringVar(&opts.AuthBaseURL, "auth-base-url", defaultAuthBaseURL, "SCP auth base URL")
	_ = cmd.PersistentFlags().MarkHidden("api-base-url")
	_ = cmd.PersistentFlags().MarkHidden("auth-base-url")

	attachOptions(cmd, opts)
	cmd.AddCommand(
		newLoginCommand(),
		newLogoutCommand(),
		newWhoamiCommand(),
		newRenewCommand(),
		newIdentifyCommand(),
		newSelfStatusCommand(),
		newSelfFailoverCommand(),
		newSelfRescueCommand(),
		newSelfSnapshotsCommand(),
		newSelfRDNSCommand(),
		newSelfTasksWaitCommand(),
	)
	return cmd
}

// selfCommandClient resolves the server ID from config and returns a ready-to-use
// app, client, context, cancel, and server ID. Returns an error if not identified.
func selfCommandClient(cmd *cobra.Command, opts *options) (*app, *netcup.Client, context.Context, context.CancelFunc, int, error) {
	a, err := newApp(opts)
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}
	if a.cfg.ServerID == 0 {
		return nil, nil, nil, nil, 0, fmt.Errorf("server not identified; run: ncserver identify")
	}
	client, source, err := a.apiClient()
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}
	ctx, cancel := contextWithTimeout(cmd.Context(), opts.Timeout)
	return a, client, ctx, func() {
		cancel()
		a.persistRefreshToken(source)
	}, a.cfg.ServerID, nil
}

// newRenewCommand forces a token refresh and persists the updated refresh token.
// It should be run periodically to prevent the offline token from expiring.
func newRenewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "renew",
		Short: "Refresh and persist the stored token",
		Long: `Forces a token refresh and writes the updated refresh token to the config file.

Offline tokens issued by the SCP auth server expire after a period of inactivity
(typically 30 days). Run this command weekly via the ncserver-token-renew systemd
timer to keep the token alive.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			a, err := newApp(opts)
			if err != nil {
				return err
			}
			if a.cfg.Refresh == "" {
				return fmt.Errorf("not logged in; run: ncserver login")
			}
			auth, err := a.authClient()
			if err != nil {
				return err
			}
			source := netcup.NewRefreshTokenSource(auth, a.cfg.Refresh)
			ctx, cancel := contextWithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			if _, err := source.Token(ctx); err != nil {
				return fmt.Errorf("token refresh failed: %w", err)
			}
			newToken := source.RefreshToken()
			if newToken != a.cfg.Refresh {
				a.cfg.Refresh = newToken
				if err := a.save(); err != nil {
					return fmt.Errorf("save config: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Token refreshed and saved.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Token refreshed.")
			}
			return nil
		},
	}
}

// newIdentifyCommand queries the SCP API to find this server by matching its
// local IP addresses, then caches the server ID in the config file.
func newIdentifyCommand() *cobra.Command {
	var serverRef string
	cmd := &cobra.Command{
		Use:   "identify",
		Short: "Identify this server and cache its ID",
		Long: `Queries the SCP API to find this server by matching local network interface
addresses against the servers in your account. The resolved server ID is
saved to the config file and used by all other ncserver commands.

If auto-detection fails, provide the server ID or name explicitly:

  ncserver identify --server-id v2202508149564377314`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			a, err := newApp(opts)
			if err != nil {
				return err
			}
			client, _, err := a.apiClient()
			if err != nil {
				return err
			}
			ctx, cancel := contextWithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()

			var serverID int
			if serverRef != "" {
				serverID, err = resolveServerID(ctx, client, serverRef)
				if err != nil {
					return err
				}
			} else {
				serverID, err = identifyServerByIP(ctx, client)
				if err != nil {
					return fmt.Errorf("%w\n\nhint: specify the server manually with --server-id", err)
				}
			}

			server, err := client.GetServer(ctx, serverID, false)
			if err != nil {
				return err
			}

			a.cfg.ServerID = serverID
			if err := a.save(); err != nil {
				return err
			}

			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), server)
			}
			name := server.Name
			if server.Nickname != nil && *server.Nickname != "" {
				name = fmt.Sprintf("%s (%s)", server.Name, *server.Nickname)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Identified as %s, ID %d\n", name, server.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&serverRef, "server-id", "", "server ID or name to use instead of auto-detection")
	return cmd
}

// identifyServerByIP iterates over non-loopback local IP addresses and queries
// the SCP API for each until a matching server is found.
func identifyServerByIP(ctx context.Context, client *netcup.Client) (int, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, fmt.Errorf("list network interfaces: %w", err)
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			continue
		}
		servers, err := client.ListServers(ctx, netcup.ListServersOptions{IP: ip.String(), Limit: 1})
		if err != nil || len(servers) == 0 {
			continue
		}
		return servers[0].ID, nil
	}
	return 0, fmt.Errorf("could not identify this server via any local IP address")
}

func newSelfStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show this server's status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			_, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			server, err := client.GetServer(ctx, serverID, true)
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), server)
			}
			return writeTable(cmd.OutOrStdout(),
				[]string{"ID", "NAME", "HOSTNAME", "NICKNAME", "STATE"},
				[][]string{{
					strconv.Itoa(server.ID),
					server.Name,
					stringPtrValue(server.Hostname),
					stringPtrValue(server.Nickname),
					serverState(server),
				}},
			)
		},
	}
}

func newSelfFailoverCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "failover", Short: "Manage failover IPs for this server"}
	cmd.AddCommand(newSelfFailoverListCommand(), newSelfFailoverRouteCommand())
	return cmd
}

func newSelfFailoverListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List failover IPs routed to this server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			a, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			filter := netcup.ListFailoverOptions{ServerID: serverID}
			v4, err := client.ListFailoverIPv4(ctx, a.cfg.UserID, filter)
			if err != nil {
				return err
			}
			v6, err := client.ListFailoverIPv6(ctx, a.cfg.UserID, filter)
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), map[string]any{"ipv4": v4, "ipv6": v6})
			}
			return writeFailoverTable(cmd, v4, v6)
		},
	}
}

func newSelfFailoverRouteCommand() *cobra.Command {
	var ips []string
	var id int
	var family string
	var wait bool
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Route one or more failover IPs to this server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if id == 0 && len(ips) == 0 {
				return fmt.Errorf("one of --id or --ip is required")
			}
			if id != 0 && len(ips) > 0 {
				return fmt.Errorf("--id and --ip cannot be combined")
			}
			if id != 0 && family == "" {
				return fmt.Errorf("--family is required when routing by --id")
			}
			opts, _ := commandOptions(cmd)
			a, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()

			tasks := make([]*netcup.TaskInfo, 0, max(1, len(ips)))
			targets := ips
			if id != 0 {
				targets = []string{""}
			}
			for _, ip := range targets {
				routeID := id
				task, err := routeFailover(ctx, client, a.cfg.UserID, family, ip, routeID, serverID)
				if err != nil {
					return err
				}
				if wait && task != nil && task.UUID != "" {
					task, err = client.WaitTask(ctx, task.UUID, 2*time.Second)
					if err != nil {
						return err
					}
				}
				tasks = append(tasks, task)
			}
			return writeTasks(cmd, opts, tasks)
		},
	}
	cmd.Flags().StringArrayVar(&ips, "ip", nil, "failover IP or IPv6 prefix; repeat for multiple IPs")
	cmd.Flags().IntVar(&id, "id", 0, "failover IP ID")
	cmd.Flags().StringVar(&family, "family", "", "IP family: v4 or v6; inferred from --ip when omitted")
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the routing task to finish")
	return cmd
}

func newSelfRescueCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "rescue", Short: "Manage this server's rescue system"}
	status := &cobra.Command{
		Use:   "status",
		Short: "Show rescue system status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			_, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			info, err := client.GetRescueSystem(ctx, serverID)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), info)
		},
	}
	enable := &cobra.Command{
		Use:   "enable",
		Short: "Enable rescue system",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			if err := confirmRisky(cmd, opts, "Enabling rescue system"); err != nil {
				return err
			}
			_, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.ActivateRescueSystem(ctx, serverID)
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	disable := &cobra.Command{
		Use:   "disable",
		Short: "Disable rescue system",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			if err := confirmRisky(cmd, opts, "Disabling rescue system"); err != nil {
				return err
			}
			_, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.DeactivateRescueSystem(ctx, serverID)
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	cmd.AddCommand(status, enable, disable)
	return cmd
}

func newSelfSnapshotsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "snapshots", Short: "Manage this server's snapshots"}
	list := &cobra.Command{
		Use:   "list",
		Short: "List snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, _ := commandOptions(cmd)
			_, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			snapshots, err := client.ListSnapshots(ctx, serverID)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), snapshots)
		},
	}
	var online bool
	create := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, _ := commandOptions(cmd)
			_, client, ctx, cancel, serverID, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.CreateSnapshot(ctx, serverID, map[string]any{
				"name":           args[0],
				"onlineSnapshot": online,
			})
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	create.Flags().BoolVar(&online, "online", false, "create snapshot while server is running")
	cmd.AddCommand(list, create)
	return cmd
}

func newSelfRDNSCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "rdns", Short: "Manage reverse DNS for this server's IPs"}
	cmd.AddCommand(newRDNSCommand().Commands()...)
	return cmd
}

func newSelfTasksWaitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tasks wait <uuid>",
		Short: "Wait for an async task to finish",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, _ := commandOptions(cmd)
			_, client, ctx, cancel, _, err := selfCommandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.WaitTask(ctx, args[0], 2*time.Second)
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
}
