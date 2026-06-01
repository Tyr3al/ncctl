package netcup

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) GetMaintenance(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/maintenance", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetOpenAPI(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/openapi", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) OpenAPIMCP(ctx context.Context, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/openapi/mcp", nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetSupportedDiskDrivers(ctx context.Context, serverID int) ([]string, error) {
	var out []string
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/disks/supported-drivers", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDisk(ctx context.Context, serverID int, diskName string) (*ServerDisk, error) {
	var out ServerDisk
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/disks/"+url.PathEscape(diskName), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetGPUDriver(ctx context.Context, serverID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/gpu-driver", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetGuestAgent(ctx context.Context, serverID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/guest-agent", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetGuestAgentStatus(ctx context.Context, serverID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/guest-agent/status", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SetupImage(ctx context.Context, serverID int, body map[string]any) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/image", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListImageFlavours(ctx context.Context, serverID int) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/imageflavours", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetServerLogs(ctx context.Context, serverID, limit, offset int) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/logs", limitOffsetQuery(limit, offset), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetServerMetric(ctx context.Context, serverID int, metric string, hours int) (map[string]any, error) {
	query := url.Values{}
	if hours != 0 {
		query.Set("hours", strconv.Itoa(hours))
	}
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/metrics/"+metric, query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetSnapshot(ctx context.Context, serverID int, name string) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/servers/"+strconv.Itoa(serverID)+"/snapshots/"+url.PathEscape(name), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) OptimizeStorage(ctx context.Context, serverID int, disks []string, startAfter bool) (*TaskInfo, error) {
	query := url.Values{}
	for _, disk := range disks {
		query.Add("disks", disk)
	}
	if startAfter {
		query.Set("startAfterOptimization", "true")
	}
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/storageoptimization", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SetupUserImage(ctx context.Context, serverID int, body map[string]any) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/user-image", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CancelTask(ctx context.Context, uuid string) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPut, "/api/v1/tasks/"+url.PathEscape(uuid)+":cancel", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetUser(ctx context.Context, userID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateUser(ctx context.Context, userID int, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodPut, "/api/v1/users/"+strconv.Itoa(userID), nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateFirewallPolicy(ctx context.Context, userID int, body map[string]any) (*FirewallPolicy, error) {
	var out FirewallPolicy
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/users/"+strconv.Itoa(userID)+"/firewall-policies", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateFirewallPolicy(ctx context.Context, userID, policyID int, body map[string]any) (*FirewallPolicy, error) {
	var out FirewallPolicy
	if err := c.DoJSON(ctx, http.MethodPut, "/api/v1/users/"+strconv.Itoa(userID)+"/firewall-policies/"+strconv.Itoa(policyID), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SaveInterfaceFirewall(ctx context.Context, serverID int, mac string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodPut, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces/"+url.PathEscape(mac)+"/firewall", nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) RestoreCopiedInterfaceFirewall(ctx context.Context, serverID int, mac string) (*TaskInfo, error) {
	var out TaskInfo
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/servers/"+strconv.Itoa(serverID)+"/interfaces/"+url.PathEscape(mac)+"/firewall:restore-copied-policies", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListUserImages(ctx context.Context, userID int) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/images", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PrepareUserImageUpload(ctx context.Context, userID int, key string, multipart bool) (map[string]any, error) {
	return c.prepareUserAssetUpload(ctx, userID, "images", key, multipart)
}

func (c *Client) GetUserImage(ctx context.Context, userID int, key string) (map[string]any, error) {
	return c.getUserAsset(ctx, userID, "images", key)
}

func (c *Client) DeleteUserImage(ctx context.Context, userID int, key string) error {
	return c.deleteUserAsset(ctx, userID, "images", key)
}

func (c *Client) CompleteUserImageUpload(ctx context.Context, userID int, key, uploadID string, parts []map[string]any) (map[string]any, error) {
	return c.completeUserAssetUpload(ctx, userID, "images", key, uploadID, parts)
}

func (c *Client) GetUserImagePartURL(ctx context.Context, userID int, key, uploadID string, partNumber int) (map[string]any, error) {
	return c.getUserAssetPartURL(ctx, userID, "images", key, uploadID, partNumber)
}

func (c *Client) ListUserISOs(ctx context.Context, userID int) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/isos", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PrepareUserISOUpload(ctx context.Context, userID int, key string, multipart bool) (map[string]any, error) {
	return c.prepareUserAssetUpload(ctx, userID, "isos", key, multipart)
}

func (c *Client) GetUserISO(ctx context.Context, userID int, key string) (map[string]any, error) {
	return c.getUserAsset(ctx, userID, "isos", key)
}

func (c *Client) DeleteUserISO(ctx context.Context, userID int, key string) error {
	return c.deleteUserAsset(ctx, userID, "isos", key)
}

func (c *Client) CompleteUserISOUpload(ctx context.Context, userID int, key, uploadID string, parts []map[string]any) (map[string]any, error) {
	return c.completeUserAssetUpload(ctx, userID, "isos", key, uploadID, parts)
}

func (c *Client) GetUserISOPartURL(ctx context.Context, userID int, key, uploadID string, partNumber int) (map[string]any, error) {
	return c.getUserAssetPartURL(ctx, userID, "isos", key, uploadID, partNumber)
}

func (c *Client) ListSSHKeys(ctx context.Context, userID int) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/ssh-keys", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateSSHKey(ctx context.Context, userID int, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/users/"+strconv.Itoa(userID)+"/ssh-keys", nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DeleteSSHKey(ctx context.Context, userID, keyID int) error {
	return c.DoJSON(ctx, http.MethodDelete, "/api/v1/users/"+strconv.Itoa(userID)+"/ssh-keys/"+strconv.Itoa(keyID), nil, nil, nil)
}

func (c *Client) ListVLans(ctx context.Context, userID, serverID int) ([]map[string]any, error) {
	query := url.Values{}
	if serverID != 0 {
		query.Set("serverId", strconv.Itoa(serverID))
	}
	var out []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/vlans", query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetUserVLan(ctx context.Context, userID, vlanID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/vlans/"+strconv.Itoa(vlanID), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetVLan(ctx context.Context, vlanID int) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/vlans/"+strconv.Itoa(vlanID), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateVLan(ctx context.Context, userID, vlanID int, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodPut, "/api/v1/users/"+strconv.Itoa(userID)+"/vlans/"+strconv.Itoa(vlanID), nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetUserLogs(ctx context.Context, userID, limit, offset int) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/logs", limitOffsetQuery(limit, offset), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) prepareUserAssetUpload(ctx context.Context, userID int, kind, key string, multipart bool) (map[string]any, error) {
	query := url.Values{}
	if multipart {
		query.Set("multipart", "true")
	}
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/users/"+strconv.Itoa(userID)+"/"+kind+"/"+url.PathEscape(key), query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) getUserAsset(ctx context.Context, userID int, kind, key string) (map[string]any, error) {
	var out map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID)+"/"+kind+"/"+url.PathEscape(key), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) deleteUserAsset(ctx context.Context, userID int, kind, key string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/api/v1/users/"+strconv.Itoa(userID)+"/"+kind+"/"+url.PathEscape(key), nil, nil, nil)
}

func (c *Client) completeUserAssetUpload(ctx context.Context, userID int, kind, key, uploadID string, parts []map[string]any) (map[string]any, error) {
	var out map[string]any
	path := "/api/v1/users/" + strconv.Itoa(userID) + "/" + kind + "/" + url.PathEscape(key) + "/" + url.PathEscape(uploadID)
	if err := c.DoJSON(ctx, http.MethodPut, path, nil, parts, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) getUserAssetPartURL(ctx context.Context, userID int, kind, key, uploadID string, partNumber int) (map[string]any, error) {
	var out map[string]any
	path := "/api/v1/users/" + strconv.Itoa(userID) + "/" + kind + "/" + url.PathEscape(key) + "/" + url.PathEscape(uploadID) + "/parts/" + strconv.Itoa(partNumber)
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func limitOffsetQuery(limit, offset int) url.Values {
	query := url.Values{}
	if limit != 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset != 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	return query
}
