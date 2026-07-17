package federation

import "strings"

// OwnerKeyFromHandle converts a federation handle to a safe directory/key segment.
func OwnerKeyFromHandle(handle string) string {
	return strings.NewReplacer("@", "_", ":", "_", "/", "_").Replace(handle)
}
