import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/recorded_track_point.dart';
import 'package:grom/services/track_recording_session.dart';
import 'package:grom/services/track_recording_state.dart';

void main() {
  test('TrackRecordingSession round-trips through JSON', () {
    final timestamp = DateTime(2026, 3, 6, 12, 30);
    final session = TrackRecordingSession(
      state: TrackRecordingState.recording,
      startTime: timestamp,
      accumulatedDurationMs: 45000,
      segmentStartedAt: timestamp.add(const Duration(seconds: 45)),
      points: [
        RecordedTrackPoint(
          latitude: 55.75,
          longitude: 37.62,
          timestamp: timestamp,
          altitude: 180,
          speedMps: 2.5,
          heading: 90,
          accuracy: 6,
        ),
      ],
    );

    final restored = TrackRecordingSession.fromJson(session.toJson());

    expect(restored.state, TrackRecordingState.recording);
    expect(restored.startTime, session.startTime);
    expect(restored.accumulatedDurationMs, 45000);
    expect(restored.segmentStartedAt, session.segmentStartedAt);
    expect(restored.points, hasLength(1));
    expect(restored.points.first.latitude, 55.75);
    expect(restored.points.first.longitude, 37.62);
    expect(restored.points.first.speedMps, 2.5);
  });

  test('TrackRecordingSession round-trips autoPaused state', () {
    final timestamp = DateTime(2026, 3, 6, 12, 30);
    final session = TrackRecordingSession(
      state: TrackRecordingState.autoPaused,
      startTime: timestamp,
      accumulatedDurationMs: 120000,
      segmentStartedAt: null,
      points: const [],
    );

    final restored = TrackRecordingSession.fromJson(session.toJson());

    expect(restored.state, TrackRecordingState.autoPaused);
  });
}
