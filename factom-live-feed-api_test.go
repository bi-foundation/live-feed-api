package main

import (
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestSetupMysqlDatabaseFailed(t *testing.T) {
	testable := func() {
		configuration := &config.DatabaseConfig{
			Database:         "mysql",
			ConnectionString: "",
		}
		setupDatabase(configuration)
	}
	output := testApplicationExit(t, "TestSetupMysqlDatabaseFailed", testable)

	t.Log(output)
	assert.Contains(t, output, "[FATAL] failed to configure database: ping failed")
}

func TestSetupNoDatabase(t *testing.T) {
	testable := func() {
		configuration := &config.DatabaseConfig{
			Database:         "",
			ConnectionString: "",
		}
		setupDatabase(configuration)
	}
	output := testApplicationExit(t, "TestSetupNoDatabase", testable)

	t.Log(output)
	assert.Contains(t, output, "[FATAL] failed to configure database:")
}

func TestSetupUnknownDatabase(t *testing.T) {
	testable := func() {
		configuration := &config.DatabaseConfig{
			Database:         "something",
			ConnectionString: "",
		}
		setupDatabase(configuration)
	}
	output := testApplicationExit(t, "TestSetupUnknownDatabase", testable)

	t.Log(output)
	assert.Contains(t, output, "[FATAL] failed to configure database: something")
}

func TestUnknownConfigFileArgument(t *testing.T) {
	testable := func() {
		configuration := loadConfiguration()
		t.Log(configuration)
	}
	output := testApplicationExit(t, "TestUnknownConfigFileArgument", testable, "--config-file", "unknown-file")

	t.Log(output)
	assert.Contains(t, output, "[FATAL] failed to find configuration: 'unknown-file'")
}

// testApplicationExit start a subprocess calling the test function again. The subprocess will execute the actual test
// function that will produce an expected fatal error. The original test process will catch the output and returns this
// to assert the output.
func testApplicationExit(t *testing.T, name string, testableFunction func(), args ...string) string {
	// Only run the failing part when a specific env variable is set
	if os.Getenv("EXITER") == "1" {
		for _, arg := range args {
			os.Args = append(os.Args, arg)
		}
		testableFunction()
	}

	// Start the actual test in a different subprocess
	cmd := exec.Command(os.Args[0], "-test.run="+name)
	cmd.Env = append(os.Environ(), "EXITER=1")
	stdout, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Check that the log fatal message is what we expected
	output, _ := ioutil.ReadAll(stdout)

	defer func() {
		// Check that the program exited
		err := cmd.Wait()
		if e, ok := err.(*exec.ExitError); !ok || e.Success() {
			t.Fatalf("Process ran with err %v, want exit status 1", err)
		}
	}()

	return string(output)
}
