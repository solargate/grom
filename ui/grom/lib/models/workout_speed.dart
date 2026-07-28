/// Max points drawn on the speed chart (downsample threshold).
const int kSpeedChartMaxPoints = 1000;

class WorkoutSpeedSample {
  const WorkoutSpeedSample({
    required this.time,
    required this.speedKmh,
    required this.distanceM,
  });

  final DateTime time;
  final double speedKmh;
  final double distanceM;

  double get distanceKm => distanceM / 1000;

  factory WorkoutSpeedSample.fromJson(Map<String, dynamic> json) {
    return WorkoutSpeedSample(
      time: DateTime.parse(json['t'] as String),
      speedKmh: (json['speed_kmh'] as num).toDouble(),
      distanceM: (json['distance_m'] as num).toDouble(),
    );
  }
}

class WorkoutSpeedSeries {
  const WorkoutSpeedSeries({
    required this.samples,
    this.speedMaxKmh,
    this.speedAvgKmh,
  });

  final List<WorkoutSpeedSample> samples;
  final double? speedMaxKmh;
  final double? speedAvgKmh;

  factory WorkoutSpeedSeries.fromJson(Map<String, dynamic> json) {
    final raw = json['samples'];
    final samples = <WorkoutSpeedSample>[];
    if (raw is List) {
      for (final item in raw) {
        if (item is Map<String, dynamic>) {
          samples.add(WorkoutSpeedSample.fromJson(item));
        }
      }
    }
    return WorkoutSpeedSeries(
      samples: samples,
      speedMaxKmh: (json['speed_max_kmh'] as num?)?.toDouble(),
      speedAvgKmh: (json['speed_avg_kmh'] as num?)?.toDouble(),
    );
  }
}

/// Evenly spaced downsample keeping first/last; no-op when under [maxPoints].
List<WorkoutSpeedSample> downsampleSpeedSamples(
  List<WorkoutSpeedSample> samples, {
  int maxPoints = kSpeedChartMaxPoints,
}) {
  if (samples.length <= maxPoints || maxPoints < 2) {
    return samples;
  }
  final out = <WorkoutSpeedSample>[];
  final last = samples.length - 1;
  var prev = -1;
  for (var i = 0; i < maxPoints; i++) {
    final idx = ((i * last) / (maxPoints - 1)).round();
    if (idx == prev) {
      continue;
    }
    out.add(samples[idx]);
    prev = idx;
  }
  return out;
}

bool _hasPositive(double? value) => value != null && value > 0;

/// Prefer workout/API metadata; otherwise average of sample speeds.
double? resolveSpeedAvgKmh(
  double? metadataAvg,
  List<WorkoutSpeedSample> samples,
) {
  if (_hasPositive(metadataAvg)) {
    return metadataAvg;
  }
  if (samples.isEmpty) {
    return null;
  }
  var sum = 0.0;
  for (final s in samples) {
    sum += s.speedKmh;
  }
  return sum / samples.length;
}

/// Prefer workout/API metadata; otherwise max of sample speeds.
double? resolveSpeedMaxKmh(
  double? metadataMax,
  List<WorkoutSpeedSample> samples,
) {
  if (_hasPositive(metadataMax)) {
    return metadataMax;
  }
  if (samples.isEmpty) {
    return null;
  }
  var max = samples.first.speedKmh;
  for (final s in samples) {
    if (s.speedKmh > max) {
      max = s.speedKmh;
    }
  }
  return max;
}
