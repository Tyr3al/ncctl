package netcup

import "time"

type ResponseError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type UserInfo struct {
	ID                string `json:"id,omitempty"`
	Username          string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	Name              string `json:"name,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	CustomerNumber    string `json:"customerNumber,omitempty"`
	Subject           string `json:"sub,omitempty"`
	RawPreferredID    string `json:"user_id,omitempty"`
	RawSCPID          string `json:"scp_user_id,omitempty"`
	RawCustomerNumber string `json:"customer_number,omitempty"`
}

type UserMinimal struct {
	ID        int     `json:"id,omitempty"`
	Username  string  `json:"username,omitempty"`
	Firstname string  `json:"firstname,omitempty"`
	Lastname  string  `json:"lastname,omitempty"`
	Email     string  `json:"email,omitempty"`
	Company   *string `json:"company,omitempty"`
}

type Site struct {
	ID   int    `json:"id,omitempty"`
	City string `json:"city,omitempty"`
}

type ServerMinimal struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ServerListMinimal struct {
	ID       int                    `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Hostname *string                `json:"hostname,omitempty"`
	Nickname *string                `json:"nickname,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
	Template *ServerTemplateMinimal `json:"template,omitempty"`
}

type ServerTemplateMinimal struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Server struct {
	ID                     int                    `json:"id,omitempty"`
	Name                   string                 `json:"name,omitempty"`
	Hostname               *string                `json:"hostname,omitempty"`
	Nickname               *string                `json:"nickname,omitempty"`
	Disabled               bool                   `json:"disabled,omitempty"`
	Template               *ServerTemplateMinimal `json:"template,omitempty"`
	ServerLiveInfo         *ServerInfo            `json:"serverLiveInfo,omitempty"`
	IPv4Addresses          []IPv4AddressMinimal   `json:"ipv4Addresses,omitempty"`
	IPv6Addresses          []IPv6AddressMinimal   `json:"ipv6Addresses,omitempty"`
	Site                   Site                   `json:"site,omitempty"`
	SnapshotCount          int                    `json:"snapshotCount,omitempty"`
	MaxCPUCount            int                    `json:"maxCpuCount,omitempty"`
	DisksAvailableSpaceMiB int64                  `json:"disksAvailableSpaceInMiB,omitempty"`
	RescueSystemActive     bool                   `json:"rescueSystemActive,omitempty"`
	SnapshotAllowed        bool                   `json:"snapshotAllowed,omitempty"`
	Architecture           string                 `json:"architecture,omitempty"`
	GPUDriverAvailable     bool                   `json:"gpuDriverAvailable,omitempty"`
}

type ServerInfo struct {
	State                    string            `json:"state,omitempty"`
	Autostart                bool              `json:"autostart,omitempty"`
	UEFI                     bool              `json:"uefi,omitempty"`
	Interfaces               []ServerInterface `json:"interfaces,omitempty"`
	Disks                    []ServerDisk      `json:"disks,omitempty"`
	UptimeSeconds            int               `json:"uptimeInSeconds,omitempty"`
	CurrentServerMemoryInMiB int64             `json:"currentServerMemoryInMiB,omitempty"`
	MaxServerMemoryInMiB     int64             `json:"maxServerMemoryInMiB,omitempty"`
	CPUCount                 int               `json:"cpuCount,omitempty"`
	CPUMaxCount              int               `json:"cpuMaxCount,omitempty"`
	MachineType              string            `json:"machineType,omitempty"`
	KeyboardLayout           string            `json:"keyboardLayout,omitempty"`
}

type IPv4AddressMinimal struct {
	ID        int     `json:"id,omitempty"`
	IP        string  `json:"ip,omitempty"`
	Netmask   string  `json:"netmask,omitempty"`
	Gateway   *string `json:"gateway,omitempty"`
	Broadcast *string `json:"broadcast,omitempty"`
}

type IPv6AddressMinimal struct {
	ID                  int     `json:"id,omitempty"`
	NetworkPrefix       string  `json:"networkPrefix,omitempty"`
	NetworkPrefixLength int     `json:"networkPrefixLength,omitempty"`
	Gateway             *string `json:"gateway,omitempty"`
}

type ServerInterface struct {
	MAC                    string   `json:"mac,omitempty"`
	Driver                 string   `json:"driver,omitempty"`
	MTU                    int      `json:"mtu,omitempty"`
	SpeedInMBits           int      `json:"speedInMBits,omitempty"`
	RXMonthlyInMiB         int      `json:"rxMonthlyInMiB,omitempty"`
	TXMonthlyInMiB         int      `json:"txMonthlyInMiB,omitempty"`
	IPv4Addresses          []string `json:"ipv4Addresses,omitempty"`
	IPv6LinkLocalAddresses []string `json:"ipv6LinkLocalAddresses,omitempty"`
	IPv6NetworkPrefixes    []string `json:"ipv6NetworkPrefixes,omitempty"`
	TrafficThrottled       bool     `json:"trafficThrottled,omitempty"`
	VLANInterface          bool     `json:"vlanInterface,omitempty"`
	VLANID                 int      `json:"vlanId,omitempty"`
}

type Interface struct {
	MAC           string       `json:"mac,omitempty"`
	Driver        string       `json:"driver,omitempty"`
	SpeedInMBits  int          `json:"speedInMBits,omitempty"`
	IPv4Addresses []ServerIPv4 `json:"ipv4Addresses,omitempty"`
	IPv6Addresses []ServerIPv6 `json:"ipv6Addresses,omitempty"`
}

type ServerIPv4 struct {
	ID            int     `json:"id,omitempty"`
	InterfaceMAC  string  `json:"interfaceMac,omitempty"`
	Type          string  `json:"type,omitempty"`
	IP            string  `json:"ip,omitempty"`
	CIDR          string  `json:"cidr,omitempty"`
	Gateway       *string `json:"gateway,omitempty"`
	RDNS          *string `json:"rdns,omitempty"`
	DestinationIP *string `json:"destinationIp,omitempty"`
	Editable      bool    `json:"editable,omitempty"`
}

type ServerIPv6 struct {
	ID            int               `json:"id,omitempty"`
	InterfaceMAC  string            `json:"interfaceMac,omitempty"`
	Type          string            `json:"type,omitempty"`
	NetworkPrefix string            `json:"networkPrefix,omitempty"`
	CIDR          string            `json:"cidr,omitempty"`
	Gateway       *string           `json:"gateway,omitempty"`
	LinkLocal     bool              `json:"linkLocal,omitempty"`
	RDNS          map[string]string `json:"rdns,omitempty"`
	DestinationIP *string           `json:"destinationIp,omitempty"`
	Editable      bool              `json:"editable,omitempty"`
}

type ServerDisk struct {
	Dev             string `json:"dev,omitempty"`
	Driver          string `json:"driver,omitempty"`
	CapacityInMiB   int64  `json:"capacityInMiB,omitempty"`
	AllocationInMiB int64  `json:"allocationInMiB,omitempty"`
}

type FailoverIPv4 struct {
	ID         int           `json:"id,omitempty"`
	IP         string        `json:"ip,omitempty"`
	CIDRSuffix int           `json:"cidrSuffix,omitempty"`
	User       UserMinimal   `json:"user,omitempty"`
	Editable   bool          `json:"editable,omitempty"`
	Site       Site          `json:"site,omitempty"`
	Server     ServerMinimal `json:"server,omitempty"`
}

type FailoverIPv6 struct {
	ID                  int           `json:"id,omitempty"`
	NetworkPrefix       string        `json:"networkPrefix,omitempty"`
	NetworkPrefixLength int           `json:"networkPrefixLength,omitempty"`
	User                UserMinimal   `json:"user,omitempty"`
	Editable            bool          `json:"editable,omitempty"`
	Site                Site          `json:"site,omitempty"`
	Server              ServerMinimal `json:"server,omitempty"`
}

type TaskInfoMinimal struct {
	UUID          string        `json:"uuid,omitempty"`
	Name          string        `json:"name,omitempty"`
	State         string        `json:"state,omitempty"`
	StartedAt     *time.Time    `json:"startedAt,omitempty"`
	FinishedAt    *time.Time    `json:"finishedAt,omitempty"`
	ExecutingUser UserMinimal   `json:"executingUser,omitempty"`
	TaskProgress  *TaskProgress `json:"taskProgress,omitempty"`
	Message       *string       `json:"message,omitempty"`
	OnRollback    bool          `json:"onRollback,omitempty"`
}

type TaskInfo struct {
	TaskInfoMinimal
	Steps         []TaskInfoStep `json:"steps,omitempty"`
	Result        map[string]any `json:"result,omitempty"`
	ResponseError *ResponseError `json:"responseError,omitempty"`
}

type TaskProgress struct {
	CurrentStep int `json:"currentStep,omitempty"`
	MaxSteps    int `json:"maxSteps,omitempty"`
	Percent     int `json:"percent,omitempty"`
}

type TaskInfoStep struct {
	Name       string     `json:"name,omitempty"`
	State      string     `json:"state,omitempty"`
	Message    *string    `json:"message,omitempty"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}

type SnapshotMinimal struct {
	UUID              string     `json:"uuid,omitempty"`
	Name              string     `json:"name,omitempty"`
	Description       *string    `json:"description,omitempty"`
	Disks             []string   `json:"disks,omitempty"`
	CreationTime      *time.Time `json:"creationTime,omitempty"`
	State             string     `json:"state,omitempty"`
	Online            bool       `json:"online,omitempty"`
	Exported          bool       `json:"exported,omitempty"`
	ExportedSizeInKiB *int64     `json:"exportedSizeInKiB,omitempty"`
}

type ISOImage struct {
	ID           int     `json:"id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Filename     string  `json:"filename,omitempty"`
	Description  *string `json:"description,omitempty"`
	Architecture string  `json:"architecture,omitempty"`
}

type FirewallPolicy struct {
	ID          int            `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Rules       []FirewallRule `json:"rules,omitempty"`
}

type FirewallRule struct {
	Description      *string  `json:"description,omitempty"`
	Direction        string   `json:"direction,omitempty"`
	Protocol         string   `json:"protocol,omitempty"`
	Action           string   `json:"action,omitempty"`
	Sources          []string `json:"sources,omitempty"`
	SourcePorts      string   `json:"sourcePorts,omitempty"`
	Destinations     []string `json:"destinations,omitempty"`
	DestinationPorts string   `json:"destinationPorts,omitempty"`
}
