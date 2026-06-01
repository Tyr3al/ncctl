package netcup

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) PatchServer(ctx context.Context, serverID int, patch map[string]any, stateOption string) (*TaskInfo, error) {
	query := url.Values{}
	if stateOption != "" {
		query.Set("stateOption", stateOption)
	}
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPatch, "/api/v1/servers/"+strconv.Itoa(serverID), query, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SetRDNSIPv4(ctx context.Context, ip, rdns string) error {
	return c.DoJSON(ctx, http.MethodPost, "/api/v1/rdns/ipv4", nil, map[string]string{"ip": ip, "rdns": rdns}, nil)
}

func (c *Client) SetRDNSIPv6(ctx context.Context, ip, rdns string) error {
	return c.DoJSON(ctx, http.MethodPost, "/api/v1/rdns/ipv6", nil, map[string]string{"ip": ip, "rdns": rdns}, nil)
}

func (c *Client) DeleteRDNSIPv4(ctx context.Context, ip string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/api/v1/rdns/ipv4/"+url.PathEscape(ip), nil, nil, nil)
}

func (c *Client) DeleteRDNSIPv6(ctx context.Context, ip string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/api/v1/rdns/ipv6/"+url.PathEscape(ip), nil, nil, nil)
}

func (c *Client) CreateInterfaceVLAN(ctx context.Context, serverID, vlanID int, driver string) (*Interface, error) {
	var out Interface
	body := map[string]any{"vlanId": vlanID, "networkDriver": driver}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateInterface(ctx context.Context, serverID int, mac, driver string) (*Interface, error) {
	var out Interface
	body := map[string]string{"driver": driver}
	if err := c.DoJSON(ctx, http.MethodPut, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces/"+url.PathEscape(mac), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteInterface(ctx context.Context, serverID int, mac string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces/"+url.PathEscape(mac), nil, nil, nil)
}

func (c *Client) ListSnapshots(ctx context.Context, serverID int) ([]SnapshotMinimal, error) {
	var out []SnapshotMinimal
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/snapshots", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateSnapshot(ctx context.Context, serverID int, body map[string]any) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/snapshots", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSnapshot(ctx context.Context, serverID int, name string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/api/v1/servers/"+strconv.Itoa(serverID)+"/snapshots/"+url.PathEscape(name), nil, nil, nil)
}

func (c *Client) RevertSnapshot(ctx context.Context, serverID int, name string) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/snapshots/"+url.PathEscape(name)+"/revert", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ExportSnapshot(ctx context.Context, serverID int, name string) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/snapshots/"+url.PathEscape(name)+"/export", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SnapshotDryRun(ctx context.Context, serverID int, body map[string]any) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/snapshots:dryrun", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetRescueSystem(ctx context.Context, serverID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/rescuesystem", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ActivateRescueSystem(ctx context.Context, serverID int) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/rescuesystem", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeactivateRescueSystem(ctx context.Context, serverID int) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodDelete, "/api/v1/servers/"+strconv.Itoa(serverID)+"/rescuesystem", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListDisks(ctx context.Context, serverID int) ([]ServerDisk, error) {
	var out []ServerDisk
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/disks", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SetDiskDriver(ctx context.Context, serverID int, body map[string]any) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPatch, "/api/v1/servers/"+strconv.Itoa(serverID)+"/disks", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) FormatDisk(ctx context.Context, serverID int, diskName string) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/disks/"+url.PathEscape(diskName)+":format", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetAttachedISO(ctx context.Context, serverID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/iso", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListISOImages(ctx context.Context, serverID int) ([]ISOImage, error) {
	var out []ISOImage
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/isoimages", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AttachISO(ctx context.Context, serverID int, body map[string]any) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/iso", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DetachISO(ctx context.Context, serverID int) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodDelete, "/api/v1/servers/"+strconv.Itoa(serverID)+"/iso", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListFirewallPolicies(ctx context.Context, userID int) ([]FirewallPolicy, error) {
	var out []FirewallPolicy
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/firewall-policies", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetFirewallPolicy(ctx context.Context, userID, policyID int) (*FirewallPolicy, error) {
	var out FirewallPolicy
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/firewall-policies/"+strconv.Itoa(policyID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteFirewallPolicy(ctx context.Context, userID, policyID int) error {
	return c.DoJSON(ctx, http.MethodDelete, "/api/v1/users/"+strconv.Itoa(userID)+"/firewall-policies/"+strconv.Itoa(policyID), nil, nil, nil)
}

func (c *Client) GetInterfaceFirewall(ctx context.Context, serverID int, mac string) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces/"+url.PathEscape(mac)+"/firewall", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ReapplyInterfaceFirewall(ctx context.Context, serverID int, mac string) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces/"+url.PathEscape(mac)+"/firewall:reapply", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
