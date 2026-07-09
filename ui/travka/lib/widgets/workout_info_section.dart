import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travka/l10n/app_localizations.dart';
import 'package:travka/l10n/sport_type_localizations.dart';
import 'package:travka/models/sport_types.dart';
import 'package:travka/models/workout.dart';

class WorkoutInfoSection extends StatelessWidget {
  const WorkoutInfoSection({
    super.key,
    required this.workout,
    this.descriptionMaxLines,
  });

  final Workout workout;
  final int? descriptionMaxLines;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final sportInfo = sportTypeById(workout.sportType);
    final icon = sportInfo?.icon ?? Icons.sports;
    final color = sportTypeColor(workout.sportType);
    final dateText = DateFormat.yMMMd(l10n.localeName).add_Hm().format(
          workout.startDate.toLocal(),
        );
    final durationText = formatDuration(l10n, workout.durationSeconds);
    final distanceText = formatDistance(l10n, workout.distance);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CircleAvatar(
          radius: 24,
          backgroundColor: color.withValues(alpha: 0.15),
          child: Icon(icon, color: color),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                sportTypeLabel(l10n, workout.sportType),
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                dateText,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                '$durationText · $distanceText',
                style: theme.textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w500,
                ),
              ),
              if (workout.description.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  workout.description,
                  maxLines: descriptionMaxLines,
                  overflow: descriptionMaxLines != null
                      ? TextOverflow.ellipsis
                      : null,
                  style: theme.textTheme.bodyMedium,
                ),
              ],
              if (workout.equipment.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  l10n.workoutEquipment,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  workout.equipment.map((item) => item.name).join(', '),
                  style: theme.textTheme.bodyMedium,
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}
