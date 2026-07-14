import 'track_recording_foreground_platform.dart';
import 'track_recording_foreground_stub.dart'
    if (dart.library.io) 'track_recording_foreground_android.dart';

export 'track_recording_foreground_platform.dart';

class TrackRecordingForeground {
  TrackRecordingForeground._();

  static final TrackRecordingForegroundPlatform _platform =
      createTrackRecordingForegroundPlatform();

  static bool get isSupported => _platform.isSupported;

  static Future<void> initialize() => _platform.initialize();

  static Future<bool> ensureNotificationPermission() =>
      _platform.ensureNotificationPermission();

  static Future<bool> isRunning() => _platform.isRunning();

  static Future<ForegroundServiceResult> start({
    required String title,
    required String text,
    required String channelName,
    String? pausedText,
    String? autoPausedText,
  }) =>
      _platform.start(
        title: title,
        text: text,
        channelName: channelName,
        pausedText: pausedText,
        autoPausedText: autoPausedText,
      );

  static Future<ForegroundServiceResult> updateNotification({
    required String title,
    required String text,
  }) =>
      _platform.updateNotification(title: title, text: text);

  static Future<void> sendCommand(String type) => _platform.sendCommand(type);

  static Future<ForegroundServiceResult> stop() => _platform.stop();

  static void registerTaskDataCallback(void Function(Object data) callback) =>
      _platform.registerTaskDataCallback(callback);
}
