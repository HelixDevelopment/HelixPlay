package theming

import (
	"strings"
	"testing"
)

func TestCSSGenerator(t *testing.T) {
	ts := DefaultTokens()
	g := NewCSSGenerator("")
	css := g.Generate(ts)

	if !strings.Contains(css, ":root {") {
		t.Error("expected :root block")
	}
	if !strings.Contains(css, "--color-primary:") {
		t.Error("expected --color-primary property")
	}
	if !strings.Contains(css, ts.Colors.Primary) {
		t.Errorf("expected primary color %s in CSS", ts.Colors.Primary)
	}
	if !strings.Contains(css, "--font-family:") {
		t.Error("expected --font-family property")
	}
}

func TestCSSGeneratorWithPrefix(t *testing.T) {
	ts := DefaultTokens()
	g := NewCSSGenerator("helix")
	css := g.Generate(ts)

	if !strings.Contains(css, "--helix-color-primary:") {
		t.Error("expected prefixed property")
	}
}

func TestCSSGeneratorContainsAllProperties(t *testing.T) {
	ts := DefaultTokens()
	g := NewCSSGenerator("")
	css := g.Generate(ts)

	expected := []string{
		"--color-background",
		"--color-surface",
		"--color-text-primary",
		"--spacing-unit",
		"--border-radius",
		"--shadow-small",
		"--bp-desktop",
	}

	for _, prop := range expected {
		if !strings.Contains(css, prop) {
			t.Errorf("expected %s in CSS", prop)
		}
	}
}
