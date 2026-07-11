import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../models/social.dart';
import 'user_avatar.dart';

class WorkoutAuthorHeader extends StatelessWidget {
  const WorkoutAuthorHeader({
    super.key,
    required this.author,
    required this.authToken,
    this.currentUserNickname,
    this.avatarRadius = 20,
  });

  final WorkoutAuthor author;
  final String authToken;
  final String? currentUserNickname;
  final double avatarRadius;

  String _authorLabel(AppLocalizations l10n) {
    final displayName =
        author.name.isNotEmpty ? author.name : author.nickname;
    final isOwnWorkout = currentUserNickname != null &&
        author.nickname == currentUserNickname;
    if (isOwnWorkout) {
      return displayName;
    }
    return l10n.workoutByAuthor(displayName);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context)!;

    return Row(
      children: [
        UserAvatar(
          nickname: author.nickname,
          hasAvatar: author.hasAvatar,
          avatarUrl: author.avatarUrl,
          authToken: authToken,
          radius: avatarRadius,
        ),
        const SizedBox(width: 10),
        Expanded(
          child: Text(
            _authorLabel(l10n),
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ),
      ],
    );
  }
}
