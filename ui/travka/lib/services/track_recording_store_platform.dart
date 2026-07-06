import 'track_recording_session.dart';

abstract class TrackRecordingStorePlatform {
  Future<void> init();
  TrackRecordingSession? get cachedSession;
  bool get hasActiveSession;
  Future<TrackRecordingSession?> load();
  Future<void> save(TrackRecordingSession session);
  Future<void> clear();
}
