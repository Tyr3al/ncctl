package cli

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tyr3al/ncctl/pkg/netcup"
)

func newSystemCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "system", Short: "Inspect SCP system endpoints"}
	var mcpBody, mcpBodyFile string
	mcp := &cobra.Command{Use: "openapi-mcp", Short: "Call the OpenAPI MCP endpoint", RunE: func(cmd *cobra.Command, _ []string) error {
		body, err := parseBodyFlags(mcpBody, mcpBodyFile)
		if err != nil {
			return err
		}
		client, _, ctx, cancel, err := commandClientFromCommand(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		out, err := client.OpenAPIMCP(ctx, body)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), out)
	}}
	mcp.Flags().StringVar(&mcpBody, "body", "", "JSON request body")
	mcp.Flags().StringVar(&mcpBodyFile, "body-file", "", "file containing JSON request body")
	cmd.AddCommand(
		&cobra.Command{Use: "ping", Short: "Check API availability", RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, ctx, cancel, err := commandClientFromCommand(cmd)
			if err != nil {
				return err
			}
			defer cancel()
			if err := client.Ping(ctx); err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), map[string]string{"status": "ok"})
		}},
		jsonCommand("maintenance", "Show maintenance information", func(cmd *cobra.Command, _ []string) (any, error) {
			client, _, ctx, cancel, err := commandClientFromCommand(cmd)
			if err != nil {
				return nil, err
			}
			defer cancel()
			return client.GetMaintenance(ctx)
		}),
		jsonCommand("openapi", "Show the SCP OpenAPI document", func(cmd *cobra.Command, _ []string) (any, error) {
			client, _, ctx, cancel, err := commandClientFromCommand(cmd)
			if err != nil {
				return nil, err
			}
			defer cancel()
			return client.GetOpenAPI(ctx)
		}),
		mcp,
	)
	return cmd
}

func newServerExtrasCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "server", Short: "Additional server operations"}
	cmd.AddCommand(
		simpleServerListCommand("gpu-driver <server>", "Get GPU driver download information", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
			return client.GetGPUDriver(ctx, serverID)
		}),
		simpleServerListCommand("guest-agent <server>", "Get guest agent data", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
			return client.GetGuestAgent(ctx, serverID)
		}),
		simpleServerListCommand("guest-agent-status <server>", "Get guest agent status", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
			return client.GetGuestAgentStatus(ctx, serverID)
		}),
		newServerLogsCommand(),
		newMetricsCommand(),
		newImageCommand(),
		newStorageOptimizationCommand(),
	)
	return cmd
}

func newServerLogsCommand() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{Use: "logs <server>", Short: "Get server logs", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		opts, _ := commandOptions(cmd)
		client, ctx, cancel, serverID, err := commandServerID(cmd, opts, args[0])
		if err != nil {
			return err
		}
		defer cancel()
		logs, err := client.GetServerLogs(ctx, serverID, limit, offset)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), logs)
	}}
	cmd.Flags().IntVar(&limit, "limit", 100, "maximum number of logs")
	cmd.Flags().IntVar(&offset, "offset", 0, "log offset")
	return cmd
}

func newMetricsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "metrics", Short: "Get server metrics"}
	for _, metric := range []string{"cpu", "disk", "network", "network/packet"} {
		metric := metric
		useName := strings.ReplaceAll(metric, "/", "-")
		var hours int
		sub := &cobra.Command{Use: useName + " <server>", Short: "Get " + metric + " metrics", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			opts, _ := commandOptions(cmd)
			client, ctx, cancel, serverID, err := commandServerID(cmd, opts, args[0])
			if err != nil {
				return err
			}
			defer cancel()
			data, err := client.GetServerMetric(ctx, serverID, metric, hours)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), data)
		}}
		sub.Flags().IntVar(&hours, "hours", 24, "number of hours")
		cmd.AddCommand(sub)
	}
	return cmd
}

func newImageCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "image", Short: "Manage server image setup"}
	cmd.AddCommand(
		simpleServerListCommand("flavours <server>", "List image flavours", func(client *netcup.Client, ctx context.Context, serverID int) (any, error) {
			return client.ListImageFlavours(ctx, serverID)
		}),
		imageSetupCommand("setup <server>", "Setup an image", true, func(client *netcup.Client, ctx context.Context, serverID int, body map[string]any) (*netcup.TaskInfo, error) {
			return client.SetupImage(ctx, serverID, body)
		}),
		imageSetupCommand("setup-user <server>", "Setup a user image", true, func(client *netcup.Client, ctx context.Context, serverID int, body map[string]any) (*netcup.TaskInfo, error) {
			return client.SetupUserImage(ctx, serverID, body)
		}),
	)
	return cmd
}

func imageSetupCommand(use, short string, risky bool, run func(*netcup.Client, context.Context, int, map[string]any) (*netcup.TaskInfo, error)) *cobra.Command {
	var jsonBody, jsonFile string
	cmd := &cobra.Command{Use: use, Short: short, Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		opts, _ := commandOptions(cmd)
		if risky {
			if err := confirmRisky(cmd, opts, short); err != nil {
				return err
			}
		}
		body, err := parseBodyFlags(jsonBody, jsonFile)
		if err != nil {
			return err
		}
		client, ctx, cancel, serverID, err := commandServerID(cmd, opts, args[0])
		if err != nil {
			return err
		}
		defer cancel()
		task, err := run(client, ctx, serverID, body)
		if err != nil {
			return err
		}
		return writeTask(cmd, opts, task)
	}}
	cmd.Flags().StringVar(&jsonBody, "body", "", "JSON request body")
	cmd.Flags().StringVar(&jsonFile, "body-file", "", "file containing JSON request body")
	return cmd
}

func newStorageOptimizationCommand() *cobra.Command {
	var disks []string
	var startAfter bool
	cmd := &cobra.Command{Use: "storage-optimization <server>", Short: "Optimize server storage", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		opts, _ := commandOptions(cmd)
		if err := confirmRisky(cmd, opts, "Optimizing server storage"); err != nil {
			return err
		}
		client, ctx, cancel, serverID, err := commandServerID(cmd, opts, args[0])
		if err != nil {
			return err
		}
		defer cancel()
		task, err := client.OptimizeStorage(ctx, serverID, disks, startAfter)
		if err != nil {
			return err
		}
		return writeTask(cmd, opts, task)
	}}
	cmd.Flags().StringArrayVar(&disks, "disk", nil, "disk to optimize; repeat for multiple disks")
	cmd.Flags().BoolVar(&startAfter, "start-after", false, "start server after optimization")
	return cmd
}

func newUserCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "user", Short: "Manage SCP user resources"}
	cmd.AddCommand(
		jsonCommand("get", "Get current SCP user", func(cmd *cobra.Command, _ []string) (any, error) {
			a, client, ctx, cancel, err := appClientFromCommand(cmd)
			if err != nil {
				return nil, err
			}
			defer cancel()
			return client.GetUser(ctx, a.cfg.UserID)
		}),
		newUserUpdateCommand(),
		newUserLogsCommand(),
		newUserImagesCommand(),
		newUserISOsCommand(),
		newSSHKeysCommand(),
		newVLansCommand(),
	)
	return cmd
}

func newUserUpdateCommand() *cobra.Command {
	var jsonBody, jsonFile string
	cmd := &cobra.Command{Use: "update", Short: "Update current SCP user", RunE: func(cmd *cobra.Command, _ []string) error {
		opts, _ := commandOptions(cmd)
		if err := confirmRisky(cmd, opts, "Updating user settings"); err != nil {
			return err
		}
		body, err := parseBodyFlags(jsonBody, jsonFile)
		if err != nil {
			return err
		}
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		out, err := client.UpdateUser(ctx, a.cfg.UserID, body)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), out)
	}}
	cmd.Flags().StringVar(&jsonBody, "body", "", "JSON request body")
	cmd.Flags().StringVar(&jsonFile, "body-file", "", "file containing JSON request body")
	return cmd
}

func newUserLogsCommand() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{Use: "logs", Short: "Get user logs", RunE: func(cmd *cobra.Command, _ []string) error {
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		logs, err := client.GetUserLogs(ctx, a.cfg.UserID, limit, offset)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), logs)
	}}
	cmd.Flags().IntVar(&limit, "limit", 100, "maximum number of logs")
	cmd.Flags().IntVar(&offset, "offset", 0, "log offset")
	return cmd
}

func newUserImagesCommand() *cobra.Command {
	return userAssetCommand("images", "image")
}

func newUserISOsCommand() *cobra.Command {
	return userAssetCommand("isos", "iso")
}

func userAssetCommand(kind, singular string) *cobra.Command {
	cmd := &cobra.Command{Use: kind, Short: "Manage user " + kind}
	cmd.AddCommand(
		jsonCommand("list", "List user "+kind, func(cmd *cobra.Command, _ []string) (any, error) {
			a, client, ctx, cancel, err := appClientFromCommand(cmd)
			if err != nil {
				return nil, err
			}
			defer cancel()
			if kind == "images" {
				return client.ListUserImages(ctx, a.cfg.UserID)
			}
			return client.ListUserISOs(ctx, a.cfg.UserID)
		}),
		userAssetGetCommand(kind, singular),
		userAssetPrepareCommand(kind, singular),
		userAssetPartCommand(kind, singular),
		userAssetCompleteCommand(kind, singular),
		userAssetDeleteCommand(kind, singular),
	)
	return cmd
}

func userAssetGetCommand(kind, singular string) *cobra.Command {
	return jsonCommand("get <key>", "Get user "+singular, func(cmd *cobra.Command, args []string) (any, error) {
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return nil, err
		}
		defer cancel()
		if kind == "images" {
			return client.GetUserImage(ctx, a.cfg.UserID, args[0])
		}
		return client.GetUserISO(ctx, a.cfg.UserID, args[0])
	}, cobra.ExactArgs(1))
}

func userAssetPrepareCommand(kind, singular string) *cobra.Command {
	var multipart bool
	cmd := jsonCommand("prepare-upload <key>", "Prepare "+singular+" upload", func(cmd *cobra.Command, args []string) (any, error) {
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return nil, err
		}
		defer cancel()
		if kind == "images" {
			return client.PrepareUserImageUpload(ctx, a.cfg.UserID, args[0], multipart)
		}
		return client.PrepareUserISOUpload(ctx, a.cfg.UserID, args[0], multipart)
	}, cobra.ExactArgs(1))
	cmd.Flags().BoolVar(&multipart, "multipart", false, "prepare multipart upload")
	return cmd
}

func userAssetPartCommand(kind, singular string) *cobra.Command {
	var part int
	cmd := jsonCommand("part-url <key> <upload-id>", "Get "+singular+" part upload URL", func(cmd *cobra.Command, args []string) (any, error) {
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return nil, err
		}
		defer cancel()
		if kind == "images" {
			return client.GetUserImagePartURL(ctx, a.cfg.UserID, args[0], args[1], part)
		}
		return client.GetUserISOPartURL(ctx, a.cfg.UserID, args[0], args[1], part)
	}, cobra.ExactArgs(2))
	cmd.Flags().IntVar(&part, "part", 1, "multipart part number")
	return cmd
}

func userAssetCompleteCommand(kind, singular string) *cobra.Command {
	var partsJSON string
	cmd := jsonCommand("complete-upload <key> <upload-id>", "Complete "+singular+" multipart upload", func(cmd *cobra.Command, args []string) (any, error) {
		parts, err := parseJSONArrayObjects(partsJSON)
		if err != nil {
			return nil, err
		}
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return nil, err
		}
		defer cancel()
		if kind == "images" {
			return client.CompleteUserImageUpload(ctx, a.cfg.UserID, args[0], args[1], parts)
		}
		return client.CompleteUserISOUpload(ctx, a.cfg.UserID, args[0], args[1], parts)
	}, cobra.ExactArgs(2))
	cmd.Flags().StringVar(&partsJSON, "parts", "[]", "JSON array of completed parts")
	return cmd
}

func userAssetDeleteCommand(kind, singular string) *cobra.Command {
	return &cobra.Command{Use: "delete <key>", Short: "Delete user " + singular, Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		opts, _ := commandOptions(cmd)
		if err := confirmRisky(cmd, opts, "Deleting user "+singular); err != nil {
			return err
		}
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		if kind == "images" {
			return client.DeleteUserImage(ctx, a.cfg.UserID, args[0])
		}
		return client.DeleteUserISO(ctx, a.cfg.UserID, args[0])
	}}
}

func newSSHKeysCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "ssh-keys", Short: "Manage SSH keys"}
	var jsonBody, jsonFile string
	create := &cobra.Command{Use: "create", Short: "Create SSH key", RunE: func(cmd *cobra.Command, _ []string) error {
		body, err := parseBodyFlags(jsonBody, jsonFile)
		if err != nil {
			return err
		}
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		out, err := client.CreateSSHKey(ctx, a.cfg.UserID, body)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), out)
	}}
	create.Flags().StringVar(&jsonBody, "body", "", "JSON request body")
	create.Flags().StringVar(&jsonFile, "body-file", "", "file containing JSON request body")
	cmd.AddCommand(
		jsonCommand("list", "List SSH keys", func(cmd *cobra.Command, _ []string) (any, error) {
			a, client, ctx, cancel, err := appClientFromCommand(cmd)
			if err != nil {
				return nil, err
			}
			defer cancel()
			return client.ListSSHKeys(ctx, a.cfg.UserID)
		}),
		create,
		&cobra.Command{Use: "delete <id>", Short: "Delete SSH key", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			opts, _ := commandOptions(cmd)
			if err := confirmRisky(cmd, opts, "Deleting SSH key"); err != nil {
				return err
			}
			keyID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			a, client, ctx, cancel, err := appClientFromCommand(cmd)
			if err != nil {
				return err
			}
			defer cancel()
			return client.DeleteSSHKey(ctx, a.cfg.UserID, keyID)
		}},
	)
	return cmd
}

func newVLansCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "vlans", Short: "Manage VLANs"}
	var serverRef string
	list := jsonCommand("list", "List VLANs", func(cmd *cobra.Command, _ []string) (any, error) {
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return nil, err
		}
		defer cancel()
		var serverID int
		if serverRef != "" {
			serverID, err = resolveServerID(ctx, client, serverRef)
			if err != nil {
				return nil, err
			}
		}
		return client.ListVLans(ctx, a.cfg.UserID, serverID)
	})
	list.Flags().StringVar(&serverRef, "server-id", "", "filter by server ID or name")
	cmd.AddCommand(
		list,
		jsonCommand("get <id>", "Get user VLAN", func(cmd *cobra.Command, args []string) (any, error) {
			vlanID, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, err
			}
			a, client, ctx, cancel, err := appClientFromCommand(cmd)
			if err != nil {
				return nil, err
			}
			defer cancel()
			return client.GetUserVLan(ctx, a.cfg.UserID, vlanID)
		}, cobra.ExactArgs(1)),
		jsonCommand("global-get <id>", "Get VLAN by global endpoint", func(cmd *cobra.Command, args []string) (any, error) {
			vlanID, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, err
			}
			_, client, ctx, cancel, err := appClientFromCommand(cmd)
			if err != nil {
				return nil, err
			}
			defer cancel()
			return client.GetVLan(ctx, vlanID)
		}, cobra.ExactArgs(1)),
		newVLanUpdateCommand(),
	)
	return cmd
}

func newVLanUpdateCommand() *cobra.Command {
	var jsonBody, jsonFile string
	cmd := &cobra.Command{Use: "update <id>", Short: "Update VLAN", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		opts, _ := commandOptions(cmd)
		if err := confirmRisky(cmd, opts, "Updating VLAN"); err != nil {
			return err
		}
		vlanID, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		body, err := parseBodyFlags(jsonBody, jsonFile)
		if err != nil {
			return err
		}
		a, client, ctx, cancel, err := appClientFromCommand(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		out, err := client.UpdateVLan(ctx, a.cfg.UserID, vlanID, body)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), out)
	}}
	cmd.Flags().StringVar(&jsonBody, "body", "", "JSON request body")
	cmd.Flags().StringVar(&jsonFile, "body-file", "", "file containing JSON request body")
	return cmd
}

func commandClientFromCommand(cmd *cobra.Command) (*netcup.Client, *netcup.RefreshTokenSource, context.Context, context.CancelFunc, error) {
	opts, _ := commandOptions(cmd)
	return commandClient(cmd, opts)
}

func appClientFromCommand(cmd *cobra.Command) (*app, *netcup.Client, context.Context, context.CancelFunc, error) {
	opts, _ := commandOptions(cmd)
	a, err := newApp(opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	client, source, err := a.apiClient()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := contextWithTimeout(cmd.Context(), opts.Timeout)
	return a, client, ctx, func() {
		cancel()
		a.persistRefreshToken(source)
	}, nil
}

func parseBodyFlags(raw, path string) (map[string]any, error) {
	if path != "" {
		return parseJSONObjectFile(path)
	}
	return parseJSONObject(raw)
}

func jsonCommand(use, short string, run func(*cobra.Command, []string) (any, error), args ...cobra.PositionalArgs) *cobra.Command {
	argValidator := cobra.NoArgs
	if len(args) > 0 {
		argValidator = args[0]
	}
	return &cobra.Command{Use: use, Short: short, Args: argValidator, RunE: func(cmd *cobra.Command, cmdArgs []string) error {
		out, err := run(cmd, cmdArgs)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), out)
	}}
}
