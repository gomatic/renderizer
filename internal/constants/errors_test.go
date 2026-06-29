package constants_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gomatic/renderizer/internal/constants"
)

func TestErrorError(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "missing template name", constants.ErrMissingTemplate.Error())
}

func TestErrorWith(t *testing.T) {
	t.Parallel()
	cause := errors.New("boom")
	tests := []struct {
		err         error
		name        string
		wantMessage string
		args        []any
		wantIsCause bool
	}{
		{
			name:        "bare sentinel",
			err:         nil,
			args:        nil,
			wantMessage: "failed to open template",
		},
		{
			name:        "with cause",
			err:         cause,
			args:        nil,
			wantMessage: "failed to open template: boom",
			wantIsCause: true,
		},
		{
			name:        "with args only",
			err:         nil,
			args:        []any{"file", "x.tmpl"},
			wantMessage: "failed to open template: file x.tmpl",
		},
		{
			name:        "with cause and args",
			err:         cause,
			args:        []any{"x.tmpl"},
			wantMessage: "failed to open template: boom: x.tmpl",
			wantIsCause: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := constants.ErrOpenTemplate.With(tt.err, tt.args...)
			assert.EqualError(t, got, tt.wantMessage)
			assert.True(t, errors.Is(got, constants.ErrOpenTemplate))
			assert.Equal(t, tt.wantIsCause, errors.Is(got, cause))
		})
	}
}

func TestErrorIsDistinct(t *testing.T) {
	t.Parallel()
	assert.False(t, errors.Is(constants.ErrParseTemplate, constants.ErrExecuteTemplate))
	assert.NotErrorIs(t, fmt.Errorf("x"), constants.ErrReadTemplate)
}
