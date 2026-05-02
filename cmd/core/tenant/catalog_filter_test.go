package tenant

import (
	"testing"
)

func TestCatalogFilterVisible(t *testing.T) {
	f := DefaultCatalogFilter()
	if !f.IsGameVisible("game-1", []string{"action"}, 8.5, "T", "steam") {
		t.Error("expected game to be visible with default filter")
	}
}

func TestCatalogFilterBlocked(t *testing.T) {
	f := CatalogFilter{
		BlockedGames: []string{"game-1"},
	}
	if f.IsGameVisible("game-1", []string{"action"}, 9.0, "E", "steam") {
		t.Error("expected blocked game to be hidden")
	}
}

func TestCatalogFilterAllowedStores(t *testing.T) {
	f := CatalogFilter{
		AllowedStores: []string{"steam", "epic"},
	}
	if !f.IsGameVisible("g1", []string{}, 5.0, "E", "steam") {
		t.Error("expected steam game to be visible")
	}
	if f.IsGameVisible("g1", []string{}, 5.0, "E", "gog") {
		t.Error("expected gog game to be hidden")
	}
}

func TestCatalogFilterMinRating(t *testing.T) {
	f := CatalogFilter{
		MinRating: 7.0,
	}
	if f.IsGameVisible("g1", []string{}, 6.0, "E", "steam") {
		t.Error("expected low-rated game to be hidden")
	}
	if !f.IsGameVisible("g1", []string{}, 8.0, "E", "steam") {
		t.Error("expected high-rated game to be visible")
	}
}

func TestCatalogFilterIncludeTags(t *testing.T) {
	f := CatalogFilter{
		IncludeTags: []string{"rpg"},
	}
	if !f.IsGameVisible("g1", []string{"rpg", "action"}, 5.0, "E", "steam") {
		t.Error("expected game with included tag to be visible")
	}
	if f.IsGameVisible("g1", []string{"shooter"}, 5.0, "E", "steam") {
		t.Error("expected game without included tag to be hidden")
	}
}

func TestCatalogFilterExcludeTags(t *testing.T) {
	f := CatalogFilter{
		ExcludeTags: []string{"horror"},
	}
	if f.IsGameVisible("g1", []string{"horror"}, 5.0, "E", "steam") {
		t.Error("expected game with excluded tag to be hidden")
	}
	if !f.IsGameVisible("g1", []string{"action"}, 5.0, "E", "steam") {
		t.Error("expected game without excluded tag to be visible")
	}
}

func TestCatalogFilterAgeRating(t *testing.T) {
	f := CatalogFilter{
		MaxAgeRating: "T",
	}
	if !f.IsGameVisible("g1", []string{}, 5.0, "E", "steam") {
		t.Error("expected E-rated game to be visible")
	}
	if !f.IsGameVisible("g1", []string{}, 5.0, "T", "steam") {
		t.Error("expected T-rated game to be visible")
	}
	if f.IsGameVisible("g1", []string{}, 5.0, "M", "steam") {
		t.Error("expected M-rated game to be hidden")
	}
}

func TestCatalogFilterValidate(t *testing.T) {
	f := CatalogFilter{MinRating: 5.0}
	if err := f.Validate(); err != nil {
		t.Fatalf("expected valid: %v", err)
	}

	f.MinRating = 11.0
	if err := f.Validate(); err == nil {
		t.Fatal("expected error for min_rating > 10")
	}
}
