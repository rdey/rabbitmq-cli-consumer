package log

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"syscall"
	"testing"

	"github.com/corvus-ch/rabbitmq-cli-consumer/config"
	"github.com/stretchr/testify/assert"
)

var createLoggerTests = []struct {
	name                string
	verbose             bool
	file                string
	expectFileContent   string
	expectBufferContent string
}{
	{"default", false, "default", "default", ""},
	{"verbose", true, "verbose", "verbose", "verbose"},
	{"noFile", true, "", "", "noFile"},
}

func TestCreateLoggerWriter(t *testing.T) {
	for _, test := range createLoggerTests {
		t.Run(test.name, func(t *testing.T) {
			w, f, buf, err := createWriter(test.file, test.verbose)
			if err != nil {
				t.Error(err)
			}
			defer f.Close()
			defer syscall.Unlink(f.Name())

			w.Write([]byte(test.name))

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf("failed to read log output: %v", err)
			}

			assert.Equal(t, test.expectFileContent, string(b))
			assert.Equal(t, test.expectBufferContent, buf.String())
		})
	}
}

var loggersTests = []struct {
	name        string
	config      string
	err         error
	hasDateTime bool
}{
	{
		"noErrorFile",
		"",
		fmt.Errorf("failed creating error log: open : no such file or directory"),
		true,
	},
	{
		"noOutFile",
		`[logs]
error = ./error.log
`,
		fmt.Errorf("failed creating info log: open : no such file or directory"),
		true,
	},
	{
		"success",
		`[logs]
error = ./error.log
info = ./info.log
`,
		nil,
		true,
	},
	{
		"noLogFiles",
		`[logs]
verbose = On
`,
		nil,
		true,
	},
	{
		"noDateTime",
		`[logs]
error = ./error.log
info = ./info.log
nodatetime = On
`,
		nil,
		false,
	},
	{
		"noLogFilesNoDateTime",
		`[logs]
verbose = On
nodatetime = On
`,
		nil,
		false,
	},
}

func TestLoggers(t *testing.T) {
	// Pattern to match date/time prefix like "2024/01/15 10:30:45 "
	dateTimePattern := regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`)

	for _, test := range loggersTests {
		t.Run(test.name, func(t *testing.T) {
			// Clean up any existing log files before test
			os.Remove("./error.log")
			os.Remove("./info.log")

			cfg, _ := config.CreateFromString(test.config)
			l, _, _, err := NewFromConfig(cfg)
			assert.Equal(t, test.err, err)

			if l != nil {
				// Clean up created log files after test
				defer os.Remove("./error.log")
				defer os.Remove("./info.log")

				// Write test messages
				l.Info("test info message")
				l.Error(fmt.Errorf("test error message"))

				// Read log file contents and check for date/time prefix
				infoContent, _ := ioutil.ReadFile("./info.log")
				errorContent, _ := ioutil.ReadFile("./error.log")

				infoHasDateTime := dateTimePattern.Match(infoContent)
				errorHasDateTime := dateTimePattern.Match(errorContent)

				if test.hasDateTime {
					// If files exist, they should have date/time
					if len(infoContent) > 0 {
						assert.True(t, infoHasDateTime, "info log should have date/time prefix, got: %s", string(infoContent))
					}
					if len(errorContent) > 0 {
						assert.True(t, errorHasDateTime, "error log should have date/time prefix, got: %s", string(errorContent))
					}
				} else {
					// Files should not have date/time prefix
					if len(infoContent) > 0 {
						assert.False(t, infoHasDateTime, "info log should not have date/time prefix, got: %s", string(infoContent))
					}
					if len(errorContent) > 0 {
						assert.False(t, errorHasDateTime, "error log should not have date/time prefix, got: %s", string(errorContent))
					}
				}
			}
		})
	}
}

func createWriter(name string, verbose bool) (io.Writer, *os.File, *bytes.Buffer, error) {
	buf := &bytes.Buffer{}

	f, err := ioutil.TempFile("", name)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create temp file: %v", err)
	}

	if len(name) > 0 {
		name = f.Name()
	}

	w, err := newWriter(name, verbose, buf)

	return w, f, buf, err
}
