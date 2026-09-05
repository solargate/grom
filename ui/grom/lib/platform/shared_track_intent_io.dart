import 'dart:async';

import 'package:cross_file/cross_file.dart';
import 'package:flutter/foundation.dart';
import 'package:receive_sharing_intent/receive_sharing_intent.dart';

import 'shared_track_detect.dart';
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

Stream<SharedTrackReceiveResult> watchSharedTracks() {
  if (!_isAndroid) {
    return const Stream.empty();
  }

  return ReceiveSharingIntent.instance.getMediaStream().asyncMap((files) async {
    final result = await _resultFromSharedFiles(files);
    await ReceiveSharingIntent.instance.reset();
    return result;
  });
}

Future<SharedTrackReceiveResult> _resultFromSharedFiles(
  List<SharedMediaFile> files,
) async {
  var readFailed = false;
  var hadCandidate = false;

  for (final file in files) {
    if (!_looksLikeTrackCandidate(file)) {
      continue;
    }
    hadCandidate = true;

    late final Uint8List bytes;
    try {
      bytes = Uint8List.fromList(await XFile(file.path).readAsBytes());
    } catch (e) {
      debugPrint('Failed to read shared track: $e');
      readFailed = true;
      continue;
    }

    final fromPath = _filenameFromPath(file.path);
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

bool _looksLikeTrackCandidate(SharedMediaFile file) {
  if (sharedTrackKindFromFilename(_filenameFromPath(file.path)) != null) {
    return true;
  }
  if (sharedTrackKindFromPreciseMime(file.mimeType) != null) {
    return true;
  }
  return isBroadSharedTrackMime(file.mimeType);
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
