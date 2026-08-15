import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

enum ProfileMenuAction {
  edit,
  deleteAccount,
}

class ProfileMenu extends StatelessWidget {
  const ProfileMenu({
    super.key,
    this.onSelected,
  });

  final ValueChanged<ProfileMenuAction>? onSelected;

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
