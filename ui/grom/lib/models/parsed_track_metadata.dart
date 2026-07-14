class ParsedTrackMetadata {
  ParsedTrackMetadata({
    this.startDate,
    this.durationSeconds,
    this.durationTotalSeconds,
    this.distanceMeters,
    this.speedMaxKmh,
    this.speedAvgKmh,
    required this.hasGps,
  });

  final DateTime? startDate;
  final int? durationSeconds;
  final int? durationTotalSeconds;
  final double? distanceMeters;
  final double? speedMaxKmh;
  final double? speedAvgKmh;
  final bool hasGps;

  factory ParsedTrackMetadata.fromJson(Map<String, dynamic> json) {
    return ParsedTrackMetadata(
      startDate: json['start_date'] != null
          ? DateTime.parse(json['start_date'] as String)
          : null,
      durationSeconds: json['duration_seconds'] as int?,
      durationTotalSeconds: json['duration_total_seconds'] as int?,
      distanceMeters: (json['distance'] as num?)?.toDouble(),
      speedMaxKmh: (json['speed_max_kmh'] as num?)?.toDouble(),
      speedAvgKmh: (json['speed_avg_kmh'] as num?)?.toDouble(),
      hasGps: json['has_gps'] as bool? ?? false,
    );
  }
}
