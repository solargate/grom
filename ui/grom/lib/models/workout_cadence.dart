import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/sport_types.dart';

/// Whether cadence charts/stats apply (cycle or foot, excluding Wheelchair).
bool supportsCadenceDisplay(String sportType) {
  if (sportType == 'Wheelchair') {
    return false;
  }
  final category = sportTypeById(sportType)?.category;
  return category == SportCategory.foot || category == SportCategory.cycle;
}

/// Raw stored cadence → display units (×2 for foot sports like Strava app).
double? displayCadence(double? raw, String sportType) {
  if (raw == null || raw <= 0) {
    return null;
  }
  if (supportsCadenceDisplay(sportType) &&
      sportTypeById(sportType)?.category == SportCategory.foot) {
    return raw * 2;
  }
  return raw;
}

class WorkoutCadenceSample {
  const WorkoutCadenceSample({
    required this.time,
    required this.cadence,
    this.distanceM,
  });

  final DateTime time;
  final double cadence;
  final double? distanceM;

  double? get distanceKm =>
      distanceM == null ? null : distanceM! / 1000;

  factory WorkoutCadenceSample.fromJson(Map<String, dynamic> json) {
    return WorkoutCadenceSample(
      time: DateTime.parse(json['t'] as String),
      cadence: (json['cadence'] as num).toDouble(),
      distanceM: (json['distance_m'] as num?)?.toDouble(),
    );
  }
}

class WorkoutCadenceSeries {
  const WorkoutCadenceSeries({
    required this.samples,
    this.cadenceMax,
    this.cadenceAvg,
    this.hasGps = false,
  });

  final List<WorkoutCadenceSample> samples;
  final double? cadenceMax;
  final double? cadenceAvg;
  final bool hasGps;

  factory WorkoutCadenceSeries.fromJson(Map<String, dynamic> json) {
    final raw = json['samples'];
    final samples = <WorkoutCadenceSample>[];
    if (raw is List) {
      for (final item in raw) {
        if (item is Map<String, dynamic>) {
          samples.add(WorkoutCadenceSample.fromJson(item));
        }
      }
    }
    return WorkoutCadenceSeries(
      samples: samples,
      cadenceMax: (json['cadence_max'] as num?)?.toDouble(),
      cadenceAvg: (json['cadence_avg'] as num?)?.toDouble(),
      hasGps: json['has_gps'] as bool? ?? false,
    );
  }
}

bool _hasPositive(double? value) => value != null && value > 0;

/// Prefer workout/API metadata; otherwise average of sample cadence (raw units).
double? resolveCadenceAvg(
  double? metadataAvg,
  List<WorkoutCadenceSample> samples,
) {
  if (_hasPositive(metadataAvg)) {
    return metadataAvg;
  }
  if (samples.isEmpty) {
    return null;
  }
  var sum = 0.0;
  for (final s in samples) {
    sum += s.cadence;
  }
  return sum / samples.length;
}

/// Prefer workout/API metadata; otherwise max of sample cadence (raw units).
double? resolveCadenceMax(
  double? metadataMax,
  List<WorkoutCadenceSample> samples,
) {
  if (_hasPositive(metadataMax)) {
    return metadataMax;
  }
  if (samples.isEmpty) {
    return null;
  }
  var max = samples.first.cadence;
  for (final s in samples) {
    if (s.cadence > max) {
      max = s.cadence;
    }
  }
  return max;
}

/// Minutes from the first sample in the series.
double cadenceMinutesFromSeriesStart(
  List<WorkoutCadenceSample> samples,
  DateTime at,
) {
  if (samples.isEmpty) {
    return 0;
  }
  return at.difference(samples.first.time).inMilliseconds / 60000.0;
}

String formatCadence(
  AppLocalizations l10n,
  double value, {
  required bool footUnits,
}) {
  final text = value.round().toString();
  return footUnits
      ? l10n.cadenceStepsPerMin(text)
      : l10n.cadenceRpm(text);
}

bool hasCadenceChart({
  required String sportType,
  required List<WorkoutCadenceSample>? samples,
}) {
  return supportsCadenceDisplay(sportType) &&
      samples != null &&
      samples.length >= 2;
}
