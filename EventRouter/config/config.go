package config

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultVersion = "0.1"

	defaultConfigName = "factom-live-feed.conf"
	defaultConfigDir  = "live-feed"

	defaultLogLevel = "info"

	defaultReceiverBindAddress = ""
	defaultReceiverPort        = 8040
	defaultReceiverProtocol    = "tcp"

	defaultRouterMaxRetries   = 3
	defaultRouterRetryTimeout = 30

	defaultSubscriptionAPIAddress  = ""
	defaultSubscriptionAPIPort     = 8700
	defaultSubscriptionAPIBasePath = "/live/feed/v" + defaultVersion

	defaultDatabase                 = "inmemory"
	defaultDatabaseConnectionString = ""
)

var defaultSubscriptionAPISchemes = "HTTP"

// Config the configuration of the live-feed api
type Config struct {
	Log          *LogConfig
	Receiver     *ReceiverConfig
	Router       *RouterConfig
	Subscription *SubscriptionConfig
	Database     *DatabaseConfig
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

// RouterConfig configuration for the event router
type RouterConfig struct {
	MaxRetries   uint16
	RetryTimeout uint
}

// SubscriptionConfig configuration for the subscription api
type SubscriptionConfig struct {
	Scheme          string
	BindAddress     string
	Port            uint16
	BasePath        string
	CertificateFile string
	PrivateKeyFile  string
}

// DatabaseConfig configuration for the database to store subscriptions
type DatabaseConfig struct {
	Database         string
	ConnectionString string
}

// LoadConfiguration from default paths for factom-live-feed.conf
// look for configuration in:
// - current path
// - /etc/factom-live-feed
// - $HOME/.factom
// - $HOME/.factom/factom-live-feed
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
	vp.SetDefault("router", buildRouterDefaults())
	vp.SetDefault("subscription", buildSubscriptionDefaults())
	vp.SetDefault("database", buildDatabaseDefaults())

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
		Router: &RouterConfig{
			MaxRetries:   defaultRouterMaxRetries,
			RetryTimeout: defaultRouterRetryTimeout,
		},
		Subscription: &SubscriptionConfig{
			Scheme:      defaultSubscriptionAPISchemes,
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

func buildRouterDefaults() map[string]interface{} {
	return map[string]interface{}{
		"MaxRetries":   defaultRouterMaxRetries,
		"RetryTimeout": defaultRouterRetryTimeout,
	}
}

func buildSubscriptionDefaults() map[string]interface{} {
	return map[string]interface{}{
		"BindAddress": defaultSubscriptionAPIAddress,
		"Port":        defaultSubscriptionAPIPort,
		"Scheme":      defaultSubscriptionAPISchemes,
		"BasePath":    defaultSubscriptionAPIBasePath,
	}
}

func buildLogDefaults() map[string]interface{} {
	return map[string]interface{}{
		"loglevel": defaultLogLevel,
	}
}

func buildDatabaseDefaults() map[string]interface{} {
	return map[string]interface{}{
		"Database":         defaultDatabase,
		"ConnectionString": defaultDatabaseConnectionString,
	}
}

func substituteHomeDir(path string) string {
	if homeDir, err := os.UserHomeDir(); err == nil {
		return strings.ReplaceAll(path, "$HOME", homeDir)
	}
	return path
}
