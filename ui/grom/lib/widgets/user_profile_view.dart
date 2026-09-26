import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../models/social.dart';
import '../widgets/user_avatar.dart';

class ProfileIdentityCard extends StatelessWidget {
  const ProfileIdentityCard({
    super.key,
    required this.nickname,
    required this.name,
    required this.hasAvatar,
    this.avatarUrl,
    this.authToken,
    this.onTap,
    this.trailing,
    this.onAvatarTap,
  });

  final String nickname;
  final String name;
  final bool hasAvatar;
  final String? avatarUrl;
  final String? authToken;
  final VoidCallback? onTap;
  final VoidCallback? onAvatarTap;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              UserAvatar(
                nickname: nickname,
                hasAvatar: hasAvatar,
                avatarUrl: avatarUrl,
                authToken: authToken,
                radius: 28,
                onTap: onAvatarTap,
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      nickname,
                      style: theme.textTheme.titleLarge,
                    ),
                    if (name.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        name,
                        style: theme.textTheme.bodyLarge?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              if (trailing != null) trailing!,
            ],
          ),
        ),
      ),
    );
  }
}

class ProfileFollowCountCards extends StatelessWidget {
  const ProfileFollowCountCards({
    super.key,
    required this.followingCount,
    required this.followersCount,
    required this.onFollowingTap,
    required this.onFollowersTap,
  });

  final int followingCount;
  final int followersCount;
  final VoidCallback onFollowingTap;
  final VoidCallback onFollowersTap;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Row(
      children: [
        Expanded(
          child: ProfileCountCard(
            label: l10n.followingCount(followingCount),
            onTap: onFollowingTap,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: ProfileCountCard(
            label: l10n.followersCount(followersCount),
            onTap: onFollowersTap,
          ),
        ),
      ],
    );
  }
}

class ProfileCountCard extends StatelessWidget {
  const ProfileCountCard({
    super.key,
    required this.label,
    required this.onTap,
  });

  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
          child: Text(
            label,
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
      ),
    );
  }
}

int activeFollowingCount(List<FollowInfo> following) =>
    following.where((f) => f.status == 'active').length;
