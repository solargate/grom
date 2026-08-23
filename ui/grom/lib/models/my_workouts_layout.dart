import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/models/workout.dart';
import 'package:intl/intl.dart';

/// Display mode for Home → My workouts.
enum MyWorkoutsLayout {
  cards,
  list,
}

/// Whether the primary compact-list metric for [sportType] is distance.
///
/// Distance: foot, cycle, water, winter.
/// Duration: strength, team, racket, other (and unknown types).
bool sportTypeUsesDistanceMetric(String sportType) {
  final category = sportTypeById(sportType)?.category;
  switch (category) {
    case SportCategory.foot:
    case SportCategory.cycle:
    case SportCategory.water:
    case SportCategory.winter:
      return true;
    case SportCategory.strength:
    case SportCategory.team:
    case SportCategory.racket:
    case SportCategory.other:
    case null:
      return false;
  }
}

/// Primary metric for a compact list row, or `null` when the value is zero/empty.
String? primaryMetricForWorkout(AppLocalizations l10n, Workout workout) {
  if (sportTypeUsesDistanceMetric(workout.sportType)) {
    if (workout.distance <= 0) {
      return null;
    }
    return formatDistanceKm(l10n, workout.distance);
  }
  if (workout.durationSeconds <= 0) {
    return null;
  }
  return formatDuration(l10n, workout.durationSeconds);
}

String formatWorkoutListDate(AppLocalizations l10n, DateTime startDate) {
  return DateFormat.yMMMd(l10n.localeName).add_Hm().format(startDate.toLocal());
}
