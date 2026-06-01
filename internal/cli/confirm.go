package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func confirmRisky(cmd *cobra.Command, opts *options, action string) error {
	if opts.Yes {
		return nil
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "%s. Type yes to continue: ", action)
	reader := bufio.NewReader(cmd.InOrStdin())
	answer, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(answer)) != "yes" {
		return fmt.Errorf("aborted")
	}
	return nil
}
