package tunnel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type ConfigManager struct {
	configPath string
	logger     zerolog.Logger
}

func NewConfigManager(logger zerolog.Logger) *ConfigManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".uniroute")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "tunnels.json")

	return &ConfigManager{
		configPath: configPath,
		logger:     logger,
	}
}

func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

func (cm *ConfigManager) Load() (*TunnelConfigFile, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TunnelConfigFile{
				Version: "1.0",
				Tunnels: []TunnelConfig{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TunnelConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.logger.Debug().
		Str("path", cm.configPath).
		Int("tunnels", len(config.Tunnels)).
		Msg("Loaded tunnel configuration")

	return &config, nil
}

func (cm *ConfigManager) Save(config *TunnelConfigFile) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.Version == "" {
		config.Version = "1.0"
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.logger.Debug().
		Str("path", cm.configPath).
		Int("tunnels", len(config.Tunnels)).
		Msg("Saved tunnel configuration")

	return nil
}

func (cm *ConfigManager) GetEnabledTunnels() ([]TunnelConfig, error) {
	config, err := cm.Load()
	if err != nil {
		return nil, err
	}

	enabled := make([]TunnelConfig, 0)
	for _, tunnel := range config.Tunnels {
		if tunnel.Enabled {
			enabled = append(enabled, tunnel)
		}
	}

	return enabled, nil
}

func (cm *ConfigManager) GetTunnelByName(name string) (*TunnelConfig, error) {
	config, err := cm.Load()
	if err != nil {
		return nil, err
	}

	for _, tunnel := range config.Tunnels {
		if tunnel.Name == name {
			return &tunnel, nil
		}
	}

	return nil, fmt.Errorf("tunnel '%s' not found", name)
}

func (cm *ConfigManager) AddTunnel(tunnel TunnelConfig) error {
	config, err := cm.Load()
	if err != nil {
		return err
	}

	for i, t := range config.Tunnels {
		if t.Name == tunnel.Name {
			config.Tunnels[i] = tunnel
			return cm.Save(config)
		}
	}

	config.Tunnels = append(config.Tunnels, tunnel)
	return cm.Save(config)
}

// RemoveTunnel removes a tunnel from the configuration
func (cm *ConfigManager) RemoveTunnel(name string) error {
	config, err := cm.Load()
	if err != nil {
		return err
	}

	for i, tunnel := range config.Tunnels {
		if tunnel.Name == name {
			config.Tunnels = append(config.Tunnels[:i], config.Tunnels[i+1:]...)
			return cm.Save(config)
		}
	}

	return fmt.Errorf("tunnel '%s' not found", name)
}
