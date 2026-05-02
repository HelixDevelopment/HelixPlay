package theming

import (
	"fmt"
	"strings"
)

// CSSGenerator produces CSS custom properties from a token set.
type CSSGenerator struct {
	prefix string
}

// NewCSSGenerator creates a generator with an optional variable prefix.
func NewCSSGenerator(prefix string) *CSSGenerator {
	if prefix != "" && !strings.HasSuffix(prefix, "-") {
		prefix = prefix + "-"
	}
	return &CSSGenerator{prefix: prefix}
}

// Generate returns a CSS string with custom properties.
func (g *CSSGenerator) Generate(ts TokenSet) string {
	var b strings.Builder
	b.WriteString(":root {\n")

	// Colors
	b.WriteString(g.prop("color-primary", ts.Colors.Primary))
	b.WriteString(g.prop("color-primary-hover", ts.Colors.PrimaryHover))
	b.WriteString(g.prop("color-secondary", ts.Colors.Secondary))
	b.WriteString(g.prop("color-secondary-hover", ts.Colors.SecondaryHover))
	b.WriteString(g.prop("color-background", ts.Colors.Background))
	b.WriteString(g.prop("color-surface", ts.Colors.Surface))
	b.WriteString(g.prop("color-text-primary", ts.Colors.TextPrimary))
	b.WriteString(g.prop("color-text-secondary", ts.Colors.TextSecondary))
	b.WriteString(g.prop("color-success", ts.Colors.Success))
	b.WriteString(g.prop("color-warning", ts.Colors.Warning))
	b.WriteString(g.prop("color-error", ts.Colors.Error))

	// Typography
	b.WriteString(g.prop("font-family", ts.Typography.FontFamily))
	b.WriteString(g.prop("font-base-size", fmt.Sprintf("%dpx", ts.Typography.BaseSize)))

	// Spacing
	b.WriteString(g.prop("spacing-unit", fmt.Sprintf("%dpx", ts.Spacing.BaseUnit)))

	// Border
	b.WriteString(g.prop("border-radius", fmt.Sprintf("%dpx", ts.Border.Radius)))
	b.WriteString(g.prop("border-width", fmt.Sprintf("%dpx", ts.Border.Width)))
	b.WriteString(g.prop("border-color", ts.Border.Color))

	// Shadows
	b.WriteString(g.prop("shadow-small", ts.Shadow.Small))
	b.WriteString(g.prop("shadow-medium", ts.Shadow.Medium))
	b.WriteString(g.prop("shadow-large", ts.Shadow.Large))

	// Breakpoints
	b.WriteString(g.prop("bp-mobile", fmt.Sprintf("%dpx", ts.Breakpoints.Mobile)))
	b.WriteString(g.prop("bp-tablet", fmt.Sprintf("%dpx", ts.Breakpoints.Tablet)))
	b.WriteString(g.prop("bp-desktop", fmt.Sprintf("%dpx", ts.Breakpoints.Desktop)))
	b.WriteString(g.prop("bp-wide", fmt.Sprintf("%dpx", ts.Breakpoints.Wide)))

	b.WriteString("}\n")
	return b.String()
}

func (g *CSSGenerator) prop(name, value string) string {
	return fmt.Sprintf("  --%s%s: %s;\n", g.prefix, name, value)
}
