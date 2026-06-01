package cli

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyr3al/netcup-api/pkg/netcup"
)

func newServersCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "servers", Short: "Manage servers"}
	var limit int
	list := &cobra.Command{
		Use:   "list",
		Short: "List servers",
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
			servers, err := client.ListServers(ctx, netcup.ListServersOptions{Limit: limit})
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), servers)
			}
			rows := make([][]string, 0, len(servers))
			for _, server := range servers {
				rows = append(rows, []string{
					strconv.Itoa(server.ID),
					server.Name,
					stringPtrValue(server.Hostname),
					stringPtrValue(server.Nickname),
					strconv.FormatBool(server.Disabled),
				})
			}
			return writeTable(cmd.OutOrStdout(), []string{"ID", "NAME", "HOSTNAME", "NICKNAME", "DISABLED"}, rows)
		},
	}
	list.Flags().IntVar(&limit, "limit", 100, "maximum number of servers to fetch")
	get := &cobra.Command{
		Use:   "get <server-id>",
		Short: "Get a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid server id: %w", err)
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
			server, err := client.GetServer(ctx, serverID, true)
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), server)
			}
			return writeTable(cmd.OutOrStdout(), []string{"ID", "NAME", "HOSTNAME", "NICKNAME", "STATE"}, [][]string{{
				strconv.Itoa(server.ID),
				server.Name,
				stringPtrValue(server.Hostname),
				stringPtrValue(server.Nickname),
				serverState(server),
			}})
		},
	}
	cmd.AddCommand(list, get)
	return cmd
}

func newInterfacesCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "interfaces", Short: "Manage server interfaces"}
	list := &cobra.Command{
		Use:   "list <server-id>",
		Short: "List server interfaces",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid server id: %w", err)
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
			ifaces, err := client.ListInterfaces(ctx, serverID, true)
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), ifaces)
			}
			rows := make([][]string, 0, len(ifaces))
			for _, iface := range ifaces {
				rows = append(rows, []string{iface.MAC, iface.Driver, strconv.Itoa(iface.SpeedInMBits), strconv.Itoa(len(iface.IPv4Addresses)), strconv.Itoa(len(iface.IPv6Addresses))})
			}
			return writeTable(cmd.OutOrStdout(), []string{"MAC", "DRIVER", "SPEED_MBIT", "IPV4", "IPV6"}, rows)
		},
	}
	get := &cobra.Command{
		Use:   "get <server-id> <mac>",
		Short: "Get a server interface",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid server id: %w", err)
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
			iface, err := client.GetInterface(ctx, serverID, args[1], true)
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), iface)
			}
			return writeTable(cmd.OutOrStdout(), []string{"MAC", "DRIVER", "SPEED_MBIT", "IPV4", "IPV6"}, [][]string{{
				iface.MAC, iface.Driver, strconv.Itoa(iface.SpeedInMBits), strconv.Itoa(len(iface.IPv4Addresses)), strconv.Itoa(len(iface.IPv6Addresses)),
			}})
		},
	}
	cmd.AddCommand(list, get)
	return cmd
}

func newFailoverCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "failover", Short: "Manage failover IPs"}
	var family, ip string
	var serverID int
	list := &cobra.Command{
		Use:   "list",
		Short: "List failover IPs",
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
			filter := netcup.ListFailoverOptions{IP: ip, ServerID: serverID}
			switch family {
			case "", "all":
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
			case "v4":
				v4, err := client.ListFailoverIPv4(ctx, a.cfg.UserID, filter)
				if err != nil {
					return err
				}
				if opts.JSON {
					return writeJSON(cmd.OutOrStdout(), v4)
				}
				return writeFailoverTable(cmd, v4, nil)
			case "v6":
				v6, err := client.ListFailoverIPv6(ctx, a.cfg.UserID, filter)
				if err != nil {
					return err
				}
				if opts.JSON {
					return writeJSON(cmd.OutOrStdout(), v6)
				}
				return writeFailoverTable(cmd, nil, v6)
			default:
				return fmt.Errorf("family must be v4, v6, or all")
			}
		},
	}
	list.Flags().StringVar(&family, "family", "all", "IP family: all, v4, or v6")
	list.Flags().StringVar(&ip, "ip", "", "filter by IP")
	list.Flags().IntVar(&serverID, "server-id", 0, "filter by server ID")
	route := newFailoverRouteCommand()
	cmd.AddCommand(list, route)
	return cmd
}

func newTasksCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "tasks", Short: "Manage tasks"}
	var limit int
	list := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
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
			tasks, err := client.ListTasks(ctx, netcup.ListTasksOptions{Limit: limit})
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), tasks)
			}
			rows := make([][]string, 0, len(tasks))
			for _, task := range tasks {
				rows = append(rows, []string{task.UUID, task.Name, task.State, stringPtrValue(task.Message)})
			}
			return writeTable(cmd.OutOrStdout(), []string{"UUID", "NAME", "STATE", "MESSAGE"}, rows)
		},
	}
	list.Flags().IntVar(&limit, "limit", 100, "maximum number of tasks to fetch")
	get := &cobra.Command{
		Use:   "get <uuid>",
		Short: "Get a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			task, err := client.GetTask(ctx, args[0])
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), task)
			}
			return writeTable(cmd.OutOrStdout(), []string{"UUID", "NAME", "STATE", "MESSAGE"}, [][]string{{task.UUID, task.Name, task.State, stringPtrValue(task.Message)}})
		},
	}
	wait := &cobra.Command{
		Use:   "wait <uuid>",
		Short: "Wait for a task to finish",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			task, err := client.WaitTask(ctx, args[0], 2*time.Second)
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), task)
			}
			return writeTable(cmd.OutOrStdout(), []string{"UUID", "NAME", "STATE", "MESSAGE"}, [][]string{{task.UUID, task.Name, task.State, stringPtrValue(task.Message)}})
		},
	}
	cmd.AddCommand(list, get, wait)
	return cmd
}

func newRDNSCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "rdns", Short: "Manage reverse DNS"}
	get := &cobra.Command{
		Use:   "get <ip>",
		Short: "Get rDNS for an IP",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			parsed := net.ParseIP(args[0])
			if parsed == nil {
				return fmt.Errorf("invalid IP address %q", args[0])
			}
			if parsed.To4() != nil {
				rdns, err := client.GetRDNSIPv4(ctx, args[0])
				if err != nil {
					return err
				}
				if opts.JSON {
					return writeJSON(cmd.OutOrStdout(), map[string]string{"ip": args[0], "rdns": rdns})
				}
				return writeTable(cmd.OutOrStdout(), []string{"IP", "RDNS"}, [][]string{{args[0], rdns}})
			}
			rdns, err := client.GetRDNSIPv6(ctx, args[0])
			if err != nil {
				return err
			}
			if opts.JSON {
				return writeJSON(cmd.OutOrStdout(), rdns)
			}
			rows := make([][]string, 0, len(rdns))
			for ip, host := range rdns {
				rows = append(rows, []string{ip, host})
			}
			return writeTable(cmd.OutOrStdout(), []string{"IP", "RDNS"}, rows)
		},
	}
	cmd.AddCommand(get)
	return cmd
}

func serverState(server *netcup.Server) string {
	if server.ServerLiveInfo == nil {
		return ""
	}
	return server.ServerLiveInfo.State
}

func writeFailoverTable(cmd *cobra.Command, v4 []netcup.FailoverIPv4, v6 []netcup.FailoverIPv6) error {
	rows := make([][]string, 0, len(v4)+len(v6))
	for _, item := range v4 {
		rows = append(rows, []string{"v4", strconv.Itoa(item.ID), item.IP, strconv.Itoa(item.Server.ID), item.Server.Name, item.Site.City, strconv.FormatBool(item.Editable)})
	}
	for _, item := range v6 {
		prefix := item.NetworkPrefix
		if item.NetworkPrefixLength != 0 {
			prefix += "/" + strconv.Itoa(item.NetworkPrefixLength)
		}
		rows = append(rows, []string{"v6", strconv.Itoa(item.ID), prefix, strconv.Itoa(item.Server.ID), item.Server.Name, item.Site.City, strconv.FormatBool(item.Editable)})
	}
	return writeTable(cmd.OutOrStdout(), []string{"FAMILY", "ID", "IP", "SERVER_ID", "SERVER", "SITE", "EDITABLE"}, rows)
}
