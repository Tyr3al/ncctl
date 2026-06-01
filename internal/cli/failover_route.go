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
	var ips []string
	var id int
	var family, serverRef string
	var wait bool
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Route one or more failover IPs to a server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if serverRef == "" {
				return fmt.Errorf("--server-id is required")
			}
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
			serverID, err := resolveServerID(ctx, client, serverRef)
			if err != nil {
				return err
			}
			tasks := make([]*netcup.TaskInfo, 0, max(1, len(ips)))
			if id != 0 {
				task, err := routeFailover(ctx, client, a.cfg.UserID, family, "", id, serverID)
				if err != nil {
					return err
				}
				if wait && task.UUID != "" {
					task, err = client.WaitTask(ctx, task.UUID, 2*time.Second)
					if err != nil {
						return err
					}
				}
				tasks = append(tasks, task)
			} else {
				for _, ip := range ips {
					task, err := routeFailover(ctx, client, a.cfg.UserID, family, ip, 0, serverID)
					if err != nil {
						return err
					}
					if wait && task.UUID != "" {
						task, err = client.WaitTask(ctx, task.UUID, 2*time.Second)
						if err != nil {
							return err
						}
					}
					tasks = append(tasks, task)
				}
			}
			return writeTasks(cmd, opts, tasks)
		},
	}
	cmd.Flags().StringArrayVar(&ips, "ip", nil, "failover IP or IPv6 prefix; repeat for multiple IPs")
	cmd.Flags().IntVar(&id, "id", 0, "failover IP ID")
	cmd.Flags().StringVar(&family, "family", "", "IP family: v4 or v6; inferred from --ip when omitted")
	cmd.Flags().StringVar(&serverRef, "server-id", "", "target server ID or name")
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the routing task to finish")
	return cmd
}

func writeTasks(cmd *cobra.Command, opts *options, tasks []*netcup.TaskInfo) error {
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), tasks)
	}
	rows := make([][]string, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			rows = append(rows, []string{"", "", "OK", ""})
			continue
		}
		rows = append(rows, []string{task.UUID, task.Name, task.State, stringPtrValue(task.Message)})
	}
	return writeTable(cmd.OutOrStdout(), []string{"UUID", "NAME", "STATE", "MESSAGE"}, rows)
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
