import 'dart:async';
import 'dart:io';

import 'package:cross_file/cross_file.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:receive_sharing_intent/receive_sharing_intent.dart';

import 'shared_track_receive.dart';

export 'shared_track_receive.dart'
    show SharedTrackPayload, SharedTrackReceiveResult;

const _contentTracks = MethodChannel('grom/shared_tracks');
const _empty = (payload: null, readFailed: false, unsupportedFormat: false);

bool get _isAndroid =>
    !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

Future<SharedTrackReceiveResult> takePendingSharedTrack() async {
  if (!_isAndroid) {
    return _empty;
  }
  final content = await _takeContentTrack();
  if (content != null) return content;

  final files = await ReceiveSharingIntent.instance.getInitialMedia();
  if (files.isEmpty) return _empty;
  return _receivePluginFiles(files);
}

Stream<SharedTrackReceiveResult> watchSharedTracks() {
  if (!_isAndroid) {
    return const Stream.empty();
  }

  return Stream.multi((controller) {
    var draining = false;
    var notified = false;
    var canceled = false;

    Future<void> drainContentTracks() async {
      notified = true;
      if (draining) return;
      draining = true;
      try {
        while (!canceled) {
          notified = false;
          final track = await _takeContentTrack();
          if (track == null) {
            if (notified) continue;
            break;
          }
          if (!canceled) controller.add(track);
        }
      } finally {
        draining = false;
      }
    }

    _contentTracks.setMethodCallHandler((_) => drainContentTracks());
    final subscription = ReceiveSharingIntent.instance
        .getMediaStream()
        .asyncMap(_receivePluginFiles)
        .listen(controller.add, onError: controller.addError);
    controller.onCancel = () async {
      canceled = true;
      _contentTracks.setMethodCallHandler(null);
      await subscription.cancel();
    };
    unawaited(drainContentTracks());
  });
}

Future<SharedTrackReceiveResult?> _takeContentTrack() async {
  final Map<String, Object?>? item;
  try {
    item = await _contentTracks.invokeMapMethod<String, Object?>(
      'takePendingTrack',
    );
  } on MissingPluginException {
    return null;
  } on PlatformException {
    return null;
  }
  if (item == null) return null;
  final path = item['path'] as String?;
  if (path == null) {
    return (payload: null, readFailed: true, unsupportedFormat: false);
  }
  final file = File(path);
  try {
    return await receiveSharedTrackFiles(
      [(path: path, mimeType: item['mimeType'] as String?)],
      readBytes: (_) => file.readAsBytes(),
    );
  } finally {
    try {
      await file.delete();
    } on FileSystemException {
      // Android can also evict cached files.
    }
  }
}

Future<SharedTrackReceiveResult> _receivePluginFiles(
  List<SharedMediaFile> files,
) async {
  final result = await receiveSharedTrackFiles(
    [for (final file in files) (path: file.path, mimeType: file.mimeType)],
    readBytes: (path) => XFile(path).readAsBytes(),
  );
  await ReceiveSharingIntent.instance.reset();
  return result;
}
