package config

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/spf13/viper"
	"os"
	"strings"
	"syscall"
)

const (
	configName                 = "livefeed"
	defaultConfigFileDir       = "$HOME/.factom/livefeed"
	defaultListenerBindAddress = ""
	defaultListenerPort        = "8040"
	defaultListenerProtocol    = "tcp"

	defaultSubscriptionApiAddress = ""
	defaultSubscriptionApiPort    = "8700"
)

var defaultSubscriptionApiSchemes = []string{"HTTP", "HTTPS"}
var possibleConfigPaths = []string{"./conf", "$HOME/.factom/livefeed", "/etc/factom-livefeed"}

type EventRouterConfig struct {
	EventListenerConfig   *EventListenerConfig
	SubscriptionApiConfig *SubscriptionApiConfig
}

type EventListenerConfig struct {
	Protocol    string
	BindAddress string
	Port        uint16
}

type SubscriptionApiConfig struct {
	Schemes     []string
	BindAddress string // (Duplicated because extended interfaces are not supported by Viper)
	Port        uint16
}

var homeDir string

func LoadEventRouterConfig() *EventRouterConfig {
	getHomeDir()

	eventRouterConfig := &EventRouterConfig{}
	vp := viper.New()
	vp.SetConfigName(configName)
	for _, path := range possibleConfigPaths {
		vp.AddConfigPath(path)
	}
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.SetEnvPrefix("factomlf")
	vp.AutomaticEnv()
	setDefaults(vp)
	if err := vp.ReadInConfig(); err != nil {
		handleConfigFileErrors(err, vp)
	}
	if err := vp.Unmarshal(&eventRouterConfig); err != nil {
		log.Error("could not read configuration file")
	}
	return eventRouterConfig
}

func handleConfigFileErrors(readErr error, vp *viper.Viper) {
	log.Warn("no configuration file could be found, running with default values")
	if _, ok := readErr.(viper.ConfigFileNotFoundError); ok {
		configFileDir := substituteHomeDir(defaultConfigFileDir)
		if _, err := os.Stat(configFileDir); os.IsNotExist(err) {
			oldMask := syscall.Umask(0)
			defer syscall.Umask(oldMask)

			if err = os.MkdirAll(configFileDir, os.ModeDir|OS_ALL_RWX); err != nil {
				log.Warn("the config file directory %s could not be created, error: %v.", configFileDir, err)
				return
			}
		}

		configFile := fmt.Sprint(configFileDir, "/", configName, ".toml")
		if err := vp.WriteConfigAs(configFile); err != nil {
			log.Warn("a default config file could not be written to %s, error: %v.", configFile, err)
			return
		}
	}
}

func setDefaults(vp *viper.Viper) {
	vp.SetDefault("EventListenerConfig", buildEventListenerDefaults())
	vp.SetDefault("SubscriptionApiConfig", buildSubscriptionApiDefaults())
}

func buildEventListenerDefaults() map[string]interface{} {
	return map[string]interface{}{
		"Protocol":    defaultListenerProtocol,
		"BindAddress": defaultListenerBindAddress,
		"Port":        defaultListenerPort,
	}
}

func buildSubscriptionApiDefaults() map[string]interface{} {
	return map[string]interface{}{
		"BindAddress": defaultSubscriptionApiAddress,
		"Port":        defaultSubscriptionApiPort,
		"Schemes":     defaultSubscriptionApiSchemes,
	}
}

func getHomeDir() {
	var err error
	if homeDir, err = os.UserHomeDir(); err != nil {
		panic(fmt.Sprintf("the user home directory could not be retrieved, error: %v.", err))
	}
}

func substituteHomeDir(path string) string {
	return strings.ReplaceAll(path, "$HOME", homeDir)
}
