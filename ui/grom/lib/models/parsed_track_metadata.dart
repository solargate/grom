class ParsedTrackMetadata {
  ParsedTrackMetadata({
    this.name,
    this.sportType,
    this.startDate,
    this.durationSeconds,
    this.durationTotalSeconds,
    this.distanceMeters,
    this.speedMaxKmh,
    this.speedAvgKmh,
    required this.hasGps,
  });

  final String? name;
  final String? sportType;
  final DateTime? startDate;
  final int? durationSeconds;
  final int? durationTotalSeconds;
  final double? distanceMeters;
  final double? speedMaxKmh;
  final double? speedAvgKmh;
  final bool hasGps;

  factory ParsedTrackMetadata.fromJson(Map<String, dynamic> json) {
    final rawName = json['name'];
    final name = rawName is String ? rawName.trim() : null;
    final rawSport = json['sport_type'];
    final sportType = rawSport is String ? rawSport.trim() : null;
    return ParsedTrackMetadata(
      name: (name == null || name.isEmpty) ? null : name,
      sportType: (sportType == null || sportType.isEmpty) ? null : sportType,
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
