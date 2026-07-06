class RecordedTrackPoint {
  const RecordedTrackPoint({
    required this.latitude,
    required this.longitude,
    required this.timestamp,
    this.altitude,
    this.speedMps,
    this.heading,
    this.accuracy,
  });

  final double latitude;
  final double longitude;
  final DateTime timestamp;
  final double? altitude;
  final double? speedMps;
  final double? heading;
  final double? accuracy;
}
