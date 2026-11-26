import 'dart:async';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/platform/shared_track_intent_io.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  const tracks = MethodChannel('grom/shared_tracks');
  const pluginEvents = MethodChannel('receive_sharing_intent/events-media');
  final messenger =
      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger;
  late Directory directory;

  setUp(() async {
    debugDefaultTargetPlatformOverride = TargetPlatform.android;
    directory = await Directory.systemTemp.createTemp('grom-shared-track-');
    messenger.setMockMethodCallHandler(pluginEvents, (_) async => null);
  });

  tearDown(() async {
    debugDefaultTargetPlatformOverride = null;
    messenger.setMockMethodCallHandler(tracks, null);
    messenger.setMockMethodCallHandler(pluginEvents, null);
    await directory.delete(recursive: true);
  });

  test('reads and removes an Open with track on cold start', () async {
    final file = File('${directory.path}/ride.gpx');
    final bytes = '<gpx version="1.1"></gpx>'.codeUnits;
    await file.writeAsBytes(bytes);
    messenger.setMockMethodCallHandler(tracks, (_) async {
      return {'path': file.path, 'mimeType': 'application/gpx+xml'};
    });

    final result = await takePendingSharedTrack();
    expect(result.payload?.filename, 'ride.gpx');
    expect(result.payload?.bytes, bytes);
    expect(await file.exists(), isFalse);
  });

  test('receives an Open with notification after the shell starts', () async {
    final empty = Completer<void>();
    Map<String, Object?>? pending;
    messenger.setMockMethodCallHandler(tracks, (_) async {
      final result = pending;
      pending = null;
      if (!empty.isCompleted) empty.complete();
      return result;
    });

    final received = watchSharedTracks().first;
    await empty.future;
    final file = File('${directory.path}/run.fit');
    final bytes = [0x0e, 0x20, 0, 0, 0, 0, 0, 0, 0x2e, 0x46, 0x49, 0x54];
    await file.writeAsBytes(bytes);
    pending = {'path': file.path, 'mimeType': 'application/vnd.ant.fit'};
    await messenger.handlePlatformMessage(
      tracks.name,
      const StandardMethodCodec().encodeMethodCall(const MethodCall('pending')),
      null,
    );

    final result = await received;
    expect(result.payload?.filename, 'run.fit');
    expect(result.payload?.bytes, bytes);
    expect(await file.exists(), isFalse);
  });
}
