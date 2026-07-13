package com.solargate.grom

import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.provider.OpenableColumns
import io.flutter.embedding.android.FlutterActivity

class MainActivity : FlutterActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        normalizeTrackIntent(intent)?.let { setIntent(it) }
        super.onCreate(savedInstanceState)
    }

    override fun onNewIntent(intent: Intent) {
        normalizeTrackIntent(intent)?.let { setIntent(it) }
        super.onNewIntent(intent)
    }

    private fun normalizeTrackIntent(intent: Intent?): Intent? {
        if (intent == null) {
            return null
        }

        when (intent.action) {
            Intent.ACTION_SEND,
            Intent.ACTION_SEND_MULTIPLE,
            Intent.ACTION_VIEW -> Unit
            else -> return intent
        }

        val uri = extractUri(intent) ?: return intent
        if (!intent.type.isNullOrBlank()) {
            return intent
        }

        intent.type = resolveMimeType(uri)
        return intent
    }

    private fun extractUri(intent: Intent): Uri? {
        return when (intent.action) {
            Intent.ACTION_SEND -> intent.parcelableExtra(Intent.EXTRA_STREAM)
            Intent.ACTION_VIEW -> intent.data
            Intent.ACTION_SEND_MULTIPLE -> {
                val uris = intent.parcelableArrayListExtra<Uri>(Intent.EXTRA_STREAM)
                uris?.firstOrNull()
            }
            else -> null
        }
    }

    private fun resolveMimeType(uri: Uri): String {
        contentResolver.getType(uri)?.takeIf { it.isNotBlank() }?.let { return it }

        val name = queryDisplayName(uri)?.lowercase().orEmpty()
        return when {
            name.endsWith(".gpx") -> "application/gpx+xml"
            name.endsWith(".fit") -> "application/vnd.ant.fit"
            else -> "application/octet-stream"
        }
    }

    private fun queryDisplayName(uri: Uri): String? {
        contentResolver.query(
            uri,
            arrayOf(OpenableColumns.DISPLAY_NAME),
            null,
            null,
            null,
        )?.use { cursor ->
            if (cursor.moveToFirst()) {
                val index = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                if (index >= 0) {
                    return cursor.getString(index)
                }
            }
        }
        return uri.lastPathSegment
    }

    private inline fun <reified T : android.os.Parcelable> Intent.parcelableExtra(key: String): T? {
        return if (Build.VERSION.SDK_INT >= 33) {
            getParcelableExtra(key, T::class.java)
        } else {
            @Suppress("DEPRECATION")
            getParcelableExtra(key)
        }
    }

    private inline fun <reified T : android.os.Parcelable> Intent.parcelableArrayListExtra(
        key: String,
    ): ArrayList<T>? {
        return if (Build.VERSION.SDK_INT >= 33) {
            getParcelableArrayListExtra(key, T::class.java)
        } else {
            @Suppress("DEPRECATION")
            getParcelableArrayListExtra(key)
        }
    }
}
