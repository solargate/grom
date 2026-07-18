import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/models/workout.dart';

class WorkoutStatItem {
  const WorkoutStatItem({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;
}

const workoutStatsPerRow = 3;

bool _isFootSport(String sportType) {
  return sportTypeById(sportType)?.category == SportCategory.foot;
}

bool _hasPositiveNumber(num? value) => value != null && value > 0;

bool _hasPositiveInt(int? value) => value != null && value > 0;

String formatElevationMeters(AppLocalizations l10n, double meters) {
  final text = meters >= 10
      ? meters.round().toString()
      : meters.toStringAsFixed(meters == meters.roundToDouble() ? 0 : 1);
  return l10n.elevationMeters(text);
}

String formatSpeedAvgKmh(AppLocalizations l10n, double kmh) {
  final text = kmh >= 10
      ? kmh.toStringAsFixed(1)
      : kmh.toStringAsFixed(2);
  return l10n.speedKmh(text);
}

String formatHeartRate(AppLocalizations l10n, double bpm) {
  return l10n.heartRateBpm(bpm.round().toString());
}

String formatSteps(AppLocalizations l10n, int steps) {
  return l10n.stepsCount(steps.toString());
}

String formatCalories(AppLocalizations l10n, double calories) {
  return l10n.caloriesKcal(calories.round().toString());
}

/// Builds ordered workout stats, skipping missing/non-positive values.
List<WorkoutStatItem> buildWorkoutStats(
  AppLocalizations l10n,
  Workout workout,
) {
  final stats = <WorkoutStatItem>[];

  if (_hasPositiveNumber(workout.distance)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutDistance,
        value: formatDistanceKm(l10n, workout.distance),
      ),
    );
  }

  final pace = workout.tempAvgKmm?.trim();
  if (_isFootSport(workout.sportType) && pace != null && pace.isNotEmpty) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutPace,
        value: pace,
      ),
    );
  }

  if (_hasPositiveInt(workout.durationSeconds)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutDuration,
        value: formatDuration(l10n, workout.durationSeconds),
      ),
    );
  }

  if (_hasPositiveNumber(workout.elevationGain)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutElevationGain,
        value: formatElevationMeters(l10n, workout.elevationGain!),
      ),
    );
  }

  if (_hasPositiveNumber(workout.speedAvgKmh)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutSpeedAvg,
        value: formatSpeedAvgKmh(l10n, workout.speedAvgKmh!),
      ),
    );
  }

  if (_hasPositiveInt(workout.durationTotalSeconds)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutTotalTime,
        value: formatDuration(l10n, workout.durationTotalSeconds!),
      ),
    );
  }

  if (_hasPositiveNumber(workout.heartRateAvg)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutHeartRateAvg,
        value: formatHeartRate(l10n, workout.heartRateAvg!),
      ),
    );
  }

  if (_hasPositiveInt(workout.stepsTotal)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutSteps,
        value: formatSteps(l10n, workout.stepsTotal!),
      ),
    );
  }

  if (_hasPositiveNumber(workout.calories)) {
    stats.add(
      WorkoutStatItem(
        label: l10n.workoutCalories,
        value: formatCalories(l10n, workout.calories!),
      ),
    );
  }

  return stats;
}

List<List<WorkoutStatItem>> chunkWorkoutStats(
  List<WorkoutStatItem> stats, {
  int perRow = workoutStatsPerRow,
  int? maxRows,
}) {
  if (stats.isEmpty || perRow <= 0) {
    return const [];
  }
  final limited = maxRows == null
      ? stats
      : stats.take(perRow * maxRows).toList(growable: false);
  final rows = <List<WorkoutStatItem>>[];
  for (var i = 0; i < limited.length; i += perRow) {
    final end = i + perRow > limited.length ? limited.length : i + perRow;
    rows.add(limited.sublist(i, end));
  }
  return rows;
}
