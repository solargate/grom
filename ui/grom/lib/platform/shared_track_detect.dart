import 'dart:convert';
import 'dart:typed_data';

/// Detected track file kind from name, MIME, or content bytes.
enum SharedTrackKind { gpx, fit }

/// Returns a track kind when [filename] has a known extension.
SharedTrackKind? sharedTrackKindFromFilename(String filename) {
  final lower = filename.toLowerCase();
  if (lower.endsWith('.gpx')) {
    return SharedTrackKind.gpx;
  }
  if (lower.endsWith('.fit')) {
    return SharedTrackKind.fit;
  }
  return null;
}

/// Precise track MIME types that imply a format without sniffing.
SharedTrackKind? sharedTrackKindFromPreciseMime(String? mimeType) {
  final mime = mimeType?.toLowerCase().trim();
  if (mime == null || mime.isEmpty) {
    return null;
  }

  return switch (mime) {
    'application/gpx+xml' ||
    'application/gpx' ||
    'application/x-gpx+xml' =>
      SharedTrackKind.gpx,
    'application/vnd.ant.fit' ||
    'application/fit' ||
    'application/x-fit' =>
      SharedTrackKind.fit,
    _ => null,
  };
}

/// Broad MIME types used by Drive / mail when the real type is unknown.
bool isBroadSharedTrackMime(String? mimeType) {
  final mime = mimeType?.toLowerCase().trim();
  if (mime == null || mime.isEmpty) {
    return true;
  }
  return mime == 'application/octet-stream' ||
      mime == 'text/xml' ||
      mime == 'application/xml' ||
      mime == '*/*';
}

/// Detect GPX (`<gpx`) or FIT (`.FIT` header signature) from file bytes.
SharedTrackKind? sharedTrackKindFromBytes(Uint8List bytes) {
  if (_looksLikeFit(bytes)) {
    return SharedTrackKind.fit;
  }
  if (_looksLikeGpx(bytes)) {
    return SharedTrackKind.gpx;
  }
  return null;
}

/// Resolve kind: filename → precise MIME → content sniff (only for broad MIME).
SharedTrackKind? resolveSharedTrackKind({
  required String filename,
  String? mimeType,
  Uint8List? bytes,
}) {
  final fromName = sharedTrackKindFromFilename(filename);
  if (fromName != null) {
    return fromName;
  }

  final fromMime = sharedTrackKindFromPreciseMime(mimeType);
  if (fromMime != null) {
    return fromMime;
  }

  if (bytes != null && isBroadSharedTrackMime(mimeType)) {
    return sharedTrackKindFromBytes(bytes);
  }
  return null;
}

String filenameForSharedTrackKind(SharedTrackKind kind) {
  return switch (kind) {
    SharedTrackKind.gpx => 'track.gpx',
    SharedTrackKind.fit => 'track.fit',
  };
}

bool _looksLikeFit(Uint8List bytes) {
  // FIT header: header size at byte 0 (12 or 14); ASCII ".FIT" at offset 8.
  if (bytes.length < 12) {
    return false;
  }
  final headerSize = bytes[0];
  if (headerSize != 12 && headerSize != 14) {
    return false;
  }
  if (bytes.length < headerSize) {
    return false;
  }
  return bytes[8] == 0x2e &&
      bytes[9] == 0x46 &&
      bytes[10] == 0x49 &&
      bytes[11] == 0x54;
}

bool _looksLikeGpx(Uint8List bytes) {
  if (bytes.isEmpty) {
    return false;
  }
  final probeLen = bytes.length < 4096 ? bytes.length : 4096;
  // Latin-1 preserves byte values for ASCII tag search.
  final probe = latin1.decode(bytes.sublist(0, probeLen), allowInvalid: true);
  return probe.toLowerCase().contains('<gpx');
}
