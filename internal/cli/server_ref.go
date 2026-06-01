package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tyr3al/ncctl/pkg/netcup"
)

func resolveServerID(ctx context.Context, client *netcup.Client, ref string) (int, error) {
	if ref == "" {
		return 0, fmt.Errorf("server is required")
	}
	if id, err := strconv.Atoi(ref); err == nil {
		return id, nil
	}
	servers, err := client.ListServers(ctx, netcup.ListServersOptions{Limit: 100})
	if err != nil {
		return 0, err
	}
	var matches []netcup.ServerListMinimal
	for _, server := range servers {
		if server.Name == ref || (server.Nickname != nil && *server.Nickname == ref) {
			matches = append(matches, server)
		}
	}
	if len(matches) == 0 {
		return 0, fmt.Errorf("server %q not found", ref)
	}
	if len(matches) > 1 {
		return 0, fmt.Errorf("server name %q is ambiguous", ref)
	}
	return matches[0].ID, nil
}
