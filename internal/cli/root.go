package cli

import (
	"io"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultAPIBaseURL  = "https://servercontrolpanel.de/scp-core"
	defaultAuthBaseURL = "https://servercontrolpanel.de"
)

type options struct {
	ConfigPath  string
	APIBaseURL  string
	AuthBaseURL string
	Timeout     time.Duration
	JSON        bool
	Yes         bool
}

// NewRootCommand creates the netcupctl command tree.
func NewRootCommand() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:           "netcupctl",
		Short:         "Administer netcup SCP resources",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "config file path")
	cmd.PersistentFlags().StringVar(&opts.APIBaseURL, "api-base-url", defaultAPIBaseURL, "SCP API base URL")
	cmd.PersistentFlags().StringVar(&opts.AuthBaseURL, "auth-base-url", defaultAuthBaseURL, "SCP auth base URL")
	cmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", 30*time.Second, "request timeout")
	cmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "write JSON output")
	cmd.PersistentFlags().BoolVarP(&opts.Yes, "yes", "y", false, "confirm risky operations")

	attachOptions(cmd, opts)
	cmd.AddCommand(
		newVersionCommand(),
		newLoginCommand(),
		newLogoutCommand(),
		newWhoamiCommand(),
		newServersCommand(),
		newInterfacesCommand(),
		newFailoverCommand(),
		newTasksCommand(),
		newRDNSCommand(),
		newSnapshotsCommand(),
		newRescueCommand(),
		newDisksCommand(),
		newISOCommand(),
		newFirewallCommand(),
	)
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := io.WriteString(cmd.OutOrStdout(), "netcupctl dev\n")
			return err
		},
	}
}
