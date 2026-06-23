package renderizer_test

import (
	"strings"
	"testing"

	"github.com/gomatic/renderizer"
)

func TestFuncs(t *testing.T) {
	if _, ok := renderizer.Funcs()["upper"]; !ok {
		t.Fatal("Funcs missing expected function 'upper'")
	}
}

func TestRender(t *testing.T) {
	out, err := renderizer.Render(renderizer.Funcs(), "error", "t", []byte(`Hello {{ .Name | upper }}`), map[string]any{"Name": "world"})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if string(out) != "Hello WORLD" {
		t.Fatalf("Render = %q", out)
	}
}

func TestRenderError(t *testing.T) {
	if _, err := renderizer.Render(renderizer.Funcs(), "error", "t", []byte(`{{ .Missing }}`), map[string]any{}); err == nil {
		t.Fatal("Render expected error for missing key")
	}
}

func TestAnalyze(t *testing.T) {
	out, err := renderizer.Analyze(renderizer.Funcs(), "t", []byte(`{{ .Name }} {{ range .Items }}{{ . }}{{ end }}`))
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if !strings.Contains(string(out), "Name") || !strings.Contains(string(out), "Items") {
		t.Fatalf("Analyze skeleton = %q", out)
	}
}

func TestAnalyzeError(t *testing.T) {
	if _, err := renderizer.Analyze(renderizer.Funcs(), "t", []byte(`{{ .Unterminated`)); err == nil {
		t.Fatal("Analyze expected parse error")
	}
}
