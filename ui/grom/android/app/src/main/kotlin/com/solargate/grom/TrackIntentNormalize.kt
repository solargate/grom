package com.solargate.grom

/**
 * Pure helpers for normalizing shared / open-with track intents.
 * Kept free of Android framework types so JVM unit tests can cover them.
 */
object TrackIntentNormalize {
    fun sanitizeFileName(name: String): String {
        val base = name.substringAfterLast('/').substringAfterLast('\\')
        val cleaned = base.replace(Regex("[^A-Za-z0-9._-]"), "_")
        if (cleaned.isBlank() || cleaned.none { it.isLetterOrDigit() }) {
            return "shared_track"
        }
        return cleaned
    }

    fun isGenericMime(type: String?): Boolean {
        val mime = type?.lowercase()?.trim() ?: return true
        return mime == "*/*" || mime == "application/octet-stream"
    }

    fun mimeFromFileName(name: String?): String? {
        val lower = name?.lowercase().orEmpty()
        return when {
            lower.endsWith(".gpx") -> "application/gpx+xml"
            lower.endsWith(".fit") -> "application/vnd.ant.fit"
            else -> null
        }
    }

    /** VIEW/EDIT content:// must be copied; SEND must not be rewritten. */
    fun shouldMaterializeContentUri(action: String?, scheme: String?): Boolean {
        if (scheme.equals("content", ignoreCase = true).not()) {
            return false
        }
        return action == "android.intent.action.VIEW" ||
            action == "android.intent.action.EDIT"
    }

    fun shouldAnnotateSendMime(action: String?): Boolean {
        return action == "android.intent.action.SEND" ||
            action == "android.intent.action.SEND_MULTIPLE"
    }
}
