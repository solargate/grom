import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../models/my_workouts_layout.dart';
import '../models/sport_types.dart';
import '../models/workout.dart';
import '../platform/is_mobile_client.dart';

/// Compact one-line workout row for Home → My workouts list layout.
class WorkoutListRow extends StatelessWidget {
  const WorkoutListRow({
    super.key,
    required this.workout,
    this.onTap,
  });

  final Workout workout;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context)!;
    final compact = isMobileClient;
    final sportInfo = sportTypeById(workout.sportType);
    final sportIcon = sportInfo?.icon ?? Icons.sports;
    final sportColor = sportTypeColor(workout.sportType);
    final dateText = formatWorkoutListDate(l10n, workout.startDate);
    final metric = primaryMetricForWorkout(l10n, workout);

    final content = Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          Icon(sportIcon, size: 24, color: sportColor),
          const SizedBox(width: 12),
          Text(
            dateText,
            style: theme.textTheme.bodyMedium?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              workout.name,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          if (metric != null) ...[
            const SizedBox(width: 12),
            Text(
              metric,
              style: theme.textTheme.bodyMedium?.copyWith(
                fontFeatures: const [FontFeature.tabularFigures()],
              ),
            ),
          ],
        ],
      ),
    );

    if (compact) {
      return Material(
        color: theme.colorScheme.surface,
        child: InkWell(
          onTap: onTap,
          child: content,
        ),
      );
    }

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: content,
      ),
    );
  }
}
