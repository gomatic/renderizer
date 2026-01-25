package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// This test file follows Go 1.25+ testing standards:
// - Uses t.Setenv() for environment variable management (automatic cleanup)
// - Uses t.TempDir() for temporary directories (automatic cleanup)
// - Uses t.Cleanup() for resource cleanup (runs in reverse order)
// - Uses t.Helper() in helper functions to improve error reporting
// - Tests call functions directly instead of executing binaries

// runRenderizer executes the renderizer run function with given args and input
func runRenderizer(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()

	// Set up arguments with program name
	fullArgs := append([]string{"renderizer"}, args...)

	// Set up stdin
	var stdinReader *bytes.Reader
	if stdin != "" {
		stdinReader = bytes.NewReader([]byte(stdin))
	} else {
		stdinReader = bytes.NewReader([]byte{})
	}

	// Set up stdout/stderr capture using os.Pipe for proper redirection
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stdoutW.Close()

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}
	defer stderrR.Close()
	defer stderrW.Close()

	// Set testing environment variable
	t.Setenv("RENDERIZER_TESTING", "true")

	// Capture output in goroutines
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutDone := make(chan bool)
	stderrDone := make(chan bool)

	go func() {
		io.Copy(&stdoutBuf, stdoutR)
		stdoutDone <- true
	}()
	go func() {
		io.Copy(&stderrBuf, stderrR)
		stderrDone <- true
	}()

	// Call run() function directly
	result := run(fullArgs, stdinReader, stdoutW, stderrW)

	// Close writers to signal EOF - this will cause the readers to finish
	stdoutW.Close()
	stderrW.Close()

	// Wait for readers to finish copying all data
	<-stdoutDone
	<-stderrDone

	return stdoutBuf.String(), stderrBuf.String(), result.ExitCode
}

// TestCLIVersion tests the version command
func TestCLIVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"version flag", []string{"--version"}},
		{"short version flag", []string{"-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, exitCode := runRenderizer(t, "", tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
			}
			if !strings.Contains(stdout, "renderizer") {
				t.Errorf("Expected version output to contain 'renderizer', got: %s", stdout)
			}
		})
	}
}

// TestCLIHelp tests the help command
func TestCLIHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"help flag", []string{"--help"}},
		{"short help flag", []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, exitCode := runRenderizer(t, "", tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
			}
			if !strings.Contains(stdout, "USAGE:") {
				t.Errorf("Expected help output to contain 'USAGE:', got: %s", stdout)
			}
		})
	}
}

// TestCLIStdinBasic tests basic stdin template rendering
func TestCLIStdinBasic(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "simple variable",
			template: "Hello, {{.Name}}!",
			args:     []string{"--stdin", "--name=World"},
			expected: "Hello, World!",
		},
		{
			name:     "multiple variables",
			template: "{{.First}} {{.Last}}",
			args:     []string{"--stdin", "--first=John", "--last=Doe"},
			expected: "John Doe",
		},
		{
			name:     "integer value",
			template: "Count: {{.Count}}",
			args:     []string{"--stdin", "--count=42"},
			expected: "Count: 42",
		},
		{
			name:     "boolean value",
			template: "{{if .Flag}}YES{{else}}NO{{end}}",
			args:     []string{"--stdin", "--flag=true"},
			expected: "YES",
		},
		{
			name:     "boolean flag without value",
			template: "{{if .Flag}}YES{{else}}NO{{end}}",
			args:     []string{"--stdin", "--flag"},
			expected: "YES",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLIMultipleValues tests handling of multiple values for same variable
func TestCLIMultipleValues(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected []string
	}{
		{
			name:     "multiple items",
			template: "{{range .Items}}{{.}},{{end}}",
			args:     []string{"--stdin", "--items=one", "--items=two", "--items=three"},
			expected: []string{"one,", "two,", "three,"},
		},
		{
			name:     "items with index",
			template: "{{range $i, $v := .Items}}{{$i}}:{{$v}} {{end}}",
			args:     []string{"--stdin", "--items=a", "--items=b", "--items=c"},
			expected: []string{"0:a", "1:b", "2:c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			for _, exp := range tt.expected {
				if !strings.Contains(stdout, exp) {
					t.Errorf("Expected output to contain %q, got: %q", exp, stdout)
				}
			}
		})
	}
}

// TestCLICapitalization tests the -C flag for capitalization toggle
func TestCLICapitalization(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "default capitalization",
			template: "{{.Name}}",
			args:     []string{"--stdin", "--name=value"},
			expected: "value",
		},
		{
			name:     "disabled capitalization",
			template: "{{.name}}",
			args:     []string{"--stdin", "-C", "--name=value"},
			expected: "value",
		},
		{
			name:     "mixed capitalization",
			template: "{{.Name}} {{.foo}}",
			args:     []string{"--stdin", "--name=first", "-C", "--foo=second"},
			expected: "first second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLIEnvironment tests environment variable access
func TestCLIEnvironment(t *testing.T) {
	// Set test environment variable using Go 1.25+ t.Setenv() which automatically cleans up
	t.Setenv("TEST_VAR", "test_value")

	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "default env name",
			template: "{{.env.TEST_VAR}}",
			args:     []string{"--stdin"},
			expected: "test_value",
		},
		{
			name:     "custom env name",
			template: "{{.myenv.TEST_VAR}}",
			args:     []string{"--stdin", "--environment=myenv"},
			expected: "test_value",
		},
		{
			name:     "env alias",
			template: "{{.custom.TEST_VAR}}",
			args:     []string{"--stdin", "--env=custom"},
			expected: "test_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLITemplateFunctions tests template functions
func TestCLITemplateFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "lower function",
			template: "{{lower .Text}}",
			args:     []string{"--stdin", "--text=HELLO"},
			expected: "hello",
		},
		{
			name:     "upper function",
			template: "{{upper .Text}}",
			args:     []string{"--stdin", "--text=hello"},
			expected: "HELLO",
		},
		{
			name:     "add function",
			template: "{{add 5 10}}",
			args:     []string{"--stdin"},
			expected: "15",
		},
		{
			name:     "inc function",
			template: "{{inc 5}}",
			args:     []string{"--stdin"},
			expected: "6",
		},
		{
			name:     "replace function",
			template: "{{replace .Text \" \" \"-\" -1}}",
			args:     []string{"--stdin", "--text=hello world"},
			expected: "hello-world",
		},
		{
			name:     "trim function",
			template: "{{trim .Text \" \"}}",
			args:     []string{"--stdin", "--text=  hello  "},
			expected: "hello",
		},
		{
			name:     "cleanse function",
			template: "{{cleanse .Text}}",
			args:     []string{"--stdin", "--text=abc123def"},
			expected: "abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLIMissingKey tests the --missing flag
func TestCLIMissingKey(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		args        []string
		expectError bool
	}{
		{
			name:        "missing key with error (default)",
			template:    "{{.Missing}}",
			args:        []string{"--stdin"},
			expectError: true,
		},
		{
			name:        "missing key with zero",
			template:    "{{.Missing}}",
			args:        []string{"--stdin", "--missing=zero"},
			expectError: false,
		},
		{
			name:        "missing key with default",
			template:    "{{.Missing}}",
			args:        []string{"--stdin", "--missing=default"},
			expectError: false,
		},
		{
			name:        "missing key with invalid",
			template:    "{{.Missing}}",
			args:        []string{"--stdin", "--missing=invalid"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			hasError := exitCode != 0
			// Log output may appear in the actual stderr (not captured), but we check the exit code
			// For error cases, exitCode should be non-zero (typically 8 for template execution errors)
			if hasError != tt.expectError {
				t.Errorf("Test %q: Expected error=%v, got exitCode=%d. Stdout: %q, Stderr: %q",
					tt.name, tt.expectError, exitCode, stdout, stderr)
			}
			// Note: log.Print() writes to os.Stderr directly, so it appears in test output but is expected
		})
	}
}

// TestCLITemplateFile tests rendering from template files
func TestCLITemplateFile(t *testing.T) {
	// Create temporary directory and template file using Go 1.25+ t.TempDir()
	tmpDir := t.TempDir()

	templatePath := tmpDir + "/test.tmpl"
	templateContent := "Hello, {{.Name}}!"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "template file with variable",
			args:     []string{templatePath, "--name=World"},
			expected: "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, "", tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLISettingsFile tests loading settings from YAML files
func TestCLISettingsFile(t *testing.T) {
	// Create temporary directory and files using Go 1.25+ t.TempDir()
	tmpDir := t.TempDir()

	templatePath := tmpDir + "/test.tmpl"
	settingsPath := tmpDir + "/settings.yaml"

	templateContent := "Name: {{.Name}}, Items: {{range .Items}}{{.}},{{end}}"
	settingsContent := "Name: FromSettings\nItems:\n  - one\n  - two\n  - three"

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "load from settings file",
			args:     []string{templatePath, "--settings=" + settingsPath},
			expected: []string{"Name: FromSettings", "one,", "two,", "three,"},
		},
		{
			name:     "settings with override",
			args:     []string{templatePath, "--settings=" + settingsPath, "--name=Overridden"},
			expected: []string{"Name: Overridden", "one,", "two,", "three,"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, "", tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			for _, exp := range tt.expected {
				if !strings.Contains(stdout, exp) {
					t.Errorf("Expected output to contain %q, got: %q", exp, stdout)
				}
			}
		})
	}
}

// TestCLIMultipleTemplates tests rendering multiple template files
func TestCLIMultipleTemplates(t *testing.T) {
	// Create temporary directory and template files using Go 1.25+ t.TempDir()
	tmpDir := t.TempDir()

	template1Path := tmpDir + "/test1.tmpl"
	template2Path := tmpDir + "/test2.tmpl"

	if err := os.WriteFile(template1Path, []byte("First: {{.Name}}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(template2Path, []byte("Second: {{.Name}}"), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, exitCode := runRenderizer(t, "", template1Path, template2Path, "--name=Test")
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "First: Test") {
		t.Errorf("Expected output to contain 'First: Test', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Second: Test") {
		t.Errorf("Expected output to contain 'Second: Test', got: %q", stdout)
	}
}

// TestCLIDottedNotation tests dotted notation for nested variables
func TestCLIDottedNotation(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "single level dot notation",
			template: "{{.Foo.Bar}}",
			args:     []string{"--stdin", "--foo.bar=value"},
			expected: "value",
		},
		{
			name:     "multiple level dot notation",
			template: "{{.A.B.C}}",
			args:     []string{"--stdin", "--a.b.c=nested"},
			expected: "nested",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLIDefaultTemplateDiscovery tests default template file discovery
func TestCLIDefaultTemplateDiscovery(t *testing.T) {
	// Create temporary directory and navigate to it using Go 1.25+ t.TempDir()
	tmpDir := t.TempDir()

	// Save current directory and use Go 1.25+ t.Cleanup() for restoration
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
	})

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test various default file names
	testFiles := []string{
		"renderizer.yaml.tmpl",
		"renderizer.json.tmpl",
		"renderizer.txt.tmpl",
	}

	for _, filename := range testFiles {
		t.Run("discover_"+filename, func(t *testing.T) {
			// Create the template file
			content := "Test: {{.Value}}"
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}

			stdout, stderr, exitCode := runRenderizer(t, "", "--value=success")

			// Clean up immediately to prevent interference with subsequent tests
			os.Remove(filename)

			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, "Test: success") {
				t.Errorf("Expected output to contain 'Test: success', got: %q", stdout)
			}
		})
	}
}

// TestCLIVerboseMode tests verbose output mode
func TestCLIVerboseMode(t *testing.T) {
	// Create temporary template file using Go 1.25+ t.TempDir()
	tmpDir := t.TempDir()

	templatePath := tmpDir + "/test.tmpl"
	if err := os.WriteFile(templatePath, []byte("{{.Name}}"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "verbose long flag",
			args: []string{templatePath, "--name=Test", "--verbose"},
		},
		{
			name: "verbose short flag",
			args: []string{templatePath, "--name=Test", "-V"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verbose mode produces output on stderr
			_, stderr, exitCode := runRenderizer(t, "", tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
			}
			// Just verify it doesn't crash with verbose mode
			// The actual verbose output goes to logs which we can't easily capture
			_ = stderr
		})
	}
}

// TestCLISettingsEnvironmentVariable tests RENDERIZER environment variable
func TestCLISettingsEnvironmentVariable(t *testing.T) {
	// Create temporary directory and files using Go 1.25+ t.TempDir()
	tmpDir := t.TempDir()

	templatePath := tmpDir + "/test.tmpl"
	settingsPath := tmpDir + "/settings.yaml"

	if err := os.WriteFile(templatePath, []byte("{{.Name}}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("Name: EnvSettings"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("RENDERIZER", settingsPath)
	t.Setenv("RENDERIZER_TESTING", "true")

	var stdout, stderr bytes.Buffer
	result := run([]string{"renderizer", templatePath}, nil, &stdout, &stderr)

	if result.ExitCode != 0 {
		t.Fatalf("Command failed with exit code %d: %v", result.ExitCode, result.Error)
	}

	if !strings.Contains(stdout.String(), "EnvSettings") {
		t.Errorf("Expected output to contain 'EnvSettings', got: %q", stdout.String())
	}
}

// TestCLIStdinFlag tests explicit --stdin flag
func TestCLIStdinFlag(t *testing.T) {
	tests := []struct {
		name     string
		stdin    string
		args     []string
		expected string
	}{
		{
			name:     "explicit stdin with -c",
			stdin:    "Hello, {{.Name}}!",
			args:     []string{"-c", "--name=World"},
			expected: "Hello, World!",
		},
		{
			name:     "explicit stdin with --stdin",
			stdin:    "Value: {{.Value}}",
			args:     []string{"--stdin", "--value=42"},
			expected: "Value: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.stdin, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLITypedValues tests that values are properly typed
func TestCLITypedValues(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "integer type",
			template: "{{if eq .Count 42}}correct{{else}}wrong{{end}}",
			args:     []string{"--stdin", "--count=42"},
			expected: "correct",
		},
		{
			name:     "float type",
			template: "{{if eq .Value 3.14}}correct{{else}}wrong{{end}}",
			args:     []string{"--stdin", "--value=3.14"},
			expected: "correct",
		},
		{
			name:     "boolean type",
			template: "{{if .Flag}}correct{{else}}wrong{{end}}",
			args:     []string{"--stdin", "--flag=true"},
			expected: "correct",
		},
		{
			name:     "string type",
			template: "{{if eq .Text \"hello\"}}correct{{else}}wrong{{end}}",
			args:     []string{"--stdin", "--text=hello"},
			expected: "correct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLIMultipleSettingsFiles tests loading from multiple settings files
func TestCLIMultipleSettingsFiles(t *testing.T) {
	tmpDir := t.TempDir()

	templatePath := tmpDir + "/test.tmpl"
	settings1Path := tmpDir + "/settings1.yaml"
	settings2Path := tmpDir + "/settings2.yaml"

	templateContent := "First: {{.First}}, Second: {{.Second}}"
	settings1Content := "First: FromFirst"
	settings2Content := "Second: FromSecond"

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings1Path, []byte(settings1Content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings2Path, []byte(settings2Content), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, exitCode := runRenderizer(t, "",
		templatePath,
		"--settings="+settings1Path,
		"--settings="+settings2Path)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "First: FromFirst") {
		t.Errorf("Expected output to contain 'First: FromFirst', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Second: FromSecond") {
		t.Errorf("Expected output to contain 'Second: FromSecond', got: %q", stdout)
	}
}

// TestCLIShortFlags tests short flag aliases
func TestCLIShortFlags(t *testing.T) {
	tmpDir := t.TempDir()

	templatePath := tmpDir + "/test.tmpl"
	settingsPath := tmpDir + "/settings.yaml"

	if err := os.WriteFile(templatePath, []byte("{{.Name}}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("Name: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "short settings flag -S with =",
			args:     []string{templatePath, "-S=" + settingsPath},
			expected: "Test",
		},
		{
			name:     "short environment flag -e with =",
			args:     []string{"--stdin", "-e=myenv", "--name=Value"},
			expected: "Value",
		},
		{
			name:     "short missing flag -m with =",
			args:     []string{"--stdin", "-m=zero", "--name=Value"},
			expected: "Value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var template string
			if strings.Contains(tt.args[0], "stdin") {
				template = "{{.Name}}"
			}
			stdout, stderr, exitCode := runRenderizer(t, template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestCLIErrorCases tests various error scenarios
func TestCLIErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		args        []string
		expectError bool
	}{
		{
			name:        "missing template file",
			args:        []string{"nonexistent.tmpl"},
			expectError: true,
		},
		{
			name:        "template parse error",
			template:    "{{.Unclosed",
			args:        []string{"--stdin"},
			expectError: true,
		},
		{
			name:        "template execution error with missing key",
			template:    "{{.Missing}}",
			args:        []string{"--stdin", "--missing=error"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, exitCode := runRenderizer(t, tt.template, tt.args...)
			hasError := exitCode != 0
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got exitCode=%d", tt.expectError, exitCode)
			}
		})
	}
}

// TestCLIDefaultSettingsFileOptional tests that default settings files are optional
func TestCLIDefaultSettingsFileOptional(t *testing.T) {
	tmpDir := t.TempDir()

	// Save current directory and use Go 1.25+ t.Cleanup() for restoration
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
	})

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	templatePath := tmpDir + "/test.tmpl"
	if err := os.WriteFile(templatePath, []byte("{{.Name}}"), 0644); err != nil {
		t.Fatal(err)
	}

	// When default settings file doesn't exist, should not error
	stdout, _, exitCode := runRenderizer(t, "", templatePath, "--name=Value")
	if exitCode != 0 {
		t.Error("Expected zero exit code when default settings file doesn't exist")
	}
	if !strings.Contains(stdout, "Value") {
		t.Errorf("Expected output to contain 'Value', got: %q", stdout)
	}
}

// TestCLICommandLineFunction tests the command_line template function
func TestCLICommandLineFunction(t *testing.T) {
	// In testing mode, command_line returns "testing"
	template := "{{command_line}}"
	stdout, stderr, exitCode := runRenderizer(t, template, "--stdin")
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "testing") {
		t.Errorf("Expected output to contain 'testing', got: %q", stdout)
	}
}

// TestCLIRange tests using range with multiple values
func TestCLIRange(t *testing.T) {
	template := "{{range .Items}}{{.}}\n{{end}}"
	stdout, stderr, exitCode := runRenderizer(t, template, "--stdin", "--items=a", "--items=b", "--items=c")
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "a") || !strings.Contains(stdout, "b") || !strings.Contains(stdout, "c") {
		t.Errorf("Expected output to contain all items, got: %q", stdout)
	}
}

// TestCLINestedVariables tests nested variable access with dotted notation
func TestCLINestedVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "two level nesting",
			template: "{{.Level1.Level2}}",
			args:     []string{"--stdin", "--level1.level2=value"},
			expected: "value",
		},
		{
			name:     "three level nesting",
			template: "{{.A.B.C}}",
			args:     []string{"--stdin", "--a.b.c=deep"},
			expected: "deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRenderizer(t, tt.template, tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}
			if !strings.Contains(stdout, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, stdout)
			}
		})
	}
}
