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
	}{
		{D},
		{I},
		{W},
		{E},
		{F},
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
			assertOutput(t, output, testCase.Level)
			writer.Reset()
		})
	}
}

func assertOutput(t *testing.T, output string, level Level) {
	switch level {
	case D:
		assert.Equal(t, "[DEBUG] D: message\n[INFO] I: message\n[WARN] W: message\n[ERROR] E: message\n", output, "output at level %v wrong", level)
	case I:
		assert.Equal(t, "[INFO] I: message\n[WARN] W: message\n[ERROR] E: message\n", output, "output at level %v wrong", level)
	case W:
		assert.Equal(t, "[WARN] W: message\n[ERROR] E: message\n", output, "output at level %v wrong", level)
	case E:
		assert.Equal(t, "[ERROR] E: message\n", output, "output at level %v wrong", level)
	case F:
		assert.Equal(t, "", output, "output at level %v wrong", level)
	}
}
