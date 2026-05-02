package theming

import (
	"testing"
)

func TestDefaultTokens(t *testing.T) {
	ts := DefaultTokens()
	if ts.Name == "" {
		t.Fatal("expected non-empty theme name")
	}
	if ts.Colors.Primary == "" {
		t.Fatal("expected primary color")
	}
	if ts.Typography.BaseSize <= 0 {
		t.Fatal("expected positive base size")
	}
}

func TestTokenSetValidate(t *testing.T) {
	ts := DefaultTokens()
	if err := ts.Validate(); err != nil {
		t.Fatalf("expected valid: %v", err)
	}

	ts.Name = ""
	if err := ts.Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}

	ts = DefaultTokens()
	ts.Typography.BaseSize = 50
	if err := ts.Validate(); err == nil {
		t.Fatal("expected error for base size > 32")
	}

	ts = DefaultTokens()
	ts.Spacing.BaseUnit = 0
	if err := ts.Validate(); err == nil {
		t.Fatal("expected error for base unit < 1")
	}
}

func TestTokenSetJSON(t *testing.T) {
	ts := DefaultTokens()
	data, err := ts.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}

	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}
	if parsed.Name != ts.Name {
		t.Errorf("expected %s, got %s", ts.Name, parsed.Name)
	}
	if parsed.Colors.Primary != ts.Colors.Primary {
		t.Errorf("expected %s, got %s", ts.Colors.Primary, parsed.Colors.Primary)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	_, err := FromJSON([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
