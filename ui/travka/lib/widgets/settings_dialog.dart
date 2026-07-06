import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../locale_storage.dart';

Future<void> showSettingsDialog(
  BuildContext context, {
  required Locale locale,
  required ValueChanged<Locale> onLocaleChanged,
}) {
  return showDialog<void>(
    context: context,
    builder: (context) => SettingsDialog(
      locale: locale,
      onLocaleChanged: onLocaleChanged,
    ),
  );
}

class SettingsDialog extends StatelessWidget {
  const SettingsDialog({
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

    return AlertDialog(
      title: Text(l10n.settings),
      content: SizedBox(
        width: 320,
        child: Column(
          mainAxisSize: MainAxisSize.min,
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
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(MaterialLocalizations.of(context).closeButtonLabel),
        ),
      ],
    );
  }
}
