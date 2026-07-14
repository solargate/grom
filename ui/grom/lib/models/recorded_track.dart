import 'recorded_track_point.dart';

class RecordedTrack {
  const RecordedTrack({
    required this.points,
    required this.startTime,
    required this.durationSeconds,
    required this.durationTotalSeconds,
    required this.distanceMeters,
    required this.gpxBytes,
  });

  final List<RecordedTrackPoint> points;
  final DateTime startTime;
  final int durationSeconds;
  final int durationTotalSeconds;
  final double distanceMeters;
  final List<int> gpxBytes;
}
