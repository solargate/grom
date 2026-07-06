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

  Map<String, dynamic> toJson() {
    return {
      'latitude': latitude,
      'longitude': longitude,
      'timestamp': timestamp.toUtc().toIso8601String(),
      if (altitude != null) 'altitude': altitude,
      if (speedMps != null) 'speedMps': speedMps,
      if (heading != null) 'heading': heading,
      if (accuracy != null) 'accuracy': accuracy,
    };
  }

  factory RecordedTrackPoint.fromJson(Map<String, dynamic> json) {
    return RecordedTrackPoint(
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      timestamp: DateTime.parse(json['timestamp'] as String).toLocal(),
      altitude: (json['altitude'] as num?)?.toDouble(),
      speedMps: (json['speedMps'] as num?)?.toDouble(),
      heading: (json['heading'] as num?)?.toDouble(),
      accuracy: (json['accuracy'] as num?)?.toDouble(),
    );
  }
}
