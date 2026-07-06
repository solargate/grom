import 'track_recording_store_platform.dart';
import 'track_recording_store_stub.dart'
    if (dart.library.io) 'track_recording_store_io.dart';

import 'track_recording_session.dart';

class TrackRecordingStore implements TrackRecordingStorePlatform {
  TrackRecordingStore._(this._platform);

  static final TrackRecordingStore instance =
      TrackRecordingStore._(createTrackRecordingStorePlatform());

  final TrackRecordingStorePlatform _platform;

  @override
  Future<void> init() => _platform.init();

  @override
  TrackRecordingSession? get cachedSession => _platform.cachedSession;

  @override
  bool get hasActiveSession => _platform.hasActiveSession;

  @override
  Future<TrackRecordingSession?> load() => _platform.load();

  @override
  Future<void> save(TrackRecordingSession session) => _platform.save(session);

  @override
  Future<void> clear() => _platform.clear();
}
