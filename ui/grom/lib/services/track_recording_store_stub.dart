import 'track_recording_session.dart';
import 'track_recording_store_platform.dart';

class _NoopTrackRecordingStore implements TrackRecordingStorePlatform {
  @override
  Future<void> init() async {}

  @override
  TrackRecordingSession? get cachedSession => null;

  @override
  bool get hasActiveSession => false;

  @override
  Future<TrackRecordingSession?> load() async => null;

  @override
  Future<void> save(TrackRecordingSession session) async {}

  @override
  Future<void> clear() async {}
}

TrackRecordingStorePlatform createTrackRecordingStorePlatform() =>
    _NoopTrackRecordingStore();
