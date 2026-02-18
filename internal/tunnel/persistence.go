package tunnel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

type TunnelPersistence struct {
	filePath string
	logger   zerolog.Logger
}

// For backward compatibility - single tunnel.
type TunnelState struct {
	TunnelID    string    `json:"tunnel_id"`
	Subdomain   string    `json:"subdomain"`
	PublicURL   string    `json:"public_url"`
	LocalURL    string    `json:"local_url"`
	ServerURL   string    `json:"server_url"`
	Protocol    string    `json:"protocol,omitempty"` // http, tcp, tls
	Host        string    `json:"host,omitempty"`     // Optional host/subdomain
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
}

type MultiTunnelState struct {
	Tunnels map[string]*TunnelState `json:"tunnels"` // Key is tunnel name or local URL
}

func NewTunnelPersistence(logger zerolog.Logger) *TunnelPersistence {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	configDir := filepath.Join(homeDir, ".uniroute")
	os.MkdirAll(configDir, 0755)
	
	filePath := filepath.Join(configDir, "tunnel-state.json")
	
	return &TunnelPersistence{
		filePath: filePath,
		logger:   logger,
	}
}

func (tp *TunnelPersistence) Save(state *TunnelState) error {
	state.LastUsed = time.Now()
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tunnel state: %w", err)
	}
	
	if err := os.WriteFile(tp.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write tunnel state: %w", err)
	}
	
	tp.logger.Debug().
		Str("subdomain", state.Subdomain).
		Str("file", tp.filePath).
		Msg("Saved tunnel state")
	
	return nil
}

func (tp *TunnelPersistence) Load() (*TunnelState, error) {
	data, err := os.ReadFile(tp.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read tunnel state: %w", err)
	}
	
	var state TunnelState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tunnel state: %w", err)
	}
	
	tp.logger.Debug().
		Str("subdomain", state.Subdomain).
		Msg("Loaded tunnel state")
	
	return &state, nil
}

// Clear clears saved tunnel state
func (tp *TunnelPersistence) Clear() error {
	if err := os.Remove(tp.filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove tunnel state: %w", err)
	}
	
	tp.logger.Debug().Msg("Cleared tunnel state")
	return nil
}

func (tp *TunnelPersistence) Exists() bool {
	_, err := os.Stat(tp.filePath)
	return err == nil
}

