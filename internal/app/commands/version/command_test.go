package version_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app/commands/version"
)

func TestCommand(t *testing.T) {
	t.Parallel()
	cmd := version.Command("renderizer", "1.2.3")
	var stdout bytes.Buffer
	cmd.Writer = &stdout
	err := cmd.Run(context.Background(), []string{"version"})
	require.NoError(t, err)
	assert.Equal(t, "renderizer version 1.2.3\n", stdout.String())
}
