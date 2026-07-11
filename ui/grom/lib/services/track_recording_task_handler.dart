import 'dart:async';

import 'package:flutter_foreground_task/flutter_foreground_task.dart';
import 'package:geolocator/geolocator.dart';

import '../models/recorded_track_point.dart';
import '../utils/geo_utils.dart';
import 'track_recording_session.dart';
import 'track_recording_state.dart';
import 'track_recording_store.dart';

const trackRecordingServiceId = 256;
const trackRecordingChannelId = 'grom_track_recording';

@pragma('vm:entry-point')
void trackRecordingStartCallback() {
  FlutterForegroundTask.setTaskHandler(TrackRecordingTaskHandler());
}

class TrackRecordingTaskHandler extends TaskHandler {
  StreamSubscription<Position>? _positionSub;
  final TrackRecordingStore _store = TrackRecordingStore.instance;

  @override
  Future<void> onStart(DateTime timestamp, TaskStarter starter) async {
    await _store.init();
    final session = await _store.load();
    if (session?.state == TrackRecordingState.recording) {
      await _startPositionStream();
    }
  }

  @override
  void onRepeatEvent(DateTime timestamp) {
    // Stream-driven recording; keep the handler alive only.
  }

  @override
  Future<void> onDestroy(DateTime timestamp, bool isTimeout) async {
    await _positionSub?.cancel();
    _positionSub = null;
  }

  @override
  void onReceiveData(Object data) {
    if (data is! Map) {
      return;
    }
    final type = data['type'];
    switch (type) {
      case 'pause':
        _pauseStream();
      case 'resume':
        unawaited(_startPositionStream());
      case 'stop':
        unawaited(_stopStream());
    }
  }

  Future<void> _startPositionStream() async {
    await _positionSub?.cancel();
    _positionSub = Geolocator.getPositionStream(
      locationSettings: AndroidSettings(
        accuracy: LocationAccuracy.high,
        distanceFilter: 5,
        intervalDuration: const Duration(seconds: 5),
        foregroundNotificationConfig: null,
      ),
    ).listen(
      _onPosition,
      onError: (Object error) {
        FlutterForegroundTask.sendDataToMain({
          'type': 'error',
          'message': error.toString(),
        });
      },
    );
  }

  Future<void> _pauseStream() async {
    await _positionSub?.cancel();
    _positionSub = null;
  }

  Future<void> _stopStream() async {
    await _pauseStream();
  }

  Future<void> _onPosition(Position position) async {
    final session = await _store.load();
    if (session == null || session.state != TrackRecordingState.recording) {
      return;
    }

    if (!isValidGpsCoordinate(
      position.latitude,
      position.longitude,
      accuracy: position.accuracy,
    )) {
      return;
    }

    final point = RecordedTrackPoint(
      latitude: position.latitude,
      longitude: position.longitude,
      timestamp: position.timestamp,
      altitude: position.altitude,
      speedMps: position.speed >= 0 ? position.speed : null,
      heading: position.heading >= 0 ? position.heading : null,
      accuracy: position.accuracy,
    );

    final updated = TrackRecordingSession(
      state: session.state,
      startTime: session.startTime,
      accumulatedDurationMs: session.accumulatedDurationMs,
      segmentStartedAt: session.segmentStartedAt,
      points: [...session.points, point],
    );
    await _store.save(updated);

    FlutterForegroundTask.sendDataToMain({
      'type': 'point',
      'point': point.toJson(),
      if (position.speed >= 0) 'speedKmh': position.speed * 3.6,
    });

    final notificationText = await FlutterForegroundTask.getData<String>(
      key: 'notificationText',
    );
    if (notificationText != null) {
      FlutterForegroundTask.updateService(
        notificationText: notificationText,
      );
    }
  }
}
