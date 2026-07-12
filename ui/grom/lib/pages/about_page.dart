import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

import 'about_page_url.dart' if (dart.library.html) 'about_page_url_web.dart';

const openStreetMapCopyrightUrl =
    'https://www.openstreetmap.org/copyright';

class AboutPage extends StatelessWidget {
  const AboutPage({super.key});

  Future<void> _openCopyrightPage(BuildContext context) async {
    if (kIsWeb) {
      openCopyrightUrlInBrowser(openStreetMapCopyrightUrl);
      return;
    }

    final uri = Uri.parse(openStreetMapCopyrightUrl);
    final launched = await launchUrl(
      uri,
      mode: LaunchMode.externalApplication,
    );
    if (!launched && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(uri.toString())),
      );
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
              l10n.mapDataAttributionTitle,
              style: theme.textTheme.labelLarge?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              l10n.openStreetMapAttribution,
              style: theme.textTheme.bodyLarge,
            ),
            const SizedBox(height: 8),
            Text(
              l10n.openStreetMapLicense,
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 12),
            TextButton.icon(
              onPressed: () => _openCopyrightPage(context),
              icon: const Icon(Icons.open_in_new, size: 18),
              label: Text(l10n.openStreetMapCopyrightLink),
            ),
          ],
        ),
      ),
    );
  }
}
