package cli

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyr3al/ncctl/pkg/netcup"
)

func newFailoverRouteCommand() *cobra.Command {
	var ip string
	var id, serverID int
	var family string
	var wait bool
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Route a failover IP to a server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if serverID == 0 {
				return fmt.Errorf("--server-id is required")
			}
			if id == 0 && ip == "" {
				return fmt.Errorf("one of --id or --ip is required")
			}
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
			task, err := routeFailover(ctx, client, a.cfg.UserID, family, ip, id, serverID)
			if err != nil {
				return err
			}
			if wait && task.UUID != "" {
				task, err = client.WaitTask(ctx, task.UUID, 2*time.Second)
				if err != nil {
					return err
				}
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), task)
			}
			return writeTable(cmd.OutOrStdout(), []string{"UUID", "NAME", "STATE", "MESSAGE"}, [][]string{{task.UUID, task.Name, task.State, stringPtrValue(task.Message)}})
		},
	}
	cmd.Flags().StringVar(&ip, "ip", "", "failover IP or IPv6 prefix")
	cmd.Flags().IntVar(&id, "id", 0, "failover IP ID")
	cmd.Flags().StringVar(&family, "family", "", "IP family: v4 or v6; inferred from --ip when omitted")
	cmd.Flags().IntVar(&serverID, "server-id", 0, "target server ID")
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the routing task to finish")
	return cmd
}

func routeFailover(ctx context.Context, client *netcup.Client, userID int, family, ip string, id, serverID int) (*netcup.TaskInfo, error) {
	if family == "" && ip != "" {
		family = inferIPFamily(ip)
	}
	if family == "" {
		family = "v4"
	}
	switch family {
	case "v4":
		if id == 0 {
			found, err := client.ListFailoverIPv4(ctx, userID, netcup.ListFailoverOptions{IP: ip})
			if err != nil {
				return nil, err
			}
			if len(found) == 0 {
				return nil, fmt.Errorf("no IPv4 failover IP found for %q", ip)
			}
			id = found[0].ID
		}
		return client.RouteFailoverIPv4(ctx, userID, id, serverID)
	case "v6":
		if id == 0 {
			found, err := client.ListFailoverIPv6(ctx, userID, netcup.ListFailoverOptions{IP: ip})
			if err != nil {
				return nil, err
			}
			if len(found) == 0 {
				return nil, fmt.Errorf("no IPv6 failover IP found for %q", ip)
			}
			id = found[0].ID
		}
		return client.RouteFailoverIPv6(ctx, userID, id, serverID)
	default:
		return nil, fmt.Errorf("family must be v4 or v6")
	}
}

func inferIPFamily(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed != nil && parsed.To4() != nil {
		return "v4"
	}
	if parsed != nil || strings.Contains(ip, ":") {
		return "v6"
	}
	if _, err := strconv.Atoi(ip); err == nil {
		return ""
	}
	return ""
}
