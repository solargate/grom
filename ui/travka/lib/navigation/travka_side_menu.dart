import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import 'travka_destination.dart';

const kSideMenuWidth = 280.0;

class TravkaSideMenu extends StatelessWidget {
  const TravkaSideMenu({
    super.key,
    required this.selectedDestination,
    required this.onDestinationSelected,
    required this.serverTitle,
    required this.isLoggedIn,
    required this.onLogout,
    required this.onOpenSettings,
    this.nickname,
  });

  final TravkaDestination selectedDestination;
  final ValueChanged<TravkaDestination> onDestinationSelected;
  final String serverTitle;
  final String? nickname;
  final bool isLoggedIn;
  final VoidCallback onLogout;
  final VoidCallback onOpenSettings;

  int get _settingsIndex => isLoggedIn ? 2 : 3;

  int get _selectedIndex {
    if (isLoggedIn) {
      return selectedDestination == TravkaDestination.home ? 0 : 0;
    }

    switch (selectedDestination) {
      case TravkaDestination.home:
        return 0;
      case TravkaDestination.login:
        return 1;
      case TravkaDestination.register:
        return 2;
    }
  }

  void _onDestinationSelected(int index) {
    if (index == _settingsIndex) {
      onOpenSettings();
      return;
    }

    if (isLoggedIn) {
      if (index == 1) {
        onLogout();
      } else {
        onDestinationSelected(TravkaDestination.home);
      }
      return;
    }

    switch (index) {
      case 0:
        onDestinationSelected(TravkaDestination.home);
      case 1:
        onDestinationSelected(TravkaDestination.login);
      case 2:
        onDestinationSelected(TravkaDestination.register);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context)!;

    return NavigationDrawer(
      selectedIndex: _selectedIndex,
      onDestinationSelected: _onDestinationSelected,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(28, 16, 16, 8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                serverTitle,
                style: theme.textTheme.titleLarge,
              ),
              if (nickname != null) ...[
                const SizedBox(height: 4),
                Text(
                  nickname!,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ],
          ),
        ),
        NavigationDrawerDestination(
          icon: const Icon(Icons.home_outlined),
          selectedIcon: const Icon(Icons.home),
          label: Text(l10n.home),
        ),
        if (!isLoggedIn) ...[
          NavigationDrawerDestination(
            icon: const Icon(Icons.login_outlined),
            selectedIcon: const Icon(Icons.login),
            label: Text(l10n.signIn),
          ),
          NavigationDrawerDestination(
            icon: const Icon(Icons.person_add_outlined),
            selectedIcon: const Icon(Icons.person_add),
            label: Text(l10n.register),
          ),
        ] else
          NavigationDrawerDestination(
            icon: const Icon(Icons.logout_outlined),
            selectedIcon: const Icon(Icons.logout),
            label: Text(l10n.signOut),
          ),
        NavigationDrawerDestination(
          icon: const Icon(Icons.settings_outlined),
          selectedIcon: const Icon(Icons.settings),
          label: Text(l10n.settings),
        ),
      ],
    );
  }
}
