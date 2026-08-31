package workouts

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestSportTypesMatchFlutterCatalog(t *testing.T) {
	goTypes := ValidSportTypes()
	dartPath := filepath.Join("..", "..", "ui", "grom", "lib", "models", "sport_types.dart")
	data, err := os.ReadFile(dartPath)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`SportTypeInfo\(id: '([^']+)'`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		t.Fatal("no sport types found in Flutter catalog")
	}
	flutterIDs := make([]string, 0, len(matches))
	seen := make(map[string]struct{})
	for _, m := range matches {
		id := m[1]
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		flutterIDs = append(flutterIDs, id)
	}
	slices.Sort(flutterIDs)
	goSorted := append([]string(nil), goTypes...)
	slices.Sort(goSorted)
	if !slices.Equal(goSorted, flutterIDs) {
		t.Fatalf("Go sport types %#v\nFlutter sport types %#v", goSorted, flutterIDs)
	}
}

func TestSportTypesCatalogNonEmpty(t *testing.T) {
	types := ValidSportTypes()
	if len(types) < 10 {
		t.Fatalf("expected substantial catalog, got %d", len(types))
	}
	for _, id := range types {
		if strings.TrimSpace(id) == "" {
			t.Fatal("empty sport type id")
		}
	}
}
