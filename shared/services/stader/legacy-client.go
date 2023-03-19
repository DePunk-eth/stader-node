package stader

// ROCKETPOOL-OWNED

import (
	"fmt"
	"io/ioutil"

	"github.com/alessio/shellescape"
	"github.com/mitchellh/go-homedir"
	"github.com/stader-labs/stader-node/shared/services/config"
)

// Config
const (
	LegacyGlobalConfigFile    = "config.yml"
	LegacyUserConfigFile      = "settings.yml"
	LegacyComposeFile         = "docker-compose.yml"
	LegacyMetricsComposeFile  = "docker-compose-metrics.yml"
	LegacyFallbackComposeFile = "docker-compose-fallback.yml"
)

// Load the global config
func (c *Client) LoadGlobalConfig_Legacy(globalConfigPath string) (config.LegacyStaderConfig, error) {
	return c.loadConfig_Legacy(globalConfigPath)
}

// Load/save the user config
func (c *Client) LoadUserConfig_Legacy(userConfigPath string) (config.LegacyStaderConfig, error) {
	return c.loadConfig_Legacy(userConfigPath)
}

// Load the merged global & user config
func (c *Client) LoadMergedConfig_Legacy(globalConfigPath string, userConfigPath string) (config.LegacyStaderConfig, error) {
	globalConfig, err := c.LoadGlobalConfig_Legacy(globalConfigPath)
	if err != nil {
		return config.LegacyStaderConfig{}, err
	}
	userConfig, err := c.LoadUserConfig_Legacy(userConfigPath)
	if err != nil {
		return config.LegacyStaderConfig{}, err
	}
	return config.Merge(&globalConfig, &userConfig)
}

// Load a config file
func (c *Client) loadConfig_Legacy(path string) (config.LegacyStaderConfig, error) {
	expandedPath, err := homedir.Expand(path)
	if err != nil {
		return config.LegacyStaderConfig{}, err
	}
	configBytes, err := ioutil.ReadFile(expandedPath)
	if err != nil {
		return config.LegacyStaderConfig{}, fmt.Errorf("Could not read Stader config at %s: %w", shellescape.Quote(path), err)
	}
	return config.Parse(configBytes)
}
