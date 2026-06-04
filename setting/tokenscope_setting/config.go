package tokenscope_setting

import (
	"github.com/QuantumNous/new-api/setting/config"
)

// TokenScopeSetting controls L2 deep diagnostics sampling behavior.
type TokenScopeSetting struct {
	Enabled        bool    `json:"enabled"`          // L2 debug capture enabled
	SampleRate     float64 `json:"sample_rate"`      // 0.0-1.0, fraction of requests to capture
	MaxPayloadSize int     `json:"max_payload_size"` // max bytes to store per request body (default 100KB)
	RetentionDays  int     `json:"retention_days"`   // auto-cleanup days (default 7)
	CaptureModels  string  `json:"capture_models"`   // comma-separated model list (empty = all)
}

var tokenscopeSetting = TokenScopeSetting{
	Enabled:        false,
	SampleRate:     0.01, // 1% default
	MaxPayloadSize: 100 * 1024,
	RetentionDays:  7,
	CaptureModels:  "",
}

func init() {
	config.GlobalConfig.Register("tokenscope_setting", &tokenscopeSetting)
}

// GetTokenScopeSetting returns the global TokenScope settings.
func GetTokenScopeSetting() *TokenScopeSetting {
	return &tokenscopeSetting
}
