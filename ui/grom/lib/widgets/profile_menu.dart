import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

enum ProfileMenuAction {
  edit,
  deleteWorkoutData,
  deleteAccount,
}

class ProfileMenu extends StatelessWidget {
  const ProfileMenu({
    super.key,
    this.onSelected,
    this.deletingWorkoutData = false,
  });

  final ValueChanged<ProfileMenuAction>? onSelected;
  final bool deletingWorkoutData;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return PopupMenuButton<ProfileMenuAction>(
      tooltip: l10n.profileActions,
      onSelected: onSelected,
      itemBuilder: (context) => [
        PopupMenuItem(
          value: ProfileMenuAction.edit,
          child: Text(l10n.editProfile),
        ),
        PopupMenuItem(
          value: ProfileMenuAction.deleteWorkoutData,
          enabled: !deletingWorkoutData,
          child: Text(
            l10n.deleteWorkoutData,
            style: deletingWorkoutData
                ? null
                : TextStyle(color: theme.colorScheme.error),
          ),
        ),
        PopupMenuItem(
          value: ProfileMenuAction.deleteAccount,
          child: Text(
            l10n.deleteAccount,
            style: TextStyle(color: theme.colorScheme.error),
          ),
        ),
      ],
    );
  }
}
