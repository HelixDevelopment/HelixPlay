// Package theming provides design token schemas and validation for white-label theming.
package theming

import (
	"encoding/json"
	"fmt"
)

// TokenSet holds a complete theme definition.
type TokenSet struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Colors      ColorTokens       `json:"colors"`
	Typography  TypographyTokens  `json:"typography"`
	Spacing     SpacingTokens     `json:"spacing"`
	Border      BorderTokens      `json:"border"`
	Shadow      ShadowTokens      `json:"shadow"`
	Breakpoints BreakpointTokens  `json:"breakpoints"`
}

// ColorTokens defines the color palette.
type ColorTokens struct {
	Primary       string `json:"primary"`
	PrimaryHover  string `json:"primary_hover"`
	Secondary     string `json:"secondary"`
	SecondaryHover string `json:"secondary_hover"`
	Background    string `json:"background"`
	Surface       string `json:"surface"`
	TextPrimary   string `json:"text_primary"`
	TextSecondary string `json:"text_secondary"`
	Success       string `json:"success"`
	Warning       string `json:"warning"`
	Error         string `json:"error"`
}

// TypographyTokens defines font settings.
type TypographyTokens struct {
	FontFamily string `json:"font_family"`
	BaseSize   int    `json:"base_size_px"`
	ScaleRatio string `json:"scale_ratio"`
}

// SpacingTokens defines spacing scale.
type SpacingTokens struct {
	BaseUnit int `json:"base_unit_px"`
}

// BorderTokens defines border settings.
type BorderTokens struct {
	Radius int    `json:"radius_px"`
	Width  int    `json:"width_px"`
	Color  string `json:"color"`
}

// ShadowTokens defines shadow presets.
type ShadowTokens struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

// BreakpointTokens defines responsive breakpoints.
type BreakpointTokens struct {
	Mobile  int `json:"mobile_px"`
	Tablet  int `json:"tablet_px"`
	Desktop int `json:"desktop_px"`
	Wide    int `json:"wide_px"`
}

// DefaultTokens returns the HelixPlay default theme.
func DefaultTokens() TokenSet {
	return TokenSet{
		Name:    "HelixPlay Default",
		Version: "1.0.0",
		Colors: ColorTokens{
			Primary:        "#6366f1",
			PrimaryHover:   "#4f46e5",
			Secondary:      "#ec4899",
			SecondaryHover: "#db2777",
			Background:     "#0f0f23",
			Surface:        "#1a1a2e",
			TextPrimary:    "#f8fafc",
			TextSecondary:  "#94a3b8",
			Success:        "#22c55e",
			Warning:        "#f59e0b",
			Error:          "#ef4444",
		},
		Typography: TypographyTokens{
			FontFamily: "Inter, system-ui, sans-serif",
			BaseSize:   16,
			ScaleRatio: "1.25",
		},
		Spacing: SpacingTokens{BaseUnit: 4},
		Border: BorderTokens{
			Radius: 8,
			Width:  1,
			Color:  "rgba(255,255,255,0.1)",
		},
		Shadow: ShadowTokens{
			Small:  "0 1px 2px rgba(0,0,0,0.3)",
			Medium: "0 4px 6px rgba(0,0,0,0.4)",
			Large:  "0 10px 15px rgba(0,0,0,0.5)",
		},
		Breakpoints: BreakpointTokens{
			Mobile:  480,
			Tablet:  768,
			Desktop: 1024,
			Wide:    1440,
		},
	}
}

// Validate checks the token set for required fields and valid values.
func (ts *TokenSet) Validate() error {
	if ts.Name == "" {
		return fmt.Errorf("theme name is required")
	}
	if ts.Colors.Primary == "" {
		return fmt.Errorf("primary color is required")
	}
	if ts.Colors.Background == "" {
		return fmt.Errorf("background color is required")
	}
	if ts.Typography.BaseSize < 8 || ts.Typography.BaseSize > 32 {
		return fmt.Errorf("base_size_px must be between 8 and 32")
	}
	if ts.Spacing.BaseUnit < 1 || ts.Spacing.BaseUnit > 16 {
		return fmt.Errorf("base_unit_px must be between 1 and 16")
	}
	return nil
}

// ToJSON serializes the token set to JSON.
func (ts *TokenSet) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ts, "", "  ")
}

// FromJSON parses a token set from JSON.
func FromJSON(data []byte) (*TokenSet, error) {
	var ts TokenSet
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("invalid token JSON: %w", err)
	}
	return &ts, nil
}
