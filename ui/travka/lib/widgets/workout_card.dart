import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../api_request.dart';
import '../models/workout.dart';
import 'workout_info_section.dart';
import 'workout_map_preview.dart';

class WorkoutCard extends StatelessWidget {
  const WorkoutCard({
    super.key,
    required this.workout,
    required this.authToken,
    this.currentUserNickname,
    this.onTap,
  });

  final Workout workout;
  final String authToken;
  final String? currentUserNickname;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final api = ApiRequest();
    final owner = workout.ownerNickname;
    final showAuthor = workout.author != null &&
        currentUserNickname != null &&
        workout.author!.nickname != currentUserNickname;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (showAuthor) ...[
                    Text(
                      AppLocalizations.of(context)!.workoutByAuthor(
                        workout.author!.name.isNotEmpty
                            ? workout.author!.name
                            : workout.author!.nickname,
                      ),
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 4),
                  ],
                  Text(
                    workout.name,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 12),
                  WorkoutInfoSection(
                    workout: workout,
                    descriptionMaxLines: 2,
                  ),
                ],
              ),
            ),
            if (workout.hasMapPreview)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                child: WorkoutMapPreview(
                  child: Image.network(
                    api.mapPreviewUrl(workout.id, owner: owner),
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
              ),
          ],
        ),
      ),
    );
  }
}
