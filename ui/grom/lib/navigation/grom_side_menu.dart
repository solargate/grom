import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import 'grom_destination.dart';

const kSideMenuWidth = 280.0;

class GromSideMenu extends StatelessWidget {
  const GromSideMenu({
    super.key,
    required this.selectedDestination,
    required this.onDestinationSelected,
    required this.serverTitle,
    required this.isLoggedIn,
    required this.onLogout,
    this.nickname,
  });

  final GromDestination selectedDestination;
  final ValueChanged<GromDestination> onDestinationSelected;
  final String serverTitle;
  final String? nickname;
  final bool isLoggedIn;
  final VoidCallback onLogout;

  int get _selectedIndex {
    if (isLoggedIn) {
      switch (selectedDestination) {
        case GromDestination.home:
          return 0;
        case GromDestination.userSearch:
          return 1;
        case GromDestination.profile:
          return 2;
        case GromDestination.equipment:
          return 3;
        case GromDestination.integration:
          return 4;
        case GromDestination.settings:
          return 5;
        case GromDestination.about:
          return 6;
        case GromDestination.login:
        case GromDestination.register:
          return 0;
      }
    }

    switch (selectedDestination) {
      case GromDestination.home:
        return 0;
      case GromDestination.login:
        return 1;
      case GromDestination.register:
        return 2;
      case GromDestination.settings:
        return 3;
      case GromDestination.about:
        return 4;
      case GromDestination.userSearch:
      case GromDestination.profile:
      case GromDestination.equipment:
      case GromDestination.integration:
        return 0;
    }
  }

  void _onDestinationSelected(int index) {
    if (isLoggedIn) {
      switch (index) {
        case 0:
          onDestinationSelected(GromDestination.home);
        case 1:
          onDestinationSelected(GromDestination.userSearch);
        case 2:
          onDestinationSelected(GromDestination.profile);
        case 3:
          onDestinationSelected(GromDestination.equipment);
        case 4:
          onDestinationSelected(GromDestination.integration);
        case 5:
          onDestinationSelected(GromDestination.settings);
        case 6:
          onDestinationSelected(GromDestination.about);
        case 7:
          onLogout();
      }
      return;
    }

    switch (index) {
      case 0:
        onDestinationSelected(GromDestination.home);
      case 1:
        onDestinationSelected(GromDestination.login);
      case 2:
        onDestinationSelected(GromDestination.register);
      case 3:
        onDestinationSelected(GromDestination.settings);
      case 4:
        onDestinationSelected(GromDestination.about);
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
        if (isLoggedIn) ...[
          NavigationDrawerDestination(
            icon: const Icon(Icons.person_search_outlined),
            selectedIcon: const Icon(Icons.person_search),
            label: Text(l10n.userSearch),
          ),
          NavigationDrawerDestination(
            icon: const Icon(Icons.person_outline),
            selectedIcon: const Icon(Icons.person),
            label: Text(l10n.profile),
          ),
          NavigationDrawerDestination(
            icon: const Icon(Icons.inventory_2_outlined),
            selectedIcon: const Icon(Icons.inventory_2),
            label: Text(l10n.equipment),
          ),
          NavigationDrawerDestination(
            icon: const Icon(Icons.integration_instructions_outlined),
            selectedIcon: const Icon(Icons.integration_instructions),
            label: Text(l10n.integration),
          ),
        ] else ...[
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
        ],
        NavigationDrawerDestination(
          icon: const Icon(Icons.settings_outlined),
          selectedIcon: const Icon(Icons.settings),
          label: Text(l10n.settings),
        ),
        NavigationDrawerDestination(
          icon: const Icon(Icons.info_outline),
          selectedIcon: const Icon(Icons.info),
          label: Text(l10n.about),
        ),
        if (isLoggedIn)
          NavigationDrawerDestination(
            icon: const Icon(Icons.logout_outlined),
            selectedIcon: const Icon(Icons.logout),
            label: Text(l10n.signOut),
          ),
      ],
    );
  }
}
