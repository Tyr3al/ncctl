package cli

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tyr3al/netcup-api/pkg/netcup"
)

func commandClient(cmd *cobra.Command, opts *options) (*netcup.Client, *netcup.RefreshTokenSource, context.Context, context.CancelFunc, error) {
	a, err := newApp(opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	client, source, err := a.apiClient()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := contextWithTimeout(cmd.Context(), opts.Timeout)
	return client, source, ctx, cancel, nil
}

func writeTask(cmd *cobra.Command, opts *options, task *netcup.TaskInfo) error {
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), task)
	}
	return writeTable(cmd.OutOrStdout(), []string{"UUID", "NAME", "STATE", "MESSAGE"}, [][]string{{task.UUID, task.Name, task.State, stringPtrValue(task.Message)}})
}

func newServerPowerCommand() *cobra.Command {
	var stateOption string
	cmd := &cobra.Command{
		Use:   "power <server-id> <ON|OFF|SUSPENDED>",
		Short: "Change server power state",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			opts, _ := commandOptions(cmd)
			if err := confirmRisky(cmd, opts, "Changing server power state"); err != nil {
				return err
			}
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.PatchServer(ctx, serverID, map[string]any{"state": args[1]}, stateOption)
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	cmd.Flags().StringVar(&stateOption, "state-option", "", "optional SCP state option")
	return cmd
}

func newServerUpdateCommand() *cobra.Command {
	var hostname, nickname string
	var autostart, uefi bool
	var setAutostart, setUEFI bool
	cmd := &cobra.Command{
		Use:   "update <server-id>",
		Short: "Update server attributes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			patch := map[string]any{}
			if hostname != "" {
				patch["hostname"] = hostname
			}
			if nickname != "" {
				patch["nickname"] = nickname
			}
			if setAutostart {
				patch["autostart"] = autostart
			}
			if setUEFI {
				patch["uefi"] = uefi
			}
			if len(patch) == 0 {
				return fmt.Errorf("no updates requested")
			}
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.PatchServer(ctx, serverID, patch, "")
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "new hostname")
	cmd.Flags().StringVar(&nickname, "nickname", "", "new nickname")
	cmd.Flags().BoolVar(&autostart, "autostart", false, "autostart value")
	cmd.Flags().BoolVar(&uefi, "uefi", false, "UEFI value")
	cmd.Flags().BoolVar(&setAutostart, "set-autostart", false, "update autostart")
	cmd.Flags().BoolVar(&setUEFI, "set-uefi", false, "update UEFI")
	return cmd
}

func addInterfaceWriteCommands(cmd *cobra.Command) {
	create := &cobra.Command{
		Use:   "create-vlan <server-id>",
		Short: "Create a VLAN interface",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			vlanID, _ := cmd.Flags().GetInt("vlan-id")
			driver, _ := cmd.Flags().GetString("driver")
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			iface, err := client.CreateInterfaceVLAN(ctx, serverID, vlanID, driver)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), iface)
		},
	}
	create.Flags().Int("vlan-id", 0, "VLAN ID")
	create.Flags().String("driver", "VIRTIO", "network driver")
	_ = create.MarkFlagRequired("vlan-id")

	update := &cobra.Command{
		Use:   "update <server-id> <mac>",
		Short: "Update an interface",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			driver, _ := cmd.Flags().GetString("driver")
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			iface, err := client.UpdateInterface(ctx, serverID, args[1], driver)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), iface)
		},
	}
	update.Flags().String("driver", "", "network driver")
	_ = update.MarkFlagRequired("driver")

	del := &cobra.Command{
		Use:   "delete <server-id> <mac>",
		Short: "Delete an interface",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			opts, _ := commandOptions(cmd)
			if err := confirmRisky(cmd, opts, "Deleting an interface"); err != nil {
				return err
			}
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			return client.DeleteInterface(ctx, serverID, args[1])
		},
	}
	cmd.AddCommand(create, update, del)
}

func newSnapshotsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "snapshots", Short: "Manage snapshots"}
	list := simpleServerListCommand("list <server-id>", "List snapshots", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
		return client.ListSnapshots(ctx, serverID)
	})
	create := &cobra.Command{
		Use:   "create <server-id> <name>",
		Short: "Create a snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			online, _ := cmd.Flags().GetBool("online")
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.CreateSnapshot(ctx, serverID, map[string]any{"name": args[1], "onlineSnapshot": online})
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	create.Flags().Bool("online", false, "create online snapshot")
	cmd.AddCommand(list, create, snapshotAction("delete", "Delete a snapshot", true), snapshotAction("revert", "Revert a snapshot", true), snapshotAction("export", "Export a snapshot", false))
	return cmd
}

func newRescueCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "rescue", Short: "Manage rescue system"}
	status := simpleServerListCommand("status <server-id>", "Show rescue status", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
		return client.GetRescueSystem(ctx, serverID)
	})
	enable := serverTaskCommand("enable <server-id>", "Enable rescue system", true, func(client *netcup.Client, ctx context.Context, serverID int) (*netcup.TaskInfo, error) {
		return client.ActivateRescueSystem(ctx, serverID)
	})
	disable := serverTaskCommand("disable <server-id>", "Disable rescue system", true, func(client *netcup.Client, ctx context.Context, serverID int) (*netcup.TaskInfo, error) {
		return client.DeactivateRescueSystem(ctx, serverID)
	})
	cmd.AddCommand(status, enable, disable)
	return cmd
}

func newDisksCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "disks", Short: "Manage disks"}
	list := simpleServerListCommand("list <server-id>", "List disks", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
		return client.ListDisks(ctx, serverID)
	})
	format := &cobra.Command{
		Use:   "format <server-id> <disk-name>",
		Short: "Format a disk",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			opts, _ := commandOptions(cmd)
			if err := confirmRisky(cmd, opts, "Formatting a disk will destroy data"); err != nil {
				return err
			}
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.FormatDisk(ctx, serverID, args[1])
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	cmd.AddCommand(list, format)
	return cmd
}

func newISOCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "iso", Short: "Manage ISO attachments"}
	cmd.AddCommand(
		simpleServerListCommand("attached <server-id>", "Show attached ISO", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
			return client.GetAttachedISO(ctx, serverID)
		}),
		simpleServerListCommand("list <server-id>", "List available ISO images", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
			return client.ListISOImages(ctx, serverID)
		}),
	)
	attach := &cobra.Command{
		Use:   "attach <server-id>",
		Short: "Attach an ISO",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			isoID, _ := cmd.Flags().GetInt("iso-id")
			userISO, _ := cmd.Flags().GetString("user-iso")
			boot, _ := cmd.Flags().GetBool("boot-cdrom")
			body := map[string]any{"changeBootDeviceToCdrom": boot}
			if isoID != 0 {
				body["isoId"] = isoID
			}
			if userISO != "" {
				body["userIsoName"] = userISO
			}
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.AttachISO(ctx, serverID, body)
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		},
	}
	attach.Flags().Int("iso-id", 0, "ISO image ID")
	attach.Flags().String("user-iso", "", "user ISO name")
	attach.Flags().Bool("boot-cdrom", false, "change boot device to CD-ROM")
	detach := serverTaskCommand("detach <server-id>", "Detach ISO", true, func(client *netcup.Client, ctx context.Context, serverID int) (*netcup.TaskInfo, error) {
		return client.DetachISO(ctx, serverID)
	})
	cmd.AddCommand(attach, detach)
	return cmd
}

func newFirewallCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "firewall", Short: "Manage firewalls"}
	policies := &cobra.Command{Use: "policies", Short: "Manage firewall policies"}
	policies.AddCommand(
		&cobra.Command{Use: "list", Short: "List firewall policies", RunE: func(cmd *cobra.Command, _ []string) error {
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
			policies, err := client.ListFirewallPolicies(ctx, a.cfg.UserID)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), policies)
		}},
	)
	iface := &cobra.Command{Use: "interface", Short: "Manage interface firewall"}
	iface.AddCommand(
		&cobra.Command{Use: "get <server-id> <mac>", Short: "Get interface firewall", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			firewall, err := client.GetInterfaceFirewall(ctx, serverID, args[1])
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), firewall)
		}},
		&cobra.Command{Use: "reapply <server-id> <mac>", Short: "Reapply interface firewall", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			opts, _ := commandOptions(cmd)
			if err := confirmRisky(cmd, opts, "Reapplying interface firewall"); err != nil {
				return err
			}
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			task, err := client.ReapplyInterfaceFirewall(ctx, serverID, args[1])
			if err != nil {
				return err
			}
			return writeTask(cmd, opts, task)
		}},
	)
	cmd.AddCommand(policies, iface)
	return cmd
}

func simpleServerListCommand(use, short string, run func(*netcup.Client, context.Context, int) (any, error)) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		serverID, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		opts, _ := commandOptions(cmd)
		client, _, ctx, cancel, err := commandClient(cmd, opts)
		if err != nil {
			return err
		}
		defer cancel()
		value, err := run(client, ctx, serverID)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), value)
	}}
}

func serverTaskCommand(use, short string, risky bool, run func(*netcup.Client, context.Context, int) (*netcup.TaskInfo, error)) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		serverID, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		opts, _ := commandOptions(cmd)
		if risky {
			if err := confirmRisky(cmd, opts, short); err != nil {
				return err
			}
		}
		client, _, ctx, cancel, err := commandClient(cmd, opts)
		if err != nil {
			return err
		}
		defer cancel()
		task, err := run(client, ctx, serverID)
		if err != nil {
			return err
		}
		return writeTask(cmd, opts, task)
	}}
}

func snapshotAction(name, short string, risky bool) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <server-id> <name>",
		Short: short,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, _ := strconv.Atoi(args[0])
			opts, _ := commandOptions(cmd)
			if risky {
				if err := confirmRisky(cmd, opts, short); err != nil {
					return err
				}
			}
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			switch name {
			case "delete":
				return client.DeleteSnapshot(ctx, serverID, args[1])
			case "revert":
				task, err := client.RevertSnapshot(ctx, serverID, args[1])
				if err != nil {
					return err
				}
				return writeTask(cmd, opts, task)
			default:
				task, err := client.ExportSnapshot(ctx, serverID, args[1])
				if err != nil {
					return err
				}
				return writeTask(cmd, opts, task)
			}
		},
	}
}

func newRDNSWriteCommands() []*cobra.Command {
	set := &cobra.Command{
		Use:   "set <ip> <hostname>",
		Short: "Set rDNS for an IP",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			if net.ParseIP(args[0]).To4() != nil {
				return client.SetRDNSIPv4(ctx, args[0], args[1])
			}
			return client.SetRDNSIPv6(ctx, args[0], args[1])
		},
	}
	del := &cobra.Command{
		Use:   "delete <ip>",
		Short: "Delete rDNS for an IP",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, _ := commandOptions(cmd)
			client, _, ctx, cancel, err := commandClient(cmd, opts)
			if err != nil {
				return err
			}
			defer cancel()
			if net.ParseIP(args[0]).To4() != nil {
				return client.DeleteRDNSIPv4(ctx, args[0])
			}
			return client.DeleteRDNSIPv6(ctx, args[0])
		},
	}
	return []*cobra.Command{set, del}
}
