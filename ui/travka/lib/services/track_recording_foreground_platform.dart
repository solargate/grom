sealed class ForegroundServiceResult {
  const ForegroundServiceResult();
}

final class ForegroundServiceSuccess extends ForegroundServiceResult {
  const ForegroundServiceSuccess();
}

final class ForegroundServiceFailure extends ForegroundServiceResult {
  const ForegroundServiceFailure({required this.error});

  final Object error;
}

abstract class TrackRecordingForegroundPlatform {
  bool get isSupported;

  Future<void> initialize();
  Future<bool> ensureNotificationPermission();
  Future<bool> isRunning();
  Future<ForegroundServiceResult> start({
    required String title,
    required String text,
    required String channelName,
    String? pausedText,
  });
  Future<ForegroundServiceResult> updateNotification({
    required String title,
    required String text,
  });
  Future<void> sendCommand(String type);
  Future<ForegroundServiceResult> stop();
  void registerTaskDataCallback(void Function(Object data) callback);
}
