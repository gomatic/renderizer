package variables_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gomatic/renderizer/internal/variables"
)

func TestTokenize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		cliArgs []string
		assigns []string
	}{
		{
			name: "empty",
		},
		{
			name:    "arbitrary long variable is an assignment",
			args:    []string{"--name=World"},
			assigns: []string{"--name=World"},
		},
		{
			name:    "bare arbitrary long variable",
			args:    []string{"--flag"},
			assigns: []string{"--flag"},
		},
		{
			name:    "capitalize toggle is an assignment",
			args:    []string{"-C"},
			assigns: []string{"-C"},
		},
		{
			name:    "known long flag passes through",
			args:    []string{"--verbose"},
			cliArgs: []string{"--verbose"},
		},
		{
			name:    "known long alias passes through",
			args:    []string{"--debug"},
			cliArgs: []string{"--debug"},
		},
		{
			name:    "known value long flag with equals",
			args:    []string{"--settings=a.yaml"},
			cliArgs: []string{"--settings=a.yaml"},
		},
		{
			name:    "known value long flag with space",
			args:    []string{"--settings", "a.yaml"},
			cliArgs: []string{"--settings", "a.yaml"},
		},
		{
			name:    "short flags pass through to cli",
			args:    []string{"-S", "a.yaml", "-V"},
			cliArgs: []string{"-S", "a.yaml", "-V"},
		},
		{
			name:    "templates and subcommands pass through",
			args:    []string{"analyze", "file.tmpl"},
			cliArgs: []string{"analyze", "file.tmpl"},
		},
		{
			name:    "mixed stream",
			args:    []string{"--name=World", "-C", "--verbose", "t.tmpl", "--settings", "s.yaml"},
			cliArgs: []string{"--verbose", "t.tmpl", "--settings", "s.yaml"},
			assigns: []string{"--name=World", "-C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := variables.Tokenize(tt.args)
			assert.Equal(t, tt.cliArgs, got.Args)
			assert.Equal(t, tt.assigns, got.Assignments)
		})
	}
}
