import 'dart:async';

import 'package:cross_file/cross_file.dart';
import 'package:flutter/foundation.dart';
import 'package:receive_sharing_intent/receive_sharing_intent.dart';

import 'shared_track_receive.dart';

export 'shared_track_receive.dart' show SharedTrackPayload, SharedTrackReceiveResult;

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

  final result = await receiveSharedTrackFiles(
    [
      for (final f in files) (path: f.path, mimeType: f.mimeType),
    ],
    readBytes: (path) async => Uint8List.fromList(await XFile(path).readAsBytes()),
  );
  await ReceiveSharingIntent.instance.reset();
  return result;
}

Stream<SharedTrackReceiveResult> watchSharedTracks() {
  if (!_isAndroid) {
    return const Stream.empty();
  }

  return ReceiveSharingIntent.instance.getMediaStream().asyncMap((files) async {
    final result = await receiveSharedTrackFiles(
      [
        for (final f in files) (path: f.path, mimeType: f.mimeType),
      ],
      readBytes: (path) async =>
          Uint8List.fromList(await XFile(path).readAsBytes()),
    );
    await ReceiveSharingIntent.instance.reset();
    return result;
  });
}
