package config

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
)

const testConfig = `
[log]
  loglevel = "info"

[router]
  maxretries = 4
  retrytimeout = 20

[receiver]
  bindaddress = "127.0.0.1"
  port = "8044"
  protocol = "tcp"

[subscription]
  bindaddress = "0.0.0.0"
  port = "8777"
  schemes = "HTTP"
`

func init() {
	log.SetLevel(log.D)
}

func Test(t *testing.T) {
	// make sure the test doesn't override or change configuration files on disk
	files := renameConfigFiles(t)
	defer restoreRenamedFiles(t, files)

	testCases := map[string]func(*testing.T){
		"TestSpecificConfigFile":     testSpecificConfigFile,
		"TestWrongTypeConfigFile":    testWrongTypeConfigFile,
		"TestConfigFileInConfigHome": testConfigFileInConfigHome,
		"TestEnvVarOverrides":        testEnvVarOverrides,
		"TestDefaultConfig":          testDefaultConfig,
		"TestNoConfigFound":          testNoConfigFound,
	}

	for name, testCase := range testCases {
		t.Run(name, testCase)
	}
}

func testSpecificConfigFile(t *testing.T) {
	configFile, cleanup := createTempConfigFile(t, testConfig)
	defer cleanup()

	config, err := LoadConfigurationFrom(configFile)

	assertConfiguration(t, config, err)
}

func testWrongTypeConfigFile(t *testing.T) {
	testConfig := `
		[router]
		  retrytimeout = "noint"
		`
	configFile, cleanup := createTempConfigFile(t, testConfig)
	defer cleanup()

	_, err := LoadConfigurationFrom(configFile)

	assert.Error(t, err, `could not read configuration file: 1 error(s) decoding:\n\n* cannot parse 'Router.RetryTimeout' as uint: strconv.ParseUint: parsing \"noint\": invalid syntax`)
}

func testConfigFileInConfigHome(t *testing.T) {
	configFilename := "$HOME/.factom/factom-live-feed.conf"
	cleanup := createConfigFile(t, configFilename)
	defer cleanup()

	config, err := LoadConfiguration()

	assertConfiguration(t, config, err)
}

func testEnvVarOverrides(t *testing.T) {
	configFilename := "$HOME/.factom/live-feed/factom-live-feed.conf"
	cleanup := createConfigFile(t, configFilename)
	defer cleanup()

	os.Setenv("FACTOM_LIVE_FEED_RECEIVER_PORT", "8666")
	os.Setenv("FACTOM_LIVE_FEED_SUBSCRIPTION_SCHEMES", "HTTPS")

	defer func() {
		os.Unsetenv("FACTOM_LIVE_FEED_RECEIVER_PORT")
		os.Unsetenv("FACTOM_LIVE_FEED_SUBSCRIPTION_SCHEMES")
	}()

	config, err := LoadConfigurationFrom(configFilename)
	if !assert.Nil(t, err) {
		t.Fatalf("load config failed: %v", err)
	}

	if !assert.NotNil(t, config) {
		t.FailNow()
	}

	assert.EqualValues(t, "info", config.Log.LogLevel)

	receiverConfig := config.Receiver
	assert.NotNil(t, receiverConfig, "ReceiverConfig shouldn't be nil")
	assert.EqualValues(t, "127.0.0.1", receiverConfig.BindAddress, "ReceiverConfig.BindAddress mismatch %s != %s", "127.0.0.1", receiverConfig.BindAddress)
	assert.EqualValues(t, "8666", strconv.Itoa(int(receiverConfig.Port)), "ReceiverConfig.Port mismatch %s != %d", 8666, receiverConfig.Port)
	assert.EqualValues(t, "tcp", receiverConfig.Protocol, "ReceiverConfig.Protocol mismatch %s != %s", "tcp", receiverConfig.Protocol)

	routerConfig := config.Router
	assert.NotNil(t, routerConfig, "routerConfig shouldn't be nil")
	assert.EqualValues(t, uint16(4), routerConfig.MaxRetries)
	assert.EqualValues(t, uint(20), routerConfig.RetryTimeout)

	subscriptionConfig := config.Subscription
	assert.NotNil(t, subscriptionConfig, "SubscriptionConfig shouldn't be nil")
	assert.EqualValues(t, "0.0.0.0", subscriptionConfig.BindAddress, "SubscriptionConfig.BindAddress mismatch %s != %s", "127.0.0.1", subscriptionConfig.BindAddress)
	assert.EqualValues(t, "8777", strconv.Itoa(int(subscriptionConfig.Port)), "SubscriptionConfig.Port mismatch %s != %d", 8777, subscriptionConfig.Port)
	assert.EqualValues(t, "HTTP", subscriptionConfig.Scheme, "SubscriptionConfig.Schemes mismatch %v != %v", []string{"HTTPS"}, subscriptionConfig.Scheme)
}

func testDefaultConfig(t *testing.T) {
	config, err := LoadConfiguration()
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	if !assert.NotNil(t, config) {
		t.FailNow()
	}

	assert.EqualValues(t, "info", config.Log.LogLevel)

	receiverConfig := config.Receiver
	assert.NotNil(t, receiverConfig, "ReceiverConfig shouldn't be nil")
	assert.EqualValues(t, defaultReceiverBindAddress, receiverConfig.BindAddress, "ReceiverConfig.BindAddress mismatch %s != %s", defaultReceiverBindAddress, receiverConfig.BindAddress)
	assert.EqualValues(t, defaultReceiverPort, receiverConfig.Port, "ReceiverConfig.Port mismatch %s != %d", defaultReceiverPort, receiverConfig.Port)
	assert.EqualValues(t, defaultReceiverProtocol, receiverConfig.Protocol, "ReceiverConfig.Protocol mismatch %s != %s", defaultReceiverProtocol, receiverConfig.Protocol)

	routerConfig := config.Router
	assert.NotNil(t, routerConfig, "routerConfig shouldn't be nil")
	assert.EqualValues(t, defaultRouterMaxRetries, routerConfig.MaxRetries, "routerConfig.MaxRetries mismatch %s != %s", defaultRouterMaxRetries, routerConfig.MaxRetries)
	assert.EqualValues(t, defaultRouterRetryTimeout, routerConfig.RetryTimeout, "routerConfig.RetryTimeout mismatch %s != %d", defaultRouterRetryTimeout, routerConfig.RetryTimeout)

	subscriptionConfig := config.Subscription
	assert.NotNil(t, subscriptionConfig, "SubscriptionConfig shouldn't be nil")
	assert.EqualValues(t, defaultSubscriptionAPIAddress, subscriptionConfig.BindAddress, "SubscriptionConfig.BindAddress mismatch %s != %s", defaultSubscriptionAPIAddress, subscriptionConfig.BindAddress)
	assert.EqualValues(t, defaultSubscriptionAPIPort, subscriptionConfig.Port, "SubscriptionConfig.Port mismatch %s != %d", defaultSubscriptionAPIPort, subscriptionConfig.Port)
	assert.EqualValues(t, defaultSubscriptionAPISchemes, subscriptionConfig.Scheme, "SubscriptionConfig.Schemes mismatch %v != %v", defaultSubscriptionAPISchemes, subscriptionConfig.Scheme)
}

func testNoConfigFound(t *testing.T) {
	_, err := LoadConfigurationFrom("not-exists.conf")
	if err == nil {
		t.Fatal("expected a config not found error")
	}
}

func assertConfiguration(t *testing.T, config *Config, err error) {
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	if !assert.NotNil(t, config) {
		t.FailNow()
	}
	if !assert.NotNil(t, config.Log) {
		t.FailNow()
	}
	assert.EqualValues(t, "info", config.Log.LogLevel)

	receiverConfig := config.Receiver
	if !assert.NotNil(t, receiverConfig, "ReceiverConfig shouldn't be nil") {
		t.FailNow()
	}
	assert.EqualValues(t, "127.0.0.1", receiverConfig.BindAddress, "ReceiverConfig.BindAddress mismatch %s != %s", "127.0.0.1", receiverConfig.BindAddress)
	assert.EqualValues(t, "8044", strconv.Itoa(int(receiverConfig.Port)), "ReceiverConfig.Port mismatch %s != %d", 8044, receiverConfig.Port)
	assert.EqualValues(t, "tcp", receiverConfig.Protocol, "ReceiverConfig.Protocol mismatch %s != %s", "tcp", receiverConfig.Protocol)

	routerConfig := config.Router
	if !assert.NotNil(t, routerConfig, "RouterConfig shouldn't be nil") {
		t.FailNow()
	}
	assert.EqualValues(t, uint16(4), routerConfig.MaxRetries)
	assert.EqualValues(t, uint(20), routerConfig.RetryTimeout)

	subscriptionConfig := config.Subscription
	if !assert.NotNil(t, subscriptionConfig, "SubscriptionConfig shouldn't be nil") {
		t.FailNow()
	}
	assert.EqualValues(t, "0.0.0.0", subscriptionConfig.BindAddress, "SubscriptionConfig.BindAddress mismatch %s != %s", "127.0.0.1", subscriptionConfig.BindAddress)
	assert.EqualValues(t, "8777", strconv.Itoa(int(subscriptionConfig.Port)), "SubscriptionConfig.Port mismatch %s != %d", 8777, subscriptionConfig.Port)
	assert.EqualValues(t, "HTTP", subscriptionConfig.Scheme, "SubscriptionConfig.Schemes mismatch %v != %v", []string{"HTTP", "HTTPS"}, subscriptionConfig.Scheme)
}

func createTempConfigFile(t *testing.T, testConfig string) (string, func()) {
	oldMask := syscall.Umask(0)
	defer syscall.Umask(oldMask)
	file, err := ioutil.TempFile("", "test.*.conf")
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	err = ioutil.WriteFile(file.Name(), []byte(testConfig), os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cleanup := func() {
		os.Remove(file.Name())
	}

	return file.Name(), cleanup
}

func createConfigFile(t *testing.T, configFilename string) func() {
	oldMask := syscall.Umask(0)
	defer syscall.Umask(oldMask)

	configFilename = substituteHomeDir(configFilename)

	// create directories
	dir, _ := filepath.Split(configFilename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			message := fmt.Sprintf("the config location %s could not be created, error: %v.", dir, err)
			if os.IsPermission(err) {
				t.Fatalf("%s\nyou can create this directory manually using sudo mkdir /etc/factom-livefeed && sudo chown $USER:root /etc/factom-livefeed", message)
			}
			t.Fatal(message)
		}
	}

	err := ioutil.WriteFile(configFilename, []byte(testConfig), os.ModePerm)
	if err != nil {
		if os.IsPermission(err) {
			t.Fatalf("could not test location %s because we don't have permissions here", configFilename)
		}
		t.Fatalf("could not write config file %s, error: %v", configFilename, err)
	}

	cleanup := func() {
		err := os.Remove(configFilename)
		if err != nil {
			t.Logf("could not cleanup test config file: %s", configFilename)
		}
	}

	return cleanup
}

func renameConfigFiles(t *testing.T) map[string]string {
	configFiles := map[string]string{}

	possiblePaths := []string{
		"",
		substituteHomeDir("$HOME/.factom/"),
		substituteHomeDir(fmt.Sprintf("$HOME/.factom/%s", defaultConfigDir)),
		fmt.Sprintf("/etc/%s", defaultConfigDir),
	}

	for _, path := range possiblePaths {
		configFilename := substituteHomeDir(filepath.Join(path, defaultConfigName))

		// detect default config file
		if _, err := os.Stat(configFilename); !os.IsNotExist(err) {
			tmpFile, err := ioutil.TempFile("", fmt.Sprintf("tmp-config.*.%s", defaultConfigName))
			if err != nil {
				t.Fatalf("failed to create temp file for existing config file: %v", err)
			}

			if err := moveFile(configFilename, tmpFile.Name()); err != nil {
				t.Fatalf("failed to move existing config file: %v", err)
			}

			configFiles[configFilename] = tmpFile.Name()
		}
	}

	return configFiles
}

func restoreRenamedFiles(t *testing.T, files map[string]string) {
	for configFile, backupFile := range files {
		err := moveFile(backupFile, configFile)
		if err != nil {
			t.Errorf("Could not restore renamed file %s to %s", configFile, backupFile)
		}
	}
}

func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}
