package com.solargate.grom

import java.io.ByteArrayOutputStream
import java.io.IOException
import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class TrackIntentNormalizeTest {
    @Test
    fun copyTrackEnforcesTheStreamLength() {
        val data = byteArrayOf(1, 2, 3)
        val output = ByteArrayOutputStream()
        TrackIntentNormalize.copyTrack(data.inputStream(), output, 3)
        assertArrayEquals(data, output.toByteArray())
        assertThrows(IOException::class.java) {
            TrackIntentNormalize.copyTrack(data.inputStream(), ByteArrayOutputStream(), 2)
        }
        assertThrows(IOException::class.java) {
            TrackIntentNormalize.copyTrack(byteArrayOf().inputStream(), ByteArrayOutputStream(), 3)
        }
    }

    @Test
    fun sanitizeFileName_stripsPathAndUnsafeChars() {
        assertEquals("ride.gpx", TrackIntentNormalize.sanitizeFileName("ride.gpx"))
        assertEquals(
            "my_ride.fit",
            TrackIntentNormalize.sanitizeFileName("/docs/my ride.fit"),
        )
        assertEquals("shared_track", TrackIntentNormalize.sanitizeFileName("!!!"))
    }

    @Test
    fun isGenericMime_detectsDriveDefaults() {
        assertTrue(TrackIntentNormalize.isGenericMime(null))
        assertTrue(TrackIntentNormalize.isGenericMime("*/*"))
        assertTrue(TrackIntentNormalize.isGenericMime("application/octet-stream"))
        assertFalse(TrackIntentNormalize.isGenericMime("application/gpx+xml"))
    }

    @Test
    fun mimeFromFileName_mapsExtensions() {
        assertEquals("application/gpx+xml", TrackIntentNormalize.mimeFromFileName("A.GPX"))
        assertEquals("application/vnd.ant.fit", TrackIntentNormalize.mimeFromFileName("x.fit"))
        assertNull(TrackIntentNormalize.mimeFromFileName("notes.txt"))
        assertNull(TrackIntentNormalize.mimeFromFileName(null))
    }

    @Test
    fun shouldMaterializeContentUri_onlyForViewEdit() {
        assertTrue(
            TrackIntentNormalize.shouldMaterializeContentUri(
                "android.intent.action.VIEW",
                "content",
            ),
        )
        assertTrue(
            TrackIntentNormalize.shouldMaterializeContentUri(
                "android.intent.action.EDIT",
                "content",
            ),
        )
        assertFalse(
            TrackIntentNormalize.shouldMaterializeContentUri(
                "android.intent.action.SEND",
                "content",
            ),
        )
        assertFalse(
            TrackIntentNormalize.shouldMaterializeContentUri(
                "android.intent.action.VIEW",
                "file",
            ),
        )
    }

    @Test
    fun shouldAnnotateSendMime_onlyForSendActions() {
        assertTrue(TrackIntentNormalize.shouldAnnotateSendMime("android.intent.action.SEND"))
        assertTrue(
            TrackIntentNormalize.shouldAnnotateSendMime("android.intent.action.SEND_MULTIPLE"),
        )
        assertFalse(TrackIntentNormalize.shouldAnnotateSendMime("android.intent.action.VIEW"))
    }
}
