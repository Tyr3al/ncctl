package netcup

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *Client) Ping(ctx context.Context) error {
	return c.DoJSON(ctx, http.MethodGet, "/api/ping", nil, nil, nil)
}

func (c *Client) UserInfo(ctx context.Context) (*UserInfo, error) {
	var out UserInfo
	if err := c.DoJSON(ctx, http.MethodGet, "/realms/scp/protocol/openid-connect/userinfo", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ListServersOptions struct {
	FirewallPolicyID int
	IP               string
	Limit            int
	Name             string
	Offset           int
	Query            string
	Sort             string
}

func (c *Client) ListServers(ctx context.Context, opts ListServersOptions) ([]ServerListMinimal, error) {
	var out []ServerListMinimal
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers", serverListQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetServer(ctx context.Context, serverID int, loadLiveInfo bool) (*Server, error) {
	query := url.Values{}
	if loadLiveInfo {
		query.Set("loadServerLiveInfo", "true")
	}
	var out Server
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID), query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListInterfaces(ctx context.Context, serverID int, loadRDNS bool) ([]Interface, error) {
	query := url.Values{}
	if loadRDNS {
		query.Set("loadRdns", "true")
	}
	var out []Interface
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces", query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetInterface(ctx context.Context, serverID int, mac string, loadRDNS bool) (*Interface, error) {
	query := url.Values{}
	if loadRDNS {
		query.Set("loadRdns", "true")
	}
	var out Interface
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces/"+url.PathEscape(mac), query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ListFailoverOptions struct {
	IP       string
	ServerID int
}

func (c *Client) ListFailoverIPv4(ctx context.Context, userID int, opts ListFailoverOptions) ([]FailoverIPv4, error) {
	var out []FailoverIPv4
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/failoverips/v4", failoverQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListFailoverIPv6(ctx context.Context, userID int, opts ListFailoverOptions) ([]FailoverIPv6, error) {
	var out []FailoverIPv6
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/failoverips/v6", failoverQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type RouteFailoverIP struct {
	ServerID int `json:"serverId"`
}

func (c *Client) RouteFailoverIPv4(ctx context.Context, userID, failoverID, serverID int) (*TaskInfo, error) {
	var out TaskInfo
	path := "/api/v1/users/" + strconv.Itoa(userID) + "/failoverips/v4/" + strconv.Itoa(failoverID)
	if err := c.DoJSON(ctx, http.MethodPatch, path, nil, RouteFailoverIP{ServerID: serverID}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RouteFailoverIPv6(ctx context.Context, userID, failoverID, serverID int) (*TaskInfo, error) {
	var out TaskInfo
	path := "/api/v1/users/" + strconv.Itoa(userID) + "/failoverips/v6/" + strconv.Itoa(failoverID)
	if err := c.DoJSON(ctx, http.MethodPatch, path, nil, RouteFailoverIP{ServerID: serverID}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetRDNSIPv4(ctx context.Context, ip string) (string, error) {
	var out string
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/rdns/ipv4/"+url.PathEscape(ip), nil, nil, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) GetRDNSIPv6(ctx context.Context, ip string) (map[string]string, error) {
	var out map[string]string
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/rdns/ipv6/"+url.PathEscape(ip), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type ListTasksOptions struct {
	Limit    int
	Offset   int
	Query    string
	ServerID int
	State    string
}

func (c *Client) ListTasks(ctx context.Context, opts ListTasksOptions) ([]TaskInfoMinimal, error) {
	var out []TaskInfoMinimal
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/tasks", taskListQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetTask(ctx context.Context, uuid string) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/tasks/"+url.PathEscape(uuid), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) WaitTask(ctx context.Context, uuid string, interval time.Duration) (*TaskInfo, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		task, err := c.GetTask(ctx, uuid)
		if err != nil {
			return nil, err
		}
		if IsTerminalTaskState(task.State) {
			if IsFailedTaskState(task.State) {
				return task, fmt.Errorf("task %s finished with state %s", uuid, task.State)
			}
			return task, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func IsTerminalTaskState(state string) bool {
	switch strings.ToUpper(state) {
	case "FINISHED", "SUCCESS", "SUCCESSFUL", "DONE", "COMPLETED", "FAILED", "ERROR", "CANCELED", "CANCELLED":
		return true
	default:
		return false
	}
}

func IsFailedTaskState(state string) bool {
	switch strings.ToUpper(state) {
	case "FAILED", "ERROR", "CANCELED", "CANCELLED":
		return true
	default:
		return false
	}
}

func serverListQuery(opts ListServersOptions) url.Values {
	query := url.Values{}
	if opts.FirewallPolicyID != 0 {
		query.Set("firewallPolicyId", strconv.Itoa(opts.FirewallPolicyID))
	}
	if opts.IP != "" {
		query.Set("ip", opts.IP)
	}
	if opts.Limit != 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Name != "" {
		query.Set("name", opts.Name)
	}
	if opts.Offset != 0 {
		query.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Query != "" {
		query.Set("q", opts.Query)
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	return query
}

func failoverQuery(opts ListFailoverOptions) url.Values {
	query := url.Values{}
	if opts.IP != "" {
		query.Set("ip", opts.IP)
	}
	if opts.ServerID != 0 {
		query.Set("serverId", strconv.Itoa(opts.ServerID))
	}
	return query
}

func taskListQuery(opts ListTasksOptions) url.Values {
	query := url.Values{}
	if opts.Limit != 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset != 0 {
		query.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Query != "" {
		query.Set("q", opts.Query)
	}
	if opts.ServerID != 0 {
		query.Set("serverId", strconv.Itoa(opts.ServerID))
	}
	if opts.State != "" {
		query.Set("state", opts.State)
	}
	return query
}
