import 'dart:async';

import 'package:cross_file/cross_file.dart';
import 'package:flutter/foundation.dart';
import 'package:receive_sharing_intent/receive_sharing_intent.dart';

import 'shared_track_intent_stub.dart';

export 'shared_track_intent_stub.dart' show SharedTrackPayload, SharedTrackReceiveResult;

bool get _isAndroid =>
    !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

Future<SharedTrackReceiveResult> takePendingSharedTrack() async {
  if (!_isAndroid) {
    return (payload: null, readFailed: false, unsupportedFormat: false);
  }

  final files = await ReceiveSharingIntent.instance.getInitialMedia();
  if (files.isEmpty) {
    return (payload: null, readFailed: false, unsupportedFormat: false);
  }

  final result = await _resultFromSharedFiles(files);
  await ReceiveSharingIntent.instance.reset();
  return result;
}

Stream<SharedTrackPayload> watchSharedTracks() {
  if (!_isAndroid) {
    return const Stream.empty();
  }

  return ReceiveSharingIntent.instance.getMediaStream().asyncExpand((files) async* {
    final result = await _resultFromSharedFiles(files);
    await ReceiveSharingIntent.instance.reset();
    if (result.payload != null) {
      yield result.payload!;
    }
  });
}

Future<SharedTrackReceiveResult> _resultFromSharedFiles(
  List<SharedMediaFile> files,
) async {
  var readFailed = false;
  var hadCandidate = false;

  for (final file in files) {
    if (!_looksLikeTrackFile(file)) {
      continue;
    }
    hadCandidate = true;

    if (!_isSupportedSharedTrack(file)) {
      continue;
    }

    final filename = _resolveTrackFilename(file);

    try {
      final bytes = await XFile(file.path).readAsBytes();
      return (
        payload: (filename: filename, bytes: Uint8List.fromList(bytes)),
        readFailed: false,
        unsupportedFormat: false,
      );
    } catch (e) {
      debugPrint('Failed to read shared track: $e');
      readFailed = true;
    }
  }

  return (
    payload: null,
    readFailed: readFailed,
    unsupportedFormat: hadCandidate && !readFailed,
  );
}

bool _looksLikeTrackFile(SharedMediaFile file) {
  if (_isSupportedTrackFilename(_filenameFromPath(file.path))) {
    return true;
  }
  if (_trackFilenameForMime(file.mimeType) != null) {
    return true;
  }

  final mime = file.mimeType?.toLowerCase().trim();
  return mime == 'application/octet-stream' || mime == '*/*';
}

bool _isSupportedSharedTrack(SharedMediaFile file) {
  if (_isSupportedTrackFilename(_filenameFromPath(file.path))) {
    return true;
  }
  return _trackFilenameForMime(file.mimeType) != null;
}

String _resolveTrackFilename(SharedMediaFile file) {
  final fromPath = _filenameFromPath(file.path);
  if (_isSupportedTrackFilename(fromPath)) {
    return fromPath;
  }
  return _trackFilenameForMime(file.mimeType) ?? fromPath;
}

String? _trackFilenameForMime(String? mimeType) {
  final mime = mimeType?.toLowerCase().trim();
  if (mime == null || mime.isEmpty) {
    return null;
  }

  return switch (mime) {
    'application/gpx+xml' => 'track.gpx',
    'application/vnd.ant.fit' => 'track.fit',
    _ => null,
  };
}

bool _isSupportedTrackFilename(String filename) {
  final lower = filename.toLowerCase();
  return lower.endsWith('.gpx') || lower.endsWith('.fit');
}

String _filenameFromPath(String path) {
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
