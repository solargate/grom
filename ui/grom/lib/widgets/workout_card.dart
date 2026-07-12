import 'package:flutter/material.dart';

import '../api_request.dart';
import '../models/workout.dart';
import 'workout_header_section.dart';
import 'workout_map_preview.dart';
import 'workout_media_strip.dart';

class WorkoutCard extends StatelessWidget {
  const WorkoutCard({
    super.key,
    required this.workout,
    required this.authToken,
    this.federationEnabled = false,
    this.onTap,
    this.onPhotoTap,
  });

  final Workout workout;
  final String authToken;
  final bool federationEnabled;
  final VoidCallback? onTap;
  final ValueChanged<int>? onPhotoTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final api = ApiRequest();
    final owner = workout.ownerNickname;

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
              child: WorkoutHeaderSection(
                workout: workout,
                authToken: authToken,
                author: workout.author,
                federationEnabled: federationEnabled,
                descriptionMaxLines: 2,
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
            if (workout.hasMedia)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                child: WorkoutMediaStrip(
                  workout: workout,
                  authToken: authToken,
                  onPhotoTap: onPhotoTap,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
