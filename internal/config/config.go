package config

type Config struct {
	APIBaseURL  string `json:"api_base_url"`
	AuthBaseURL string `json:"auth_base_url"`
	UserID      int    `json:"user_id,omitempty"`
	Refresh     string `json:"refresh_token,omitempty"`
}
