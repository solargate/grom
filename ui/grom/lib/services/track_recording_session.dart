import '../models/recorded_track_point.dart';
import 'track_recording_state.dart';

class TrackRecordingSession {
  const TrackRecordingSession({
    required this.state,
    required this.startTime,
    required this.accumulatedDurationMs,
    this.segmentStartedAt,
    this.finishedDurationTotalSeconds,
    required this.points,
  });

  final TrackRecordingState state;
  final DateTime startTime;
  final int accumulatedDurationMs;
  final DateTime? segmentStartedAt;
  final int? finishedDurationTotalSeconds;
  final List<RecordedTrackPoint> points;

  bool get isActive =>
      state == TrackRecordingState.recording ||
      state == TrackRecordingState.autoPaused ||
      state == TrackRecordingState.paused;

  Map<String, dynamic> toJson() {
    return {
      'state': state.name,
      'startTime': startTime.toUtc().toIso8601String(),
      'accumulatedDurationMs': accumulatedDurationMs,
      if (segmentStartedAt != null)
        'segmentStartedAt': segmentStartedAt!.toUtc().toIso8601String(),
      if (finishedDurationTotalSeconds != null)
        'finishedDurationTotalSeconds': finishedDurationTotalSeconds,
      'points': points.map((point) => point.toJson()).toList(),
    };
  }

  factory TrackRecordingSession.fromJson(Map<String, dynamic> json) {
    return TrackRecordingSession(
      state: TrackRecordingState.values.byName(json['state'] as String),
      startTime: DateTime.parse(json['startTime'] as String).toLocal(),
      accumulatedDurationMs: json['accumulatedDurationMs'] as int,
      segmentStartedAt: json['segmentStartedAt'] == null
          ? null
          : DateTime.parse(json['segmentStartedAt'] as String).toLocal(),
      finishedDurationTotalSeconds:
          json['finishedDurationTotalSeconds'] as int?,
      points: (json['points'] as List<dynamic>)
          .map(
            (item) => RecordedTrackPoint.fromJson(
              Map<String, dynamic>.from(item as Map),
            ),
          )
          .toList(),
    );
  }
}
