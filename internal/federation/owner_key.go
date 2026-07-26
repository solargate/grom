package federation

import "strings"

// OwnerKeyFromHandle converts a federation handle to a safe directory/key segment.
func OwnerKeyFromHandle(handle string) string {
	return strings.NewReplacer("@", "_", ":", "_", "/", "_").Replace(handle)
}

// OwnerHandleFromKey reverses OwnerKeyFromHandle for typical @ handles.
func OwnerHandleFromKey(ownerKey string) string {
	return strings.ReplaceAll(ownerKey, "_", "@")
}

// OwnerNicknameFromKey extracts the local part of a handle-derived owner key.
func OwnerNicknameFromKey(ownerKey string) string {
	handle := OwnerHandleFromKey(ownerKey)
	if idx := strings.Index(handle, "@"); idx > 0 {
		return handle[:idx]
	}
	return handle
}
