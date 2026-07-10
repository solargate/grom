package federation

import "testing"

func TestExtractIconURLString(t *testing.T) {
	url := ExtractIconURL(map[string]any{
		"icon": "https://example.com/a.webp",
	})
	if url != "https://example.com/a.webp" {
		t.Fatalf("unexpected url: %q", url)
	}
}

func TestExtractIconURLImageObject(t *testing.T) {
	url := ExtractIconURL(map[string]any{
		"icon": map[string]any{
			"type":      "Image",
			"mediaType": "image/webp",
			"url":       "https://example.com/a.webp",
		},
	})
	if url != "https://example.com/a.webp" {
		t.Fatalf("unexpected url: %q", url)
	}
}
