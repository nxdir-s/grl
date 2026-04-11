package valobj

type Config struct {
	ActiveEnvID     string `json:"active_env_id"`
	DefaultMethod   string `json:"default_method"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	FollowRedirects bool   `json:"follow_redirects"`
	HistoryLimit    int    `json:"history_limit"`
}
