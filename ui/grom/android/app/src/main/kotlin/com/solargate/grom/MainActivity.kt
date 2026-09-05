package com.solargate.grom

import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.provider.OpenableColumns
import android.util.Log
import io.flutter.embedding.android.FlutterActivity
import java.io.File
import java.io.FileOutputStream

/**
 * Normalizes incoming track intents before [receive_sharing_intent] reads them.
 *
 * Google Drive "Open with" uses DocumentsContract content URIs. The sharing plugin
 * only copies known document providers (external / downloads / media) and otherwise
 * returns [Uri.getPath], which is not a readable file — so VIEW from Drive never
 * reaches Flutter. We copy those URIs into cache and retarget to a file:// VIEW.
 *
 * ACTION_SEND / SEND_MULTIPLE (file-manager share) are left unchanged.
 */
class MainActivity : FlutterActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        intent?.let { setIntent(normalizeIncomingIntent(Intent(it))) }
        super.onCreate(savedInstanceState)
    }

    override fun onNewIntent(intent: Intent) {
        val normalized = normalizeIncomingIntent(Intent(intent))
        super.onNewIntent(normalized)
        setIntent(normalized)
    }

    private fun normalizeIncomingIntent(intent: Intent): Intent {
        when (intent.action) {
            Intent.ACTION_VIEW,
            Intent.ACTION_EDIT -> return materializeViewTrackIntent(intent)
            Intent.ACTION_SEND,
            Intent.ACTION_SEND_MULTIPLE -> return annotateSendTrackIntent(intent)
            else -> return intent
        }
    }

    /** File-manager share: only fill MIME from display name when missing/generic. */
    private fun annotateSendTrackIntent(intent: Intent): Intent {
        val uri = when (intent.action) {
            Intent.ACTION_SEND -> intent.parcelableExtra<Uri>(Intent.EXTRA_STREAM)
            Intent.ACTION_SEND_MULTIPLE -> {
                intent.parcelableArrayListExtra<Uri>(Intent.EXTRA_STREAM)?.firstOrNull()
            }
            else -> null
        } ?: return intent

        val fromName = TrackIntentNormalize.mimeFromFileName(queryDisplayName(uri))
        if (fromName != null &&
            (intent.type.isNullOrBlank() || TrackIntentNormalize.isGenericMime(intent.type))
        ) {
            intent.type = fromName
        }
        return intent
    }

    /**
     * Drive / Open with: copy content:// document URIs to cache so the Flutter
     * plugin can read a real file path.
     */
    private fun materializeViewTrackIntent(intent: Intent): Intent {
        val uri = intent.data
            ?: intent.parcelableExtra<Uri>(Intent.EXTRA_STREAM)
            ?: return intent

        val displayName = queryDisplayName(uri)
        val mime = TrackIntentNormalize.mimeFromFileName(displayName)
            ?: intent.type?.takeIf { it.isNotBlank() && it != "*/*" }
            ?: contentResolver.getType(uri)?.takeIf { it.isNotBlank() }
            ?: "application/octet-stream"

        if (uri.scheme.equals("file", ignoreCase = true)) {
            intent.action = Intent.ACTION_VIEW
            intent.setDataAndType(uri, mime)
            return intent
        }

        if (!TrackIntentNormalize.shouldMaterializeContentUri(intent.action, uri.scheme)) {
            intent.type = mime
            return intent
        }

        // Always copy content:// for VIEW/EDIT. receive_sharing_intent cannot read
        // Google Drive DocumentsContract URIs (returns Uri.path). SEND is untouched.
        val cached = copyUriToCache(uri, displayName) ?: run {
            Log.w(TAG, "Failed to materialize shared track URI: $uri")
            intent.type = mime
            return intent
        }

        intent.action = Intent.ACTION_VIEW
        @Suppress("DEPRECATION")
        intent.setDataAndType(Uri.fromFile(cached), mime)
        intent.removeExtra(Intent.EXTRA_STREAM)
        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
        return intent
    }

    private fun copyUriToCache(uri: Uri, displayName: String?): File? {
        val safeName = TrackIntentNormalize.sanitizeFileName(
            displayName ?: uri.lastPathSegment ?: "shared_track",
        )
        val target = File(cacheDir, "shared_tracks").apply { mkdirs() }.let { dir ->
            File(dir, "${System.currentTimeMillis()}_$safeName")
        }
        return try {
            contentResolver.openInputStream(uri)?.use { input ->
                FileOutputStream(target).use { output -> input.copyTo(output) }
            } ?: return null
            if (!target.exists() || target.length() == 0L) {
                target.delete()
                return null
            }
            target
        } catch (e: Exception) {
            Log.w(TAG, "copyUriToCache failed", e)
            target.delete()
            null
        }
    }

    private fun queryDisplayName(uri: Uri): String? {
        try {
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
        } catch (e: Exception) {
            Log.w(TAG, "queryDisplayName failed for $uri", e)
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

    companion object {
        private const val TAG = "GromMainActivity"
    }
}
