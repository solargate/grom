class WorkoutHeartRateSample {
  const WorkoutHeartRateSample({
    required this.time,
    required this.heartRateBpm,
    this.distanceM,
  });

  final DateTime time;
  final double heartRateBpm;
  final double? distanceM;

  double? get distanceKm =>
      distanceM == null ? null : distanceM! / 1000;

  factory WorkoutHeartRateSample.fromJson(Map<String, dynamic> json) {
    return WorkoutHeartRateSample(
      time: DateTime.parse(json['t'] as String),
      heartRateBpm: (json['heart_rate_bpm'] as num).toDouble(),
      distanceM: (json['distance_m'] as num?)?.toDouble(),
    );
  }
}

class WorkoutHeartRateSeries {
  const WorkoutHeartRateSeries({
    required this.samples,
    this.heartRateMax,
    this.heartRateAvg,
    this.hasGps = false,
  });

  final List<WorkoutHeartRateSample> samples;
  final double? heartRateMax;
  final double? heartRateAvg;
  final bool hasGps;

  factory WorkoutHeartRateSeries.fromJson(Map<String, dynamic> json) {
    final raw = json['samples'];
    final samples = <WorkoutHeartRateSample>[];
    if (raw is List) {
      for (final item in raw) {
        if (item is Map<String, dynamic>) {
          samples.add(WorkoutHeartRateSample.fromJson(item));
        }
      }
    }
    return WorkoutHeartRateSeries(
      samples: samples,
      heartRateMax: (json['heart_rate_max'] as num?)?.toDouble(),
      heartRateAvg: (json['heart_rate_avg'] as num?)?.toDouble(),
      hasGps: json['has_gps'] as bool? ?? false,
    );
  }
}

bool _hasPositive(double? value) => value != null && value > 0;

/// Prefer workout/API metadata; otherwise average of sample BPM.
double? resolveHeartRateAvg(
  double? metadataAvg,
  List<WorkoutHeartRateSample> samples,
) {
  if (_hasPositive(metadataAvg)) {
    return metadataAvg;
  }
  if (samples.isEmpty) {
    return null;
  }
  var sum = 0.0;
  for (final s in samples) {
    sum += s.heartRateBpm;
  }
  return sum / samples.length;
}

/// Prefer workout/API metadata; otherwise max of sample BPM.
double? resolveHeartRateMax(
  double? metadataMax,
  List<WorkoutHeartRateSample> samples,
) {
  if (_hasPositive(metadataMax)) {
    return metadataMax;
  }
  if (samples.isEmpty) {
    return null;
  }
  var max = samples.first.heartRateBpm;
  for (final s in samples) {
    if (s.heartRateBpm > max) {
      max = s.heartRateBpm;
    }
  }
  return max;
}

/// Minutes from the first sample in the series.
double minutesFromSeriesStart(
  List<WorkoutHeartRateSample> samples,
  DateTime at,
) {
  if (samples.isEmpty) {
    return 0;
  }
  return at.difference(samples.first.time).inMilliseconds / 60000.0;
}
