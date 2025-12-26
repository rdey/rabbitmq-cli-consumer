package main

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var exitErrHandlerTests = []struct {
	name string
	err  error
	out  string
	code int
}{
	{
		"empty",
		nil,
		"",
		0,
	},
	{
		"exitCode",
		cli.NewExitError("", 42),
		"",
		42,
	},
	{
		"output",
		fmt.Errorf("normal error"),
		"normal error\n",
		1,
	},
	{
		"exitCodeOutput",
		cli.NewExitError("exit code error", 42),
		"exit code error\n",
		42,
	},
	{
		"outputFormatted",
		errors.WithMessage(fmt.Errorf("error"), "nested"),
		"error\nnested\n",
		1,
	},
}

func TestExitErrHandler(t *testing.T) {
	log.SetFlags(0)
	for _, test := range exitErrHandlerTests {
		t.Run(test.name, func(t *testing.T) {
			var exitCode int
			originalOsExit := osExit
			osExit = func(code int) {
				exitCode = code
			}
			defer func() { osExit = originalOsExit }()

			buf := &bytes.Buffer{}
			log.SetOutput(buf)

			ExitErrHandler(nil, test.err)

			if test.code != 0 {
				assert.Equal(t, test.code, exitCode, "unexpected exit code")
			}
			assert.Equal(t, test.out, buf.String())
		})
	}
}
