package workouts

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

func SanitizeDownloadBasename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "workout"
	}

	var b strings.Builder
	for _, r := range name {
		switch {
		case r == '/' || r == '\\' || r == 0:
			b.WriteRune('_')
		case unicode.IsControl(r):
			continue
		default:
			b.WriteRune(r)
		}
	}

	result := strings.TrimSpace(b.String())
	if result == "" {
		return "workout"
	}
	return result
}

func TrackDownloadFilename(workoutName, ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "bin"
	}
	return SanitizeDownloadBasename(workoutName) + "." + ext
}

func ContentDispositionAttachment(filename string) string {
	return fmt.Sprintf(
		`attachment; filename="%s"; filename*=UTF-8''%s`,
		toASCIIFilename(filename),
		url.PathEscape(filename),
	)
}

func toASCIIFilename(filename string) string {
	var b strings.Builder
	for _, r := range filename {
		if r < 128 && r != '"' && r != '\\' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	result := b.String()
	if result == "" {
		return "download"
	}
	return result
}
