package config

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultVersion = "0.1"

	defaultConfigName = "live-feed.conf"
	defaultConfigDir  = "live-feed"

	defaultLogLevel = "info"

	defaultReceiverBindAddress = ""
	defaultReceiverPort        = 8040
	defaultReceiverProtocol    = "tcp"

	defaultSubscriptionAPIAddress  = ""
	defaultSubscriptionAPIPort     = 8700
	defaultSubscriptionAPIBasePath = "/live/feed/v" + defaultVersion
)

var defaultSubscriptionAPISchemes = []string{"HTTP", "HTTPS"}

// Config the configuration of the live-feed api
type Config struct {
	Log          *LogConfig
	Receiver     *ReceiverConfig
	Sender       *SenderConfig
	Subscription *SubscriptionConfig
}

// LogConfig configuration for logging
type LogConfig struct {
	LogLevel string
}

// ReceiverConfig configuration for the receiver
type ReceiverConfig struct {
	Protocol    string
	BindAddress string
	Port        uint16
}

// SenderConfig configuration for the sender
type SenderConfig struct {
	MaxEventRetries          uint16
	MaxReconnectRetries      uint16
	ReconnectHoldOffDuration time.Duration
}

// SubscriptionConfig configuration for the subcription api
type SubscriptionConfig struct {
	Schemes     []string
	BindAddress string // (Duplicated because extended interfaces are not supported by Viper)
	Port        uint16
	BasePath    string
}

// LoadConfiguration from default paths for live-feed.conf
// look for configuration in:
// - current path
// - /etc/factom-live-feed
// - $HOME/.factom
// - $HOME/.factom/live-feed
// - current path
func LoadConfiguration() (*Config, error) {
	configPaths := []string{
		"",
		substituteHomeDir("$HOME/.factom/"),
		substituteHomeDir(fmt.Sprintf("$HOME/.factom/%s", defaultConfigDir)),
		fmt.Sprintf("/etc/%s", defaultConfigDir),
	}

	for _, path := range configPaths {
		configFile := filepath.Join(path, defaultConfigName)
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue
		}

		return loadConfigurationFrom(configFile)
	}

	log.Warn(`failed to find configuration in: ["%s"]`, strings.Join(configPaths, `", "`))
	return defaultConfig(), nil
}

// LoadConfigurationFrom specific a file
func LoadConfigurationFrom(filename string) (*Config, error) {
	filename = substituteHomeDir(filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to find configuration: '%s'", filename)
	}

	return loadConfigurationFrom(filename)
}

func loadConfigurationFrom(configFile string) (*Config, error) {
	vp := viper.New()
	vp.SetConfigType("toml")
	vp.SetConfigFile(configFile)

	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.SetEnvPrefix("factom_live_feed")
	vp.AutomaticEnv()

	// set default configuration values
	vp.SetDefault("log", buildLogDefaults())
	vp.SetDefault("receiver", buildReceiverDefaults())
	vp.SetDefault("subscription", buildSubscriptionDefaults())

	// read/build configuration
	if err := vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read configuration: %v", err)
	}

	config := &Config{}
	if err := vp.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("could not read configuration file: %v", err)
	}

	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		Log: &LogConfig{LogLevel: defaultLogLevel},
		Receiver: &ReceiverConfig{
			Protocol:    defaultReceiverProtocol,
			BindAddress: defaultReceiverBindAddress,
			Port:        defaultReceiverPort,
		},
		Sender: &SenderConfig{
			MaxEventRetries:          3,
			MaxReconnectRetries:      4,
			ReconnectHoldOffDuration: 15 * time.Minute,
		},
		Subscription: &SubscriptionConfig{
			Schemes:     defaultSubscriptionAPISchemes,
			BindAddress: defaultSubscriptionAPIAddress,
			Port:        defaultSubscriptionAPIPort,
			BasePath:    defaultSubscriptionAPIBasePath,
		},
	}
}

func buildReceiverDefaults() map[string]interface{} {
	return map[string]interface{}{
		"Protocol":    defaultReceiverProtocol,
		"BindAddress": defaultReceiverBindAddress,
		"Port":        defaultReceiverPort,
	}
}

func buildSubscriptionDefaults() map[string]interface{} {
	return map[string]interface{}{
		"BindAddress": defaultSubscriptionAPIAddress,
		"Port":        defaultSubscriptionAPIPort,
		"Schemes":     defaultSubscriptionAPISchemes,
		"BasePath":    defaultSubscriptionAPIBasePath,
	}
}

func buildLogDefaults() map[string]interface{} {
	return map[string]interface{}{
		"loglevel": defaultLogLevel,
	}
}

func substituteHomeDir(path string) string {
	if homeDir, err := os.UserHomeDir(); err == nil {
		return strings.ReplaceAll(path, "$HOME", homeDir)
	}
	return path
}
