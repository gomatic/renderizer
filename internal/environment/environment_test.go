package environment_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gomatic/renderizer/internal/environment"
)

func TestLoad(t *testing.T) {
	t.Parallel()
	environ := func() []string {
		return []string{"USER=alice", "PATH=/bin:/usr/bin", "EMPTY=", "NOEQUALS"}
	}
	got := environment.Load(environ)
	want := environment.Variables{
		"USER":     "alice",
		"PATH":     "/bin:/usr/bin",
		"EMPTY":    "",
		"NOEQUALS": "",
	}
	assert.Equal(t, want, got)
}
