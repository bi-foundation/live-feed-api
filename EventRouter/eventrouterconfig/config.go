package eventrouterconfig

import (
	"errors"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	defaultConfigName          = "livefeed"
	defaultListenerBindAddress = ""
	defaultListenerPort        = "8040"
	defaultListenerProtocol    = "tcp"

	defaultSubscriptionApiAddress = ""
	defaultSubscriptionApiPort    = "8700"
)

var defaultSubscriptionApiSchemes = []string{"HTTP", "HTTPS"}
var possibleConfigPaths = []string{}

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

func LoadEventRouterConfig() (*EventRouterConfig, error) {
	return LoadEventRouterConfigFrom("")
}

func LoadEventRouterConfigFrom(configFilePath string) (*EventRouterConfig, error) {
	vp := viper.New()
	vp.SetConfigName(defaultConfigName)
	if len(configFilePath) > 0 {
		vp.SetConfigFile(configFilePath)
	} else {
		possibleConfigPaths = append(possibleConfigPaths, "./conf")
		possibleConfigPaths = append(possibleConfigPaths, "/etc/factom-livefeed")
		getHomeDir()
		if len(homeDir) > 0 {
			possibleConfigPaths = append(possibleConfigPaths, "$HOME/.factom/livefeed")
		}
	}

	eventRouterConfig := &EventRouterConfig{}
	for _, path := range possibleConfigPaths {
		vp.AddConfigPath(path)
	}
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.SetEnvPrefix("factomlf")
	vp.AutomaticEnv()
	setDefaults(vp)
	if err := vp.ReadInConfig(); err != nil {
		return nil, reformatConfigFileErrors(err, vp)
	}
	if err := vp.Unmarshal(&eventRouterConfig); err != nil {
		log.Error("could not read configuration file")
	}
	return eventRouterConfig, nil
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
		log.Warn(fmt.Sprintf("the user home directory could not be retrieved, the '$HOME/.factom/livefeed' location will be skipped. error: %v.", err))
	}
}

func substituteHomeDir(path string) string {
	if len(homeDir) > 0 {
		return strings.ReplaceAll(path, "$HOME", homeDir)
	} else {
		return path
	}
}
