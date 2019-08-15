package log

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestLoggers(t *testing.T) {
	testCases := []struct {
		Level
		Expecting string
	}{
		{D, "[DEBUG] D: message\n[INFO] I: message\n[WARN] W: message\n[ERROR] E: message\n"},
		{I, "[INFO] I: message\n[WARN] W: message\n[ERROR] E: message\n"},
		{W, "[WARN] W: message\n[ERROR] E: message\n"},
		{E, "[ERROR] E: message\n"},
		{F, ""},
	}

	var writer bytes.Buffer
	logger := log.New(&writer, "", 0)
	SetLogger(logger)

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("test level: %v", testCase.Level), func(t *testing.T) {
			SetLevel(testCase.Level)
			Debug("D: %s", "message")
			Info("I: %s", "message")
			Warn("W: %s", "message")
			Error("E: %s", "message")

			output := writer.String()
			assert.Equal(t, testCase.Expecting, output, "output at level %v wrong", testCase.Level)

			writer.Reset()
		})
	}
}
