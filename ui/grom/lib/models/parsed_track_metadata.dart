class ParsedTrackMetadata {
  ParsedTrackMetadata({
    this.startDate,
    this.durationSeconds,
    this.distanceMeters,
    required this.hasGps,
  });

  final DateTime? startDate;
  final int? durationSeconds;
  final double? distanceMeters;
  final bool hasGps;

  factory ParsedTrackMetadata.fromJson(Map<String, dynamic> json) {
    return ParsedTrackMetadata(
      startDate: json['start_date'] != null
          ? DateTime.parse(json['start_date'] as String)
          : null,
      durationSeconds: json['duration_seconds'] as int?,
      distanceMeters: (json['distance'] as num?)?.toDouble(),
      hasGps: json['has_gps'] as bool? ?? false,
    );
  }
}
