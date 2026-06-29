package settings_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/settings"
	"github.com/gomatic/renderizer/internal/variables"
)

const timeFormat = variables.TimeFormat("20060102T150405")

// reader builds a fake ReadFile from a path→content map; an absent path returns
// os.ErrNotExist.
func reader(files map[string]string) settings.ReadFile {
	return func(name string) ([]byte, error) {
		content, ok := files[name]
		if !ok {
			return nil, os.ErrNotExist
		}
		return []byte(content), nil
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()
	read := reader(map[string]string{
		"a.yaml":   "name: First\nitems:\n  - one\n  - two\n",
		"b.yaml":   "items:\n  - three\n",
		"bad.yaml": "::: not yaml :::",
	})

	tests := []struct {
		wantErr error
		want    variables.Context
		name    string
		files   []settings.File
	}{
		{
			name:  "single file",
			files: []settings.File{{Path: "a.yaml"}},
			want:  variables.Context{"name": "First", "items": []any{"one", "two"}},
		},
		{
			name:  "merge appends slices",
			files: []settings.File{{Path: "a.yaml"}, {Path: "b.yaml"}},
			want:  variables.Context{"name": "First", "items": []any{"one", "two", "three"}},
		},
		{
			name:  "optional missing is skipped",
			files: []settings.File{{Path: "missing.yaml", Optional: true}},
			want:  variables.Context{},
		},
		{
			name:    "required missing errors",
			files:   []settings.File{{Path: "missing.yaml"}},
			wantErr: constants.ErrReadSettings,
		},
		{
			name:    "invalid yaml errors",
			files:   []settings.File{{Path: "bad.yaml"}},
			wantErr: constants.ErrParseSettings,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := settings.Load(read, tt.files, timeFormat)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadMergeConflict(t *testing.T) {
	t.Parallel()
	read := reader(map[string]string{
		"slice.yaml": "a:\n  - 1\n",
		"map.yaml":   "a:\n  b: 2\n",
	})
	_, err := settings.Load(read, []settings.File{{Path: "map.yaml"}, {Path: "slice.yaml"}}, timeFormat)
	require.ErrorIs(t, err, constants.ErrMergeContext)
}
