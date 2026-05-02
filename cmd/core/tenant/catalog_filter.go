package tenant

import (
	"fmt"
	"strings"
)

// CatalogFilter defines which games are visible to a tenant's users.
type CatalogFilter struct {
	IncludeTags    []string
	ExcludeTags    []string
	MinRating      float64
	MaxAgeRating   string
	AllowedStores  []string
	BlockedGames   []string
}

// DefaultCatalogFilter allows all games.
func DefaultCatalogFilter() CatalogFilter {
	return CatalogFilter{
		IncludeTags:   []string{},
		ExcludeTags:   []string{},
		MinRating:     0,
		MaxAgeRating:  "",
		AllowedStores: []string{},
		BlockedGames:  []string{},
	}
}

// IsGameVisible returns true if the game passes all filter criteria.
func (f *CatalogFilter) IsGameVisible(gameID string, tags []string, rating float64, ageRating, store string) bool {
	// Blocked games always hidden
	for _, blocked := range f.BlockedGames {
		if strings.EqualFold(blocked, gameID) {
			return false
		}
	}

	// Allowed stores restriction
	if len(f.AllowedStores) > 0 {
		found := false
		for _, s := range f.AllowedStores {
			if strings.EqualFold(s, store) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Minimum rating
	if rating < f.MinRating {
		return false
	}

	// Max age rating (simple string comparison)
	if f.MaxAgeRating != "" && ageRating != "" {
		if compareAgeRating(ageRating, f.MaxAgeRating) > 0 {
			return false
		}
	}

	// Include tags (must have at least one if specified)
	if len(f.IncludeTags) > 0 {
		hasTag := false
		for _, it := range f.IncludeTags {
			for _, gt := range tags {
				if strings.EqualFold(it, gt) {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	// Exclude tags (must not have any)
	for _, et := range f.ExcludeTags {
		for _, gt := range tags {
			if strings.EqualFold(et, gt) {
				return false
			}
		}
	}

	return true
}

// Validate checks the filter configuration for consistency.
func (f *CatalogFilter) Validate() error {
	if f.MinRating < 0 || f.MinRating > 10 {
		return fmt.Errorf("min_rating must be between 0 and 10")
	}
	return nil
}

// ageRatingOrder defines a simple ordering for common age ratings.
var ageRatingOrder = map[string]int{
	"E":    1,  // Everyone
	"E10+": 2,  // Everyone 10+
	"T":    3,  // Teen
	"M":    4,  // Mature
	"AO":   5,  // Adults Only
}

func compareAgeRating(a, b string) int {
	ao := ageRatingOrder[strings.ToUpper(a)]
	bo := ageRatingOrder[strings.ToUpper(b)]
	if ao == 0 {
		ao = 99
	}
	if bo == 0 {
		bo = 99
	}
	return ao - bo
}
