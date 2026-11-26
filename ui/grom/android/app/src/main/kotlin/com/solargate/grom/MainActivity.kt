package com.solargate.grom

import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.provider.OpenableColumns
import android.util.Log
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import java.io.File
import java.io.FileOutputStream
import java.util.concurrent.Executors

/** Bridges Android Open with content URIs to Flutter after copying them off the UI thread. */
class MainActivity : FlutterActivity() {
    private val trackReader = Executors.newSingleThreadExecutor()
    private val pendingTracks = ArrayDeque<Map<String, Any>>()
    private var trackChannel: MethodChannel? = null

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        trackChannel = MethodChannel(flutterEngine.dartExecutor.binaryMessenger, "grom/shared_tracks")
        trackChannel?.setMethodCallHandler { call, result ->
            if (call.method == "takePendingTrack") {
                result.success(pendingTracks.removeFirstOrNull())
            } else {
                result.notImplemented()
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        val incoming = intent?.let(::Intent)
        val contentUri = incoming?.let(::viewContentUri)
        if (incoming != null) {
            setIntent(if (contentUri == null) normalizeIncomingIntent(incoming) else Intent(Intent.ACTION_MAIN))
        }
        super.onCreate(savedInstanceState)
        if (contentUri != null) {
            readSharedTrack(contentUri, incoming.type)
        }
    }

    override fun onNewIntent(intent: Intent) {
        val incoming = Intent(intent)
        val contentUri = viewContentUri(incoming)
        if (contentUri != null) {
            readSharedTrack(contentUri, incoming.type)
            return
        }
        val normalized = normalizeIncomingIntent(incoming)
        super.onNewIntent(normalized)
        setIntent(normalized)
    }

    override fun onDestroy() {
        trackChannel?.setMethodCallHandler(null)
        trackReader.shutdownNow()
        for (pending in pendingTracks) {
            (pending["path"] as? String)?.let { File(it).delete() }
        }
        super.onDestroy()
    }

    private fun viewContentUri(intent: Intent): Uri? {
        val uri = intent.data ?: intent.parcelableExtra<Uri>(Intent.EXTRA_STREAM) ?: return null
        return uri.takeIf { TrackIntentNormalize.shouldMaterializeContentUri(intent.action, it.scheme) }
    }

    private fun readSharedTrack(uri: Uri, intentType: String?) {
        trackReader.execute {
            val displayName = queryDisplayName(uri)
            val mime = TrackIntentNormalize.mimeFromFileName(displayName)
                ?: intentType?.takeUnless(TrackIntentNormalize::isGenericMime)
                ?: runCatching { contentResolver.getType(uri) }.getOrNull()
                ?: "application/octet-stream"
            val file = copyUriToCache(uri, displayName)
            runOnUiThread {
                if (isDestroyed) {
                    file?.delete()
                } else {
                    pendingTracks.addLast(
                        if (file == null) mapOf("failed" to true)
                        else mapOf("path" to file.path, "mimeType" to mime),
                    )
                    trackChannel?.invokeMethod("pending", null)
                }
            }
        }
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

    private fun materializeViewTrackIntent(intent: Intent): Intent {
        val uri = intent.data ?: intent.parcelableExtra<Uri>(Intent.EXTRA_STREAM) ?: return intent
        if (uri.scheme.equals("file", ignoreCase = true)) {
            val mime = TrackIntentNormalize.mimeFromFileName(uri.lastPathSegment)
                ?: intent.type ?: "application/octet-stream"
            intent.action = Intent.ACTION_VIEW
            intent.setDataAndType(uri, mime)
        }
        return intent
    }

    private fun copyUriToCache(uri: Uri, displayName: String?): File? {
        val safeName = TrackIntentNormalize.sanitizeFileName(
            displayName ?: uri.lastPathSegment ?: "shared_track",
        )
        var target: File? = null
        var complete = false
        return try {
            val dir = File(cacheDir, "shared_tracks").apply { mkdirs() }
            val file = File.createTempFile("track-", "-${safeName.takeLast(160)}", dir)
            target = file
            contentResolver.openInputStream(uri)?.use { input ->
                FileOutputStream(file).use { output ->
                    TrackIntentNormalize.copyTrack(input, output, MAX_TRACK_BYTES)
                }
            } ?: return null
            complete = true
            file
        } catch (e: Exception) {
            Log.w(TAG, "Unable to read shared track", e)
            null
        } finally {
            if (!complete) target?.delete()
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
        // Matches MaxTrackSizeBytes in internal/tracks.
        private const val MAX_TRACK_BYTES = 20L * 1024 * 1024
        private const val TAG = "GromMainActivity"
    }
}
