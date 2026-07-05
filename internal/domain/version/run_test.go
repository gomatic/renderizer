package version_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	version "github.com/gomatic/renderizer/internal/domain/version"
)

func TestRun(t *testing.T) {
	t.Parallel()
	result, err := version.Run(context.Background(), nil, version.Config{App: "renderizer", Build: "1.2.3"})
	require.NoError(t, err)
	assert.Equal(t, "renderizer version 1.2.3\n", string(result.Output))
}

func TestRunEmptyBuild(t *testing.T) {
	t.Parallel()
	result, err := version.Run(context.Background(), nil, version.Config{App: "renderizer"})
	require.NoError(t, err)
	assert.Equal(t, "renderizer version \n", string(result.Output))
}
