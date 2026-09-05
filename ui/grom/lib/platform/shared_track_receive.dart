import 'dart:typed_data';

import 'shared_track_detect.dart';
import 'shared_track_intent_stub.dart';

export 'shared_track_intent_stub.dart' show SharedTrackPayload, SharedTrackReceiveResult;

/// Platform-agnostic shared file handle (path + optional MIME).
typedef SharedTrackFileRef = ({String path, String? mimeType});

/// Whether add-workout may open for a shared track without forcing login.
///
/// [isLoggedIn] is the shell nickname session; a JWT alone is enough when `/me`
/// has not loaded yet (offline / slow cold start from Drive Open with).
bool sharedTrackImportAllowed({required bool isLoggedIn, String? token}) {
  if (isLoggedIn) {
    return true;
  }
  return token != null && token.isNotEmpty;
}

/// Map shared files to a track payload (or failure flags).
///
/// [readBytes] is injected so tests do not need the filesystem or sharing plugin.
Future<SharedTrackReceiveResult> receiveSharedTrackFiles(
  List<SharedTrackFileRef> files, {
  required Future<Uint8List> Function(String path) readBytes,
}) async {
  var readFailed = false;
  var hadCandidate = false;

  for (final file in files) {
    if (!_looksLikeTrackCandidate(file)) {
      continue;
    }
    hadCandidate = true;

    late final Uint8List bytes;
    try {
      bytes = await readBytes(file.path);
    } catch (_) {
      readFailed = true;
      continue;
    }

    final fromPath = filenameFromSharedPath(file.path);
    final kind = resolveSharedTrackKind(
      filename: fromPath,
      mimeType: file.mimeType,
      bytes: bytes,
    );
    if (kind == null) {
      continue;
    }

    final filename = sharedTrackKindFromFilename(fromPath) != null
        ? fromPath
        : filenameForSharedTrackKind(kind);

    return (
      payload: (filename: filename, bytes: bytes),
      readFailed: false,
      unsupportedFormat: false,
    );
  }

  return (
    payload: null,
    readFailed: readFailed,
    unsupportedFormat: hadCandidate && !readFailed,
  );
}

bool _looksLikeTrackCandidate(SharedTrackFileRef file) {
  if (sharedTrackKindFromFilename(filenameFromSharedPath(file.path)) != null) {
    return true;
  }
  if (sharedTrackKindFromPreciseMime(file.mimeType) != null) {
    return true;
  }
  return isBroadSharedTrackMime(file.mimeType);
}

/// Last path segment of a filesystem path or file URI.
String filenameFromSharedPath(String path) {
  final uri = Uri.tryParse(path);
  if (uri != null && uri.pathSegments.isNotEmpty) {
    return uri.pathSegments.last;
  }
  final separator = path.lastIndexOf('/');
  if (separator >= 0 && separator < path.length - 1) {
    return path.substring(separator + 1);
  }
  return path;
}
