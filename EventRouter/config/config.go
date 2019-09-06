package config

import (
	"errors"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/spf13/viper"
	"os"
	"path"
	"strings"
)

const (
	defaultConfigName = "livefeed"

	defaultLogLevel = log.D

	defaultListenerBindAddress = ""
	defaultListenerPort        = "8040"
	defaultListenerProtocol    = "tcp"

	defaultSubscriptionApiAddress = ""
	defaultSubscriptionApiPort    = "8700"
)

var defaultSubscriptionApiSchemes = []string{"HTTP", "HTTPS"}
var possibleConfigPaths = []string{}

type Config struct {
	Log                *LogConfig
	ReceiverConfig     *ReceiverConfig
	SubscriptionConfig *SubscriptionConfig
}

type LogConfig struct {
	LogLevel log.Level
}

type ReceiverConfig struct {
	Protocol    string
	BindAddress string
	Port        uint16
}

type SubscriptionConfig struct {
	Schemes     []string
	BindAddress string // (Duplicated because extended interfaces are not supported by Viper)
	Port        uint16
}

var homeDir string

func LoadConfiguration() (*Config, error) {
	return LoadConfigurationFrom("")
}

// read configuration from file
func LoadConfigurationFrom(configFilePath string) (*Config, error) {
	vp := viper.New()

	vp.SetConfigName(defaultConfigName)
	if len(configFilePath) > 0 {
		vp.SetConfigFile(configFilePath)

		// use toml config structure
		if strings.HasPrefix(path.Ext(configFilePath), ".conf") {
			vp.SetConfigType("toml")
		}
	} else {
		// look for configuration in default paths if user doesn't give configuration argument
		possibleConfigPaths = append(possibleConfigPaths, "./conf")
		possibleConfigPaths = append(possibleConfigPaths, "/etc/factom-livefeed")
		getHomeDir()
		if len(homeDir) > 0 {
			possibleConfigPaths = append(possibleConfigPaths, "$HOME/.factom/livefeed")
		}
	}

	// read/build configuration
	config := &Config{}
	for _, path := range possibleConfigPaths {
		vp.AddConfigPath(path)
	}
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.SetEnvPrefix("factomlf")
	vp.AutomaticEnv()

	// set default configuration values
	setDefaults(vp)

	// read configuration and override defaults if needed
	if err := vp.ReadInConfig(); err != nil {
		return nil, reformatConfigFileErrors(err, vp)
	}

	if err := vp.Unmarshal(&config); err != nil {
		log.Error("could not read configuration file")
	}

	return config, nil
}

func reformatConfigFileErrors(readErr error, vp *viper.Viper) error {
	var builder strings.Builder
	fmt.Fprintln(&builder, "no configuration file could be loaded from one of the following locations:")
	for _, path := range possibleConfigPaths {
		fmt.Fprintln(&builder, "\t", path)
	}
	fmt.Fprintf(&builder, "error: %v", readErr)
	return errors.New(builder.String())
}

func setDefaults(vp *viper.Viper) {
	vp.SetDefault("log", buildReceiverDefaults())
	vp.SetDefault("receiver", buildReceiverDefaults())
	vp.SetDefault("subscription", buildSubscriptionDefaults())
}

func buildReceiverDefaults() map[string]interface{} {
	return map[string]interface{}{
		"Protocol":    defaultListenerProtocol,
		"BindAddress": defaultListenerBindAddress,
		"Port":        defaultListenerPort,
	}
}

func buildSubscriptionDefaults() map[string]interface{} {
	return map[string]interface{}{
		"BindAddress": defaultSubscriptionApiAddress,
		"Port":        defaultSubscriptionApiPort,
		"Schemes":     defaultSubscriptionApiSchemes,
	}
}

func buildLogDefaults() map[string]interface{} {
	return map[string]interface{}{
		"loglevel": defaultLogLevel,
	}
}

func getHomeDir() {
	var err error
	if homeDir, err = os.UserHomeDir(); err != nil {
		log.Warn("the user home directory could not be retrieved, the '$HOME/.factom/livefeed' location will be skipped. error: %v.", err)
	}
}

func substituteHomeDir(path string) string {
	if len(homeDir) > 0 {
		return strings.ReplaceAll(path, "$HOME", homeDir)
	} else {
		return path
	}
}
