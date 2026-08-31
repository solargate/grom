package equipment

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
)

func TestEquipmentTypesMatchFlutterCatalog(t *testing.T) {
	goTypes := []string{TypeBike, TypeShoes, TypeWater, TypeOther}
	dartPath := filepath.Join("..", "..", "ui", "grom", "lib", "models", "equipment_types.dart")
	data, err := os.ReadFile(dartPath)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`case EquipmentType\.\w+:\s*\n\s*return '([^']+)'`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	flutterIDs := make([]string, 0, len(matches))
	for _, m := range matches {
		flutterIDs = append(flutterIDs, m[1])
	}
	slices.Sort(flutterIDs)
	goSorted := append([]string(nil), goTypes...)
	slices.Sort(goSorted)
	if !slices.Equal(goSorted, flutterIDs) {
		t.Fatalf("Go equipment types %#v\nFlutter equipment types %#v", goSorted, flutterIDs)
	}
}
