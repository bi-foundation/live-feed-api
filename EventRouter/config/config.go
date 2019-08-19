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
	defaultConfigFileLocation  = "$HOME/.factom/livefeed"
	defaultListenerBindAddress = ""
	defaultListenerPort        = "8040"
	defaultListenerProtocol    = "tcp"

	defaultSubscriptionApiAddress = ""
	defaultSubscriptionApiPort    = "8700"
)

var defaultSubscriptionApiSchemes = []string{"HTTP", "HTTPS"}

type EventRouterConfig struct {
	EventListenerConfig   *ListenerConfig
	SubscriptionApiConfig *SubscriptionApiConfig
}

type ListenerConfig struct {
	Protocol    string
	BindAddress string
	Port        uint16
}

type SubscriptionApiConfig struct {
	BindAddress string
	Port        uint16
	Schemes     []string
}

func LoadEventRouterConfig() *EventRouterConfig {
	eventRouterConfig := &EventRouterConfig{}
	vp := viper.New()
	vp.SetConfigName(configName)
	vp.AddConfigPath("./conf")
	vp.AddConfigPath("$HOME/.factom/livefeed")
	vp.AddConfigPath("/etc/factom-livefeed")
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.SetEnvPrefix("FACTOMLF")
	vp.AutomaticEnv()
	setDefaults(vp)
	if err := vp.ReadInConfig(); err != nil {
		handleConfigFileErrors(err, vp)
	}
	if err := vp.Unmarshal(&eventRouterConfig); err != nil {
		log.Error("Could not read configuration file.")
	}
	return eventRouterConfig
}

func handleConfigFileErrors(readErr error, vp *viper.Viper) {
	log.Warn("No configuration file could be found, running with default values.")

	if _, ok := readErr.(viper.ConfigFileNotFoundError); ok {
		var home string
		var err error
		if home, err = os.UserHomeDir(); err != nil {
			log.Warn("The user home directory could not be fetched. Error: %v.", err)
			return
		}

		configFileLocation := strings.ReplaceAll(defaultConfigFileLocation, "$HOME", home)
		if _, err = os.Stat(configFileLocation); os.IsNotExist(err) {
			oldMask := syscall.Umask(0)
			defer syscall.Umask(oldMask)

			if err = os.MkdirAll(configFileLocation, os.ModeDir|OS_ALL_RWX); err != nil {
				log.Warn("The config location %s could not be created. Error: %v.", configFileLocation, err)
				return
			}
		}
		configFile := fmt.Sprint(configFileLocation, "/", configName, ".toml")

		if err = vp.WriteConfigAs(configFile); err != nil {
			log.Warn("A default config file could not be written to %s. Error: %v.", configFile, err)
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
		"Port":        defaultListenerPort}
}

func buildSubscriptionApiDefaults() map[string]interface{} {
	return map[string]interface{}{
		"BindAddress": defaultSubscriptionApiAddress,
		"Port":        defaultSubscriptionApiPort,
		"Schemes":     defaultSubscriptionApiSchemes,
	}
}
