package cli

import (
	"fmt"
	"io"
	"text/tabwriter"
)

func writeTable(w io.Writer, header []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, col := range header {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, col)
	}
	fmt.Fprintln(tw)
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, col)
		}
		fmt.Fprintln(tw)
	}
	return tw.Flush()
}

func stringPtrValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
