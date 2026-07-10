package federation

import "strings"

func ExtractIconURL(actor map[string]any) string {
	icon, ok := actor["icon"]
	if !ok || icon == nil {
		return ""
	}
	switch v := icon.(type) {
	case string:
		return strings.TrimSpace(v)
	case map[string]any:
		return extractImageURL(v)
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				if url := extractImageURL(m); url != "" {
					return url
				}
			}
		}
	}
	return ""
}

func extractImageURL(obj map[string]any) string {
	if url, ok := obj["url"].(string); ok && url != "" {
		return strings.TrimSpace(url)
	}
	if href, ok := obj["href"].(string); ok && href != "" {
		return strings.TrimSpace(href)
	}
	switch urlVal := obj["url"].(type) {
	case map[string]any:
		if href, ok := urlVal["href"].(string); ok {
			return strings.TrimSpace(href)
		}
	case []any:
		for _, item := range urlVal {
			switch entry := item.(type) {
			case string:
				if entry != "" {
					return strings.TrimSpace(entry)
				}
			case map[string]any:
				if href, ok := entry["href"].(string); ok && href != "" {
					return strings.TrimSpace(href)
				}
				if url, ok := entry["url"].(string); ok && url != "" {
					return strings.TrimSpace(url)
				}
			}
		}
	}
	return ""
}

func ExtractActorName(actor map[string]any) string {
	name, _ := actor["name"].(string)
	return strings.TrimSpace(name)
}
