import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../locale_storage.dart';
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
    required this.locale,
    required this.onLocaleChanged,
    this.nickname,
  });

  final TravkaDestination selectedDestination;
  final ValueChanged<TravkaDestination> onDestinationSelected;
  final String serverTitle;
  final String? nickname;
  final bool isLoggedIn;
  final VoidCallback onLogout;
  final Locale locale;
  final ValueChanged<Locale> onLocaleChanged;

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

  String _languageLabel(AppLocalizations l10n, String code) {
    switch (code) {
      case 'ru':
        return l10n.languageRussian;
      case 'de':
        return l10n.languageGerman;
      default:
        return l10n.languageEnglish;
    }
  }

  Widget _buildLanguageSelector(BuildContext context, AppLocalizations l10n) {
    final theme = Theme.of(context);

    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(28, 8, 16, 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            const Divider(),
            const SizedBox(height: 8),
            Text(
              l10n.language,
              style: theme.textTheme.labelLarge?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 8),
            DropdownButtonFormField<String>(
              key: ValueKey(locale.languageCode),
              initialValue: locale.languageCode,
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                isDense: true,
              ),
              items: [
                for (final code in supportedLocaleCodes)
                  DropdownMenuItem(
                    value: code,
                    child: Text(_languageLabel(l10n, code)),
                  ),
              ],
              onChanged: (code) {
                if (code != null) {
                  onLocaleChanged(Locale(code));
                }
              },
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context)!;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Expanded(
          child: NavigationDrawer(
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
            ],
          ),
        ),
        _buildLanguageSelector(context, l10n),
      ],
    );
  }
}
