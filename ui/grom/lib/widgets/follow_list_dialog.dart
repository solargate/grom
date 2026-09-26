import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../models/social.dart';
import '../navigation/open_user_profile.dart';
import '../widgets/user_avatar.dart';

Future<void> showFollowersDialog(
  BuildContext context, {
  required List<FollowerInfo> followers,
  String? authToken,
  String? selfNickname,
  bool federationEnabled = false,
}) {
  final l10n = AppLocalizations.of(context)!;
  return showFollowListDialog(
    context,
    title: l10n.followers,
    emptyMessage: l10n.noFollowersYet,
    itemCount: followers.length,
    itemBuilder: (context, index) {
      final follower = followers[index];
      final theme = Theme.of(context);
      return ListTile(
        onTap: () {
          Navigator.of(context).pop();
          openUserProfile(
            context,
            handle: follower.followerHandle,
            nickname: follower.followerNickname,
            selfNickname: selfNickname,
            federationEnabled: federationEnabled,
          );
        },
        leading: UserAvatar(
          nickname: follower.followerNickname,
          hasAvatar: follower.followerHasAvatar,
          avatarUrl: follower.followerAvatarUrl,
          authToken: authToken,
          radius: 20,
        ),
        title: Text(follower.followerNickname),
        subtitle: Text(
          follower.followerName.isNotEmpty
              ? '${follower.followerName} · ${follower.followerHandle}'
              : follower.followerHandle,
          style: theme.textTheme.bodyMedium?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      );
    },
  );
}

Future<void> showFollowingDialog(
  BuildContext context, {
  required List<FollowInfo> following,
  String? authToken,
  String? selfNickname,
  bool federationEnabled = false,
}) {
  final l10n = AppLocalizations.of(context)!;
  return showFollowListDialog(
    context,
    title: l10n.following,
    emptyMessage: l10n.noFollowingYet,
    itemCount: following.length,
    itemBuilder: (context, index) {
      final follow = following[index];
      final theme = Theme.of(context);
      return ListTile(
        onTap: () {
          Navigator.of(context).pop();
          openUserProfile(
            context,
            handle: follow.targetHandle,
            nickname: follow.targetNickname,
            selfNickname: selfNickname,
            federationEnabled: federationEnabled,
          );
        },
        leading: UserAvatar(
          nickname: follow.targetNickname,
          hasAvatar: follow.targetHasAvatar,
          avatarUrl: follow.targetAvatarUrl,
          authToken: authToken,
          radius: 20,
        ),
        title: Text(follow.targetNickname),
        subtitle: Text(
          follow.targetName.isNotEmpty
              ? '${follow.targetName} · ${follow.targetHandle}'
              : follow.targetHandle,
          style: theme.textTheme.bodyMedium?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
        trailing: follow.status == 'pending'
            ? Text(
                l10n.followPending,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              )
            : null,
      );
    },
  );
}

Future<void> showFollowListDialog(
  BuildContext context, {
  required String title,
  required String emptyMessage,
  required int itemCount,
  required IndexedWidgetBuilder itemBuilder,
}) {
  final width = MediaQuery.sizeOf(context).width;
  if (width >= 600) {
    return showDialog<void>(
      context: context,
      builder: (dialogContext) => Dialog(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520, maxHeight: 560),
          child: FollowListDialog(
            title: title,
            emptyMessage: emptyMessage,
            itemCount: itemCount,
            itemBuilder: itemBuilder,
          ),
        ),
      ),
    );
  }

  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (sheetContext) => FollowListDialog(
      title: title,
      emptyMessage: emptyMessage,
      itemCount: itemCount,
      itemBuilder: itemBuilder,
    ),
  );
}

class FollowListDialog extends StatelessWidget {
  const FollowListDialog({
    super.key,
    required this.title,
    required this.emptyMessage,
    required this.itemCount,
    required this.itemBuilder,
  });

  final String title;
  final String emptyMessage;
  final int itemCount;
  final IndexedWidgetBuilder itemBuilder;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final maxListHeight = MediaQuery.sizeOf(context).height * 0.55;

    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 16, 8, 16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(title, style: theme.textTheme.titleLarge),
          ),
          const SizedBox(height: 12),
          if (itemCount == 0)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 24, 16, 24),
              child: Text(
                emptyMessage,
                style: theme.textTheme.bodyLarge,
                textAlign: TextAlign.center,
              ),
            )
          else
            ConstrainedBox(
              constraints: BoxConstraints(maxHeight: maxListHeight),
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: itemCount,
                itemBuilder: itemBuilder,
              ),
            ),
        ],
      ),
    );
  }
}
