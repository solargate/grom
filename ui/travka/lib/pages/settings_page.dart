import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../locale_storage.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({
    super.key,
    required this.locale,
    required this.onLocaleChanged,
  });

  final Locale locale;
  final ValueChanged<Locale> onLocaleChanged;

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

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 480),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
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
}
