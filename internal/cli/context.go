package cli

import "github.com/spf13/cobra"

type optionsKey struct{}

func attachOptions(cmd *cobra.Command, opts *options) {
	cmd.SetContext(contextWithOptions(cmd.Context(), opts))
}

func commandOptions(cmd *cobra.Command) (*options, bool) {
	opts, ok := optionsFromContext(cmd.Context())
	return opts, ok
}
