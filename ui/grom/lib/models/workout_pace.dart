import 'package:grom/l10n/app_localizations.dart';

import 'sport_types.dart';
import 'workout_speed.dart';

/// Whether the detail screen should show a pace chart (foot sport + ≥2 samples).
bool hasPaceChart({
  required String sportType,
  required List<WorkoutPaceSample> paceSamples,
}) {
  return isFootSport(sportType) && paceSamples.length >= 2;
}

/// Whether the detail screen should show a speed chart (non-foot + ≥2 samples).
bool hasSpeedChart({
  required String sportType,
  required List<WorkoutSpeedSample>? speedSamples,
}) {
  return !isFootSport(sportType) &&
      speedSamples != null &&
      speedSamples.length >= 2;
}

/// One pace sample (seconds per km) derived from a speed chart point.
class WorkoutPaceSample {
  const WorkoutPaceSample({
    required this.time,
    required this.paceSecPerKm,
    required this.distanceM,
  });

  final DateTime time;
  final double paceSecPerKm;
  final double distanceM;

  double get distanceKm => distanceM / 1000;
}

/// Converts km/h to seconds per km. Returns null when speed is non-positive
/// or the result is not finite.
double? speedKmhToPaceSec(double speedKmh) {
  if (speedKmh <= 0) {
    return null;
  }
  final pace = 3600.0 / speedKmh;
  if (pace.isNaN || pace.isInfinite || pace <= 0) {
    return null;
  }
  return pace;
}

/// Builds pace samples from speed chart points, omitting non-positive speeds.
List<WorkoutPaceSample> paceSamplesFromSpeed(List<WorkoutSpeedSample> samples) {
  if (samples.isEmpty) {
    return const [];
  }
  final out = <WorkoutPaceSample>[];
  for (final s in samples) {
    final pace = speedKmhToPaceSec(s.speedKmh);
    if (pace == null) {
      continue;
    }
    out.add(
      WorkoutPaceSample(
        time: s.time,
        paceSecPerKm: pace,
        distanceM: s.distanceM,
      ),
    );
  }
  return out;
}

/// Parses `m:ss` / `mm:ss` pace strings (e.g. workout `temp_avg_kmm`).
double? parsePaceMmSs(String? raw) {
  if (raw == null) {
    return null;
  }
  final trimmed = raw.trim();
  if (trimmed.isEmpty) {
    return null;
  }
  final parts = trimmed.split(':');
  if (parts.length != 2) {
    return null;
  }
  final mins = int.tryParse(parts[0]);
  final secs = int.tryParse(parts[1]);
  if (mins == null || secs == null || mins < 0 || secs < 0 || secs >= 60) {
    return null;
  }
  return (mins * 60 + secs).toDouble();
}

/// Formats pace seconds as `m:ss` (minutes may exceed 59).
String formatPaceMmSs(double paceSec) {
  if (paceSec <= 0 || paceSec.isNaN || paceSec.isInfinite) {
    return '--';
  }
  final total = paceSec.round();
  final mins = total ~/ 60;
  final secs = total % 60;
  return '$mins:${secs.toString().padLeft(2, '0')}';
}

/// Formats pace with localized unit (e.g. `6:00 /km`).
String formatPaceWithUnit(AppLocalizations l10n, double paceSec) {
  return l10n.paceMinKm(formatPaceMmSs(paceSec));
}

/// Appends the localized pace unit to an already formatted `m:ss` string.
String formatPaceStringWithUnit(AppLocalizations l10n, String paceMmSs) {
  return l10n.paceMinKm(paceMmSs);
}

bool _hasPositive(double? value) => value != null && value > 0;

/// Prefer `temp_avg_kmm`, then speed avg → pace, else mean of pace samples.
double? resolveAvgPaceSec({
  String? tempAvgKmm,
  double? speedAvgKmh,
  required List<WorkoutPaceSample> samples,
}) {
  final fromMeta = parsePaceMmSs(tempAvgKmm);
  if (_hasPositive(fromMeta)) {
    return fromMeta;
  }
  final fromSpeed = speedKmhToPaceSec(speedAvgKmh ?? 0);
  if (_hasPositive(fromSpeed)) {
    return fromSpeed;
  }
  if (samples.isEmpty) {
    return null;
  }
  var sum = 0.0;
  for (final s in samples) {
    sum += s.paceSecPerKm;
  }
  return sum / samples.length;
}

/// Prefer pace from max speed (best = fastest); else min pace in samples.
double? resolveBestPaceSec({
  double? speedMaxKmh,
  required List<WorkoutPaceSample> samples,
}) {
  final fromMax = speedKmhToPaceSec(speedMaxKmh ?? 0);
  if (_hasPositive(fromMax)) {
    return fromMax;
  }
  if (samples.isEmpty) {
    return null;
  }
  var best = samples.first.paceSecPerKm;
  for (final s in samples) {
    if (s.paceSecPerKm < best) {
      best = s.paceSecPerKm;
    }
  }
  return best;
}
