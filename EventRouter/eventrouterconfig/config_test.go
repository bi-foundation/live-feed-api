package eventrouterconfig

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
)

const testConfig = `
[eventlistenerconfig]
  bindaddress = "127.0.0.1"
  port = "8044"
  protocol = "udp"

[subscriptionapiconfig]
  bindaddress = "0.0.0.0"
  port = "8777"
  schemes = ["HTTP","HTTPS"]
`

func init() {
	log.SetLevel(log.D)
}

func Test(t *testing.T) {
	files := renameConfigFiles()
	defer restoreRenamedFiles(files)

	t.Run("TestLocateConfigFile", testLocateConfigFile)
	t.Run("TestSpecificConfigFile", testSpecificConfigFile)
	t.Run("testEnvVarOverrides", testEnvVarOverrides)
	t.Run("TestNoConfigFound", testNoConfigFound)
}

func testNoConfigFound(t *testing.T) {
	_, err := LoadEventRouterConfigFrom("")
	if err == nil {
		t.Fatal("Expected a config not found error")
	}
}

func testLocateConfigFile(t *testing.T) {
	testLocateAndVerifyConfigFile(t, "")
}

func testSpecificConfigFile(t *testing.T) {
	testLocateAndVerifyConfigFile(t, "test-config")
}

func testLocateAndVerifyConfigFile(t *testing.T, specifiedConfigFileName string) {
	configFileDir := "$HOME/.factom/livefeed"
	getHomeDir()
	configFileDir = substituteHomeDir(configFileDir)
	err := makeDirs(configFileDir)
	if err != nil {
		message := fmt.Sprintf("the config location %s could not be created, error: %v.", configFileDir, err)
		if os.IsPermission(err) {
			log.Warn(message + "\nyou can create this directory manually using sudo mkdir /etc/factom-livefeed && sudo chown $USER:root /etc/factom-livefeed")
			return
		} else {
			t.Errorf(message)
		}
	}

	var configFile string
	if len(specifiedConfigFileName) == 0 {
		configFile = fmt.Sprint(configFileDir, "/", defaultConfigName, ".toml")
	} else {
		configFile = fmt.Sprint(configFileDir, "/", specifiedConfigFileName, ".conf")
	}
	err = ioutil.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		message := fmt.Sprintf("could not test location %s because we don't have permissions here", configFileDir)
		if os.IsPermission(err) {
			log.Warn(message + "\nyou change the owner of this directory manually using sudo chown $USER:root /etc/factom-livefeed")
			return
		} else {
			t.Errorf("could not create config location %s, error: %v", configFileDir, err)
		}
	}
	defer deleteTestConfigFile(configFile)

	config, err := LoadEventRouterConfigFrom(configFile)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, config) {
		return
	}
	listenerConfig := config.EventListenerConfig
	if !assert.NotNil(t, listenerConfig, "EventListenerConfig shouldn't be nil") {
		return
	}
	assert.EqualValues(t, "127.0.0.1", listenerConfig.BindAddress,
		"EventListenerConfig.BindAddress mismatch %s != %s", "127.0.0.1", listenerConfig.BindAddress)
	assert.EqualValues(t, "8044", strconv.Itoa(int(listenerConfig.Port)),
		"EventListenerConfig.Port mismatch %s != %d", 8044, listenerConfig.Port)
	assert.EqualValues(t, "udp", listenerConfig.Protocol,
		"EventListenerConfig.Protocol mismatch %s != %s", "udp", listenerConfig.Protocol)
	subscriptionConfig := config.SubscriptionApiConfig
	if !assert.NotNil(t, subscriptionConfig, "SubscriptionApiConfig shouldn't be nil") {
		return
	}
	assert.EqualValues(t, "0.0.0.0", subscriptionConfig.BindAddress,
		"SubscriptionApiConfig.BindAddress mismatch %s != %s", "127.0.0.1", subscriptionConfig.BindAddress)
	assert.EqualValues(t, "8777", strconv.Itoa(int(subscriptionConfig.Port)),
		"SubscriptionApiConfig.Port mismatch %s != %d", 8777, subscriptionConfig.Port)
	assert.EqualValues(t, []string{"HTTP", "HTTPS"}, subscriptionConfig.Schemes,
		"SubscriptionApiConfig.Schemes mismatch %v != %v", []string{"HTTP", "HTTPS"}, subscriptionConfig.Schemes)
}

func testEnvVarOverrides(t *testing.T) {
	configFileDir := substituteHomeDir("$HOME/.factom/livefeed")
	configFile := fmt.Sprint(configFileDir, "/", defaultConfigName, ".toml")
	err := ioutil.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		t.Errorf("Could not write config file, error: %v", err)
	}
	defer deleteTestConfigFile(configFile)

	os.Setenv("FACTOMLF_EVENTLISTENERCONFIG_PORT", "8666")
	os.Setenv("FACTOMLF_SUBSCRIPTIONAPICONFIG_SCHEMES", "HTTPS,HTTP")
	config, err := LoadEventRouterConfigFrom(configFile)
	assert.Nil(t, err)
	assert.NotNil(t, config)
	listenerConfig := config.EventListenerConfig
	assert.NotNil(t, listenerConfig, "EventListenerConfig shouldn't be nil")
	assert.EqualValues(t, "127.0.0.1", listenerConfig.BindAddress,
		"EventListenerConfig.BindAddress mismatch %s != %s", "127.0.0.1", listenerConfig.BindAddress)
	assert.EqualValues(t, "8666", strconv.Itoa(int(listenerConfig.Port)),
		"EventListenerConfig.Port mismatch %s != %d", 8666, listenerConfig.Port)
	assert.EqualValues(t, "udp", listenerConfig.Protocol,
		"EventListenerConfig.Protocol mismatch %s != %s", "udp", listenerConfig.Protocol)
	subscriptionConfig := config.SubscriptionApiConfig
	assert.NotNil(t, subscriptionConfig, "SubscriptionApiConfig shouldn't be nil")
	assert.EqualValues(t, "0.0.0.0", subscriptionConfig.BindAddress,
		"SubscriptionApiConfig.BindAddress mismatch %s != %s", "127.0.0.1", subscriptionConfig.BindAddress)
	assert.EqualValues(t, "8777", strconv.Itoa(int(subscriptionConfig.Port)),
		"SubscriptionApiConfig.Port mismatch %s != %d", 8777, subscriptionConfig.Port)
	assert.EqualValues(t, []string{"HTTPS", "HTTP"}, subscriptionConfig.Schemes,
		"SubscriptionApiConfig.Schemes mismatch %v != %v", []string{"HTTPS", "HTTP"}, subscriptionConfig.Schemes)
}

func makeDirs(configFileDir string) error {
	if _, err := os.Stat(configFileDir); os.IsNotExist(err) {
		oldMask := syscall.Umask(0)
		defer syscall.Umask(oldMask)

		if err = os.MkdirAll(configFileDir, os.ModeDir|OS_ALL_RWX); err != nil {
			return err
		}
	}
	return nil
}

func deleteTestConfigFile(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		log.Warn("could not cleanup test config file", fileName)
	}
}

func renameConfigFiles() []string {
	postFix := time.Now().Format("20060102150405")
	result := make([]string, 0)
	result = renameConfigFile("./conf", postFix, result)
	result = renameConfigFile("$HOME/.factom/livefeed", postFix, result)
	result = renameConfigFile("/etc/factom-livefeed", postFix, result)
	return result
}

func restoreRenamedFiles(files []string) {
	for _, backupFile := range files {
		orgFile := backupFile[:len(backupFile)-14]
		err := os.Rename(backupFile, orgFile)
		if err != nil {
			log.Error("could not restore renamed file", backupFile, "to", orgFile)
		}
	}
}

func renameConfigFile(configFileDir string, postFix string, result []string) []string {
	configFile := substituteHomeDir(fmt.Sprint(configFileDir, "/", defaultConfigName, ".toml"))
	_, err := os.Stat(configFile)
	if err == nil {
		backupFile := configFile + postFix
		err := os.Rename(configFile, backupFile)
		if err != nil {
			panic(fmt.Sprintf("Could not restore renamed file %s to %s", configFile, backupFile))
		}
		result = append(result, backupFile)
	}
	return result
}
