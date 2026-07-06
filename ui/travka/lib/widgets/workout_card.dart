import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travka/l10n/app_localizations.dart';
import 'package:travka/l10n/sport_type_localizations.dart';
import 'package:travka/models/sport_types.dart';
import 'package:travka/models/workout.dart';

import '../api_request.dart';

const double _mapPreviewAspectRatio = 640 / 360;
const double _mapPreviewMaxWidth = 640;

class WorkoutCard extends StatelessWidget {
  const WorkoutCard({
    super.key,
    required this.workout,
    required this.authToken,
  });

  final Workout workout;
  final String authToken;

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
    final api = ApiRequest();

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
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
                        workout.name,
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(height: 4),
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
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.bodyMedium,
                        ),
                      ],
                    ],
                  ),
                ),
              ],
            ),
          ),
          if (workout.hasMapPreview)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              child: LayoutBuilder(
                builder: (context, constraints) {
                  final displayWidth = math.min(
                    _mapPreviewMaxWidth,
                    constraints.maxWidth,
                  );
                  return Align(
                    alignment: Alignment.centerLeft,
                    child: SizedBox(
                      width: displayWidth,
                      height: displayWidth / _mapPreviewAspectRatio,
                      child: Image.network(
                        api.mapPreviewUrl(workout.id),
                        headers: {'Authorization': 'Bearer $authToken'},
                        fit: BoxFit.contain,
                        loadingBuilder: (context, child, loadingProgress) {
                          if (loadingProgress == null) {
                            return child;
                          }
                          return ColoredBox(
                            color: theme.colorScheme.surfaceContainerHighest,
                            child: const Center(
                              child: CircularProgressIndicator(),
                            ),
                          );
                        },
                        errorBuilder: (context, error, stackTrace) {
                          return const SizedBox.shrink();
                        },
                      ),
                    ),
                  );
                },
              ),
            ),
        ],
      ),
    );
  }
}
