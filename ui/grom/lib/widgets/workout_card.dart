import 'package:flutter/material.dart';

import '../api_request.dart';
import '../models/workout.dart';
import 'workout_header_section.dart';
import 'workout_like_bar.dart';
import 'workout_map_preview.dart';
import 'workout_media_strip.dart';

class WorkoutCard extends StatelessWidget {
  const WorkoutCard({
    super.key,
    required this.workout,
    required this.authToken,
    this.federationEnabled = false,
    this.compact = false,
    this.onTap,
    this.onPhotoTap,
  });

  final Workout workout;
  final String authToken;
  final bool federationEnabled;
  final bool compact;
  final VoidCallback? onTap;
  final ValueChanged<int>? onPhotoTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final content = _buildContent(context, theme);

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

  Widget _buildContent(BuildContext context, ThemeData theme) {
    final api = ApiRequest();
    final owner = workout.ownerNickname;
    final EdgeInsets mapPadding;
    final EdgeInsets mediaPadding;
    if (compact) {
      mapPadding = EdgeInsets.only(
        bottom: workout.hasMedia ? 0 : kWorkoutMediaGap,
      );
      mediaPadding = EdgeInsets.only(
        top: workout.hasMapPreview ? kWorkoutMediaGap : 0,
        bottom: kWorkoutMediaGap,
      );
    } else {
      mapPadding = const EdgeInsets.fromLTRB(16, 0, 16, 16);
      mediaPadding = const EdgeInsets.fromLTRB(16, 0, 16, 16);
    }

    return Column(
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
            statsMaxRows: 1,
          ),
        ),
        if (workout.hasMapPreview)
          Padding(
            padding: mapPadding,
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
            padding: mediaPadding,
            child: WorkoutMediaStrip(
              workout: workout,
              authToken: authToken,
              onPhotoTap: onPhotoTap,
            ),
          ),
        WorkoutLikeBar(
          workout: workout,
          authToken: authToken,
        ),
      ],
    );
  }
}
