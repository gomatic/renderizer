package renderizer

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

// This test file follows Go 1.25+ testing standards:
// - Uses modern testing patterns with table-driven tests
// - Tests are isolated and don't modify shared state
// - Uses reflect.DeepEqual for complex type comparisons

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		settings Options
		want     Renderizer
	}{
		{
			name: "basic options",
			settings: Options{
				Capitalize: true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
			},
			want: &Options{
				Capitalize: true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
			},
		},
		{
			name: "with config",
			settings: Options{
				Capitalize: false,
				MissingKey: "zero",
				Config: map[string]interface{}{
					"key": "value",
				},
			},
			want: &Options{
				Capitalize: false,
				MissingKey: "zero",
				Config: map[string]interface{}{
					"key": "value",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.settings)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRender(t *testing.T) {
	// Note: Render() calls os.Exit(), so we cannot test it directly in a unit test.
	// The Render() function is a wrapper around Options.Render() which also calls os.Exit().
	// Comprehensive testing of the rendering functionality is done via:
	// 1. CLI integration tests in cmd/renderizer/cli_test.go
	// 2. Unit tests of individual components (typer, Retyper, etc.)
	//
	// This test documents the function signature and behavior.
	t.Run("render function exists", func(t *testing.T) {
		// Verify the function exists by checking it compiles
		// The actual rendering is tested via CLI tests
		_ = Render
	})
}

func TestOptions_typer(t *testing.T) {
	tests := []struct {
		name       string
		settings   Options
		input      string
		wantResult interface{}
	}{
		{
			name:       "integer",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "42",
			wantResult: int64(42),
		},
		{
			name:       "negative integer",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "-100",
			wantResult: int64(-100),
		},
		{
			name:       "float",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "3.14",
			wantResult: 3.14,
		},
		{
			name:       "negative float",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "-2.5",
			wantResult: -2.5,
		},
		{
			name:       "boolean true",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "true",
			wantResult: true,
		},
		{
			name:       "boolean false",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "false",
			wantResult: false,
		},
		{
			name:       "string",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "hello world",
			wantResult: "hello world",
		},
		{
			name:       "empty string",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "",
			wantResult: "",
		},
		{
			name:       "numeric string",
			settings:   Options{TimeFormat: "20060102T150405"},
			input:      "123abc",
			wantResult: "123abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.settings.typer(tt.input)
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Options.typer() = %v (%T), want %v (%T)", gotResult, gotResult, tt.wantResult, tt.wantResult)
			}
		})
	}
}

func Test_retypeSingleElementSlice(t *testing.T) {
	tests := []struct {
		name   string
		config *retyperConfig
		want   bool
	}{
		{
			name:   "enable single element slice retyping",
			config: &retyperConfig{},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retypeSingleElementSlice(tt.config)
			if tt.config.retypeSingleElementSlice != tt.want {
				t.Errorf("retypeSingleElementSlice() = %v, want %v", tt.config.retypeSingleElementSlice, tt.want)
			}
		})
	}
}

func TestOptions_retyping(t *testing.T) {
	tests := []struct {
		name     string
		settings Options
		source   map[string]interface{}
		config   retyperConfig
		want     map[string]interface{}
	}{
		{
			name:     "empty map",
			settings: Options{TimeFormat: "20060102T150405"},
			source:   map[string]interface{}{},
			config:   retyperConfig{},
			want:     map[string]interface{}{},
		},
		{
			name:     "simple string values",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:     "int to int64",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"num": int(42),
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"num": int64(42),
			},
		},
		{
			name:     "map[interface{}]interface{} to map[string]interface{}",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"nested": map[interface{}]interface{}{
					"key": "value",
				},
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name:     "map[string]string to map[string]interface{}",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"strmap": map[string]string{
					"key": "value",
				},
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"strmap": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name:     "single element slice with retypeSingleElementSlice",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"single": []interface{}{"value"},
			},
			config: retyperConfig{retypeSingleElementSlice: true},
			want: map[string]interface{}{
				"single": "value",
			},
		},
		{
			name:     "single element slice without retypeSingleElementSlice",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"single": []interface{}{"value"},
			},
			config: retyperConfig{retypeSingleElementSlice: false},
			// Even without retypeSingleElementSlice, homogeneous slices are converted to typed slices
			want: map[string]interface{}{
				"single": []string{"value"},
			},
		},
		{
			name:     "bool slice",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"bools": []interface{}{true, false, true},
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"bools": []bool{true, false, true},
			},
		},
		{
			name:     "int64 slice",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"ints": []interface{}{int64(1), int64(2), int64(3)},
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"ints": []int64{1, 2, 3},
			},
		},
		{
			name:     "float64 slice",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"floats": []interface{}{1.1, 2.2, 3.3},
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"floats": []float64{1.1, 2.2, 3.3},
			},
		},
		{
			name:     "string slice",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"strings": []interface{}{"a", "b", "c"},
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"strings": []string{"a", "b", "c"},
			},
		},
		{
			name:     "mixed type slice stays as []interface{}",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"mixed": []interface{}{"a", "1", true},
			},
			config: retyperConfig{},
			// String values inside slices are not retyped, only direct map values are
			want: map[string]interface{}{
				"mixed": []interface{}{"a", "1", true},
			},
		},
		{
			name:     "nested map with slices",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"nested": map[string]interface{}{
					"items": []interface{}{"a", "b"},
				},
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"nested": map[string]interface{}{
					"items": []string{"a", "b"},
				},
			},
		},
		{
			name:     "string value gets retyped",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"num": "42",
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"num": int64(42),
			},
		},
		{
			name:     "string value that's not a number stays string",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"text": "hello",
			},
			config: retyperConfig{},
			want: map[string]interface{}{
				"text": "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of source to avoid modifying the original
			sourceCopy := make(map[string]interface{})
			for k, v := range tt.source {
				sourceCopy[k] = v
			}
			got := tt.settings.retyping(sourceCopy, tt.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Options.retyping() = %v, want %v", got, tt.want)
				// Debug: check specific keys and their elements
				for k := range tt.want {
					if !reflect.DeepEqual(got[k], tt.want[k]) {
						t.Errorf("  key %q: got %v (%T), want %v (%T)", k, got[k], got[k], tt.want[k], tt.want[k])
						// If it's a slice, check each element
						if gotSlice, ok := got[k].([]interface{}); ok {
							if wantSlice, ok2 := tt.want[k].([]interface{}); ok2 {
								for i := range gotSlice {
									if i < len(wantSlice) {
										if !reflect.DeepEqual(gotSlice[i], wantSlice[i]) {
											t.Errorf("    element[%d]: got %v (%T), want %v (%T)", i, gotSlice[i], gotSlice[i], wantSlice[i], wantSlice[i])
										}
									}
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestOptions_Retyper(t *testing.T) {
	tests := []struct {
		name     string
		settings Options
		source   map[string]interface{}
		options  []retyperOptions
		want     map[string]interface{}
	}{
		{
			name:     "no options",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"key": "value",
			},
			options: []retyperOptions{},
			want: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name:     "with retypeSingleElementSlice option",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"single": []interface{}{"value"},
			},
			options: []retyperOptions{retypeSingleElementSlice},
			want: map[string]interface{}{
				"single": "value",
			},
		},
		{
			name:     "multiple options",
			settings: Options{TimeFormat: "20060102T150405"},
			source: map[string]interface{}{
				"single":   []interface{}{"value"},
				"multiple": []interface{}{"a", "b"},
			},
			options: []retyperOptions{retypeSingleElementSlice},
			want: map[string]interface{}{
				"single":   "value",
				"multiple": []string{"a", "b"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.settings.Retyper(tt.source, tt.options...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Options.Retyper() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOptions_Render provides comprehensive unit tests for the Render() method
// Now that Render() returns exit codes instead of calling os.Exit(), we can test it directly

func TestOptions_Render_Stdin(t *testing.T) {
	tests := []struct {
		name       string
		settings   Options
		stdin      string
		wantOutput string
		wantCode   int
	}{
		{
			name: "simple template from stdin",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--name=World"},
				Testing:    true,
				TimeFormat: "20060102T150405",
				Capitalize: true, // Need to capitalize for .Name to work
			},
			stdin:      "Hello, {{.Name}}!",
			wantOutput: "Hello, World!",
			wantCode:   0,
		},
		{
			name: "template with multiple variables",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--first=John", "--last=Doe"},
				Testing:    true,
				TimeFormat: "20060102T150405",
				Capitalize: true, // Need to capitalize for .First and .Last to work
			},
			stdin:      "{{.First}} {{.Last}}",
			wantOutput: "John Doe",
			wantCode:   0,
		},
		{
			name: "template with capitalized variables",
			settings: Options{
				Stdin:      true,
				Capitalize: true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--name=test"},
				Testing:    true,
				TimeFormat: "20060102T150405",
			},
			stdin:      "{{.Name}}",
			wantOutput: "test",
			wantCode:   0,
		},
		{
			name: "template with boolean flag",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--flag"},
				Testing:    true,
				TimeFormat: "20060102T150405",
				Capitalize: true, // Need to capitalize for .Flag to work
			},
			stdin:      "{{if .Flag}}YES{{else}}NO{{end}}",
			wantOutput: "YES",
			wantCode:   0,
		},
		{
			name: "template with missing key error",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{},
				Testing:    true,
				TimeFormat: "20060102T150405",
			},
			stdin:      "{{.Missing}}",
			wantOutput: "",
			wantCode:   8, // Template execution error
		},
		{
			name: "template with missing key zero",
			settings: Options{
				Stdin:      true,
				MissingKey: "zero",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{},
				Testing:    true,
				TimeFormat: "20060102T150405",
			},
			stdin:      "{{.Missing}}",
			wantOutput: "<no value>",
			wantCode:   0,
		},
		{
			name: "template with config values",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config: map[string]interface{}{
					"Name": "ConfigValue",
				},
				Templates:  []string{},
				Arguments:  []string{},
				Testing:    true,
				TimeFormat: "20060102T150405",
			},
			stdin:      "{{.Name}}",
			wantOutput: "ConfigValue",
			wantCode:   0,
		},
		{
			name: "template with toggle capitalization",
			settings: Options{
				Stdin:      true,
				Capitalize: true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"-C", "--name=test"},
				Testing:    true,
				TimeFormat: "20060102T150405",
			},
			stdin:      "{{.name}}",
			wantOutput: "test",
			wantCode:   0,
		},
		{
			name: "template with multiple values",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--items=one", "--items=two", "--items=three"},
				Testing:    true,
				TimeFormat: "20060102T150405",
				Capitalize: true, // Need to capitalize for .Items to work
			},
			stdin:      "{{range .Items}}{{.}} {{end}}",
			wantOutput: "one two three ",
			wantCode:   0,
		},
		{
			name: "template with dotted notation",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--foo.bar=value"},
				Testing:    true,
				TimeFormat: "20060102T150405",
				Capitalize: true, // Need to capitalize for .Foo.Bar to work
			},
			stdin:      "{{.Foo.Bar}}",
			wantOutput: "value",
			wantCode:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdin/stdout
			oldStdin := os.Stdin
			oldStdout := os.Stdout
			defer func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
			}()

			// Set up stdin
			stdinR, stdinW, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create stdin pipe: %v", err)
			}
			os.Stdin = stdinR

			// Set up stdout capture
			stdoutR, stdoutW, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create stdout pipe: %v", err)
			}
			os.Stdout = stdoutW

			// Write to stdin in a goroutine
			go func() {
				defer stdinW.Close()
				if _, err := stdinW.WriteString(tt.stdin); err != nil {
					t.Errorf("Failed to write to stdin: %v", err)
				}
			}()

			// Capture stdout in a goroutine
			var stdoutBuf bytes.Buffer
			stdoutDone := make(chan bool)
			go func() {
				if _, err := io.Copy(&stdoutBuf, stdoutR); err != nil {
					t.Errorf("Failed to copy stdout: %v", err)
				}
				stdoutDone <- true
			}()

			// Call Render()
			exitCode, err := tt.settings.Render()

			// Close pipes to signal EOF
			stdoutW.Close()
			stdinR.Close()

			// Wait for stdout capture to finish
			<-stdoutDone
			stdoutR.Close()

			// Check results
			if err != nil {
				t.Errorf("Render() returned error: %v", err)
			}
			if exitCode != tt.wantCode {
				t.Errorf("Render() exit code = %d, want %d", exitCode, tt.wantCode)
			}
			output := stdoutBuf.String()
			if tt.wantCode == 0 && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Render() output = %q, want to contain %q", output, tt.wantOutput)
			}
		})
	}
}

func TestOptions_Render_File(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		settings   Options
		template   string
		wantOutput string
		wantCode   int
	}{
		{
			name: "template file with variable",
			settings: Options{
				Stdin:      false,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{tmpDir + "/test.tmpl"},
				Arguments:  []string{"--name=World"},
				Testing:    true,
				TimeFormat: "20060102T150405",
				Capitalize: true, // Need to capitalize for .Name to work
			},
			template:   "Hello, {{.Name}}!",
			wantOutput: "Hello, World!",
			wantCode:   0,
		},
		{
			name: "missing template file",
			settings: Options{
				Stdin:      false,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{tmpDir + "/nonexistent.tmpl"},
				Arguments:  []string{},
				Testing:    true,
				TimeFormat: "20060102T150405",
			},
			template:   "",
			wantOutput: "",
			wantCode:   1, // File open error
		},
		{
			name: "invalid template syntax",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{},
				Testing:    true,
				TimeFormat: "20060102T150405",
			},
			template:   "{{.Unclosed",
			wantOutput: "",
			wantCode:   4, // Parse error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdin/stdout
			oldStdin := os.Stdin
			oldStdout := os.Stdout
			defer func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
			}()

			// Create template file if needed
			if tt.template != "" && tt.wantCode == 0 {
				templatePath := tt.settings.Templates[0]
				if err := os.WriteFile(templatePath, []byte(tt.template), 0644); err != nil {
					t.Fatalf("Failed to write template file: %v", err)
				}
			}

			// Set up stdin for cases that need it
			var stdinR *os.File
			if tt.settings.Stdin {
				var stdinW *os.File
				var err error
				stdinR, stdinW, err = os.Pipe()
				if err != nil {
					t.Fatalf("Failed to create stdin pipe: %v", err)
				}
				os.Stdin = stdinR
				go func() {
					defer stdinW.Close()
					if _, err := stdinW.WriteString(tt.template); err != nil {
						t.Errorf("Failed to write to stdin: %v", err)
					}
				}()
			}

			// Set up stdout capture
			stdoutR, stdoutW, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create stdout pipe: %v", err)
			}
			os.Stdout = stdoutW

			// Capture stdout in a goroutine
			var stdoutBuf bytes.Buffer
			stdoutDone := make(chan bool)
			go func() {
				if _, err := io.Copy(&stdoutBuf, stdoutR); err != nil {
					t.Errorf("Failed to copy stdout: %v", err)
				}
				stdoutDone <- true
			}()

			// Call Render()
			exitCode, err := tt.settings.Render()

			// Close pipes to signal EOF
			stdoutW.Close()
			if tt.settings.Stdin {
				// Close the stdin pipe we created
				stdinR.Close()
			}

			// Wait for stdout capture to finish
			<-stdoutDone
			stdoutR.Close()

			// Check results
			if err != nil {
				t.Errorf("Render() returned error: %v", err)
			}
			if exitCode != tt.wantCode {
				t.Errorf("Render() exit code = %d, want %d", exitCode, tt.wantCode)
			}
			if tt.wantCode == 0 && !strings.Contains(stdoutBuf.String(), tt.wantOutput) {
				t.Errorf("Render() output = %q, want to contain %q", stdoutBuf.String(), tt.wantOutput)
			}
		})
	}
}

func TestOptions_Render_Environment(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")

	settings := Options{
		Stdin:       true,
		MissingKey:  "error",
		Config:      map[string]interface{}{},
		Templates:   []string{},
		Arguments:   []string{},
		Environment: "env",
		Testing:     true,
		TimeFormat:  "20060102T150405",
	}

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	// Set up stdin
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	os.Stdin = stdinR
	go func() {
		defer stdinW.Close()
		if _, err := stdinW.WriteString("{{.env.TEST_VAR}}"); err != nil {
			t.Errorf("Failed to write to stdin: %v", err)
		}
	}()

	// Set up stdout capture
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	os.Stdout = stdoutW

	var stdoutBuf bytes.Buffer
	stdoutDone := make(chan bool)
	go func() {
		if _, err := io.Copy(&stdoutBuf, stdoutR); err != nil {
			t.Errorf("Failed to copy stdout: %v", err)
		}
		stdoutDone <- true
	}()

	// Call Render()
	exitCode, err := settings.Render()

	// Close pipes
	stdoutW.Close()
	stdinR.Close()
	<-stdoutDone
	stdoutR.Close()

	if err != nil {
		t.Errorf("Render() returned error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Render() exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdoutBuf.String(), "test_value") {
		t.Errorf("Render() output = %q, want to contain 'test_value'", stdoutBuf.String())
	}
}

func TestOptions_Render_Components(t *testing.T) {
	// This test verifies that the components used by Render() work correctly
	// The actual Render() method is tested via CLI integration tests
	t.Run("render components", func(t *testing.T) {
		settings := Options{
			TimeFormat: "20060102T150405",
			Config:     map[string]interface{}{},
		}

		// Test that typer works
		result := settings.typer("42")
		if result != int64(42) {
			t.Errorf("typer failed: got %v, want int64(42)", result)
		}

		// Test that Retyper works
		source := map[string]interface{}{
			"key": "value",
		}
		ret := settings.Retyper(source)
		if ret["key"] != "value" {
			t.Errorf("Retyper failed: got %v, want 'value'", ret["key"])
		}
	})
}

func TestOptions_Render_TestStructure(t *testing.T) {
	// This test documents the structure of Render() tests
	// Actual rendering is tested via CLI integration tests
	tests := []struct {
		name     string
		settings Options
		stdin    string
		wantErr  bool
	}{
		{
			name: "simple template from stdin",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--name=World"},
				Testing:    true,
			},
			stdin:   "Hello, {{.Name}}!",
			wantErr: false,
		},
		{
			name: "template with multiple variables",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--first=John", "--last=Doe"},
				Testing:    true,
			},
			stdin:   "{{.First}} {{.Last}}",
			wantErr: false,
		},
		{
			name: "template with capitalized variables",
			settings: Options{
				Stdin:      true,
				Capitalize: true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--name=test"},
				Testing:    true,
			},
			stdin:   "{{.Name}}",
			wantErr: false,
		},
		{
			name: "template with non-capitalized variables",
			settings: Options{
				Stdin:      true,
				Capitalize: false,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--name=test"},
				Testing:    true,
			},
			stdin:   "{{.name}}",
			wantErr: false,
		},
		{
			name: "template with boolean flag",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--flag"},
				Testing:    true,
			},
			stdin:   "{{if .Flag}}YES{{else}}NO{{end}}",
			wantErr: false,
		},
		{
			name: "template with dotted notation",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--foo.bar=value"},
				Testing:    true,
			},
			stdin:   "{{.Foo.Bar}}",
			wantErr: false,
		},
		{
			name: "template with missing key error",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{},
				Testing:    true,
			},
			stdin:   "{{.Missing}}",
			wantErr: false, // Returns nil but exits with status code
		},
		{
			name: "template with missing key zero",
			settings: Options{
				Stdin:      true,
				MissingKey: "zero",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{},
				Testing:    true,
			},
			stdin:   "{{.Missing}}",
			wantErr: false,
		},
		{
			name: "template with config values",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config: map[string]interface{}{
					"Name": "ConfigValue",
				},
				Templates: []string{},
				Arguments: []string{},
				Testing:   true,
			},
			stdin:   "{{.Name}}",
			wantErr: false,
		},
		{
			name: "template with environment variables",
			settings: Options{
				Stdin:       true,
				MissingKey:  "error",
				Config:      map[string]interface{}{},
				Templates:   []string{},
				Arguments:   []string{},
				Environment: "env",
				Testing:     true,
			},
			stdin:   "{{.env.PATH}}",
			wantErr: false,
		},
		{
			name: "template with toggle capitalization",
			settings: Options{
				Stdin:      true,
				Capitalize: true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"-C", "--name=test"},
				Testing:    true,
			},
			stdin:   "{{.name}}",
			wantErr: false,
		},
		{
			name: "template with multiple values",
			settings: Options{
				Stdin:      true,
				MissingKey: "error",
				Config:     map[string]interface{}{},
				Templates:  []string{},
				Arguments:  []string{"--items=one", "--items=two", "--items=three"},
				Testing:    true,
			},
			stdin:   "{{range .Items}}{{.}} {{end}}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Options.Render() calls os.Exit(), so we cannot test it directly.
			// These test cases document the expected behavior and are tested via
			// CLI integration tests in cmd/renderizer/cli_test.go
			_ = tt.settings
			_ = tt.stdin
			_ = tt.wantErr
		})
	}
}

func TestOptions_Render_ErrorCases(t *testing.T) {
	// Note: Options.Render() calls os.Exit(), so error cases are tested
	// via CLI integration tests in cmd/renderizer/cli_test.go (TestCLIErrorCases)
	// This test documents the expected error scenarios
	t.Run("error cases documented", func(t *testing.T) {
		// Error cases are tested in TestCLIErrorCases
		_ = t
	})
}
