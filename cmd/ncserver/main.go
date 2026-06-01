package main

import (
	"fmt"
	"os"

	"github.com/tyr3al/ncctl/internal/cli"
)

func main() {
	if err := cli.NewServerRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
