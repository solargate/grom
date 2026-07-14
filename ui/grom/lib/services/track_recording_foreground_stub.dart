import 'track_recording_foreground_platform.dart';

class _NoopTrackRecordingForeground implements TrackRecordingForegroundPlatform {
  @override
  bool get isSupported => false;

  @override
  Future<void> initialize() async {}

  @override
  Future<bool> ensureNotificationPermission() async => true;

  @override
  Future<bool> isRunning() async => false;

  @override
  Future<ForegroundServiceResult> start({
    required String title,
    required String text,
    required String channelName,
    String? pausedText,
    String? autoPausedText,
  }) async =>
      const ForegroundServiceFailure(error: 'unsupported_platform');

  @override
  Future<ForegroundServiceResult> updateNotification({
    required String title,
    required String text,
  }) async =>
      const ForegroundServiceFailure(error: 'unsupported_platform');

  @override
  Future<void> sendCommand(String type) async {}

  @override
  Future<ForegroundServiceResult> stop() async =>
      const ForegroundServiceFailure(error: 'unsupported_platform');

  @override
  void registerTaskDataCallback(void Function(Object data) callback) {}
}

TrackRecordingForegroundPlatform createTrackRecordingForegroundPlatform() =>
    _NoopTrackRecordingForeground();
