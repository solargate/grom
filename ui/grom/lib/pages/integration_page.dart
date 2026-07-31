import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

import '../services/strava_import_service.dart';
import 'about_page_url.dart' if (dart.library.html) 'about_page_url_web.dart';

class IntegrationPage extends StatefulWidget {
  const IntegrationPage({super.key});

  @override
  State<IntegrationPage> createState() => _IntegrationPageState();
}

class _IntegrationPageState extends State<IntegrationPage> {
  final StravaImportService _importService = StravaImportService.instance;
  late final TapGestureRecognizer _stravaDownloadLinkRecognizer;

  @override
  void initState() {
    super.initState();
    _stravaDownloadLinkRecognizer = TapGestureRecognizer()
      ..onTap = _openStravaDownloadUrl;
    _importService.addListener(_onImportChanged);
    _importService.syncFromServer();
  }

  @override
  void dispose() {
    _stravaDownloadLinkRecognizer.dispose();
    _importService.removeListener(_onImportChanged);
    super.dispose();
  }

  void _onImportChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  Future<void> _openStravaDownloadUrl() async {
    final l10n = AppLocalizations.of(context)!;
    final url = l10n.stravaDownloadArchiveUrl;
    if (kIsWeb) {
      openCopyrightUrlInBrowser(url);
      return;
    }

    final uri = Uri.parse(url);
    final launched = await launchUrl(
      uri,
      mode: LaunchMode.externalApplication,
    );
    if (!launched && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(uri.toString())),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final state = _importService.state;

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text(
          l10n.strava,
          style: theme.textTheme.titleLarge,
        ),
        const SizedBox(height: 16),
        Text.rich(
          TextSpan(
            style: theme.textTheme.bodyMedium,
            children: [
              TextSpan(text: l10n.stravaImportDescriptionBefore),
              TextSpan(
                text: l10n.stravaDownloadArchiveUrl,
                style: TextStyle(
                  color: theme.colorScheme.primary,
                  decoration: TextDecoration.underline,
                ),
                recognizer: _stravaDownloadLinkRecognizer,
              ),
              TextSpan(text: l10n.stravaImportDescriptionAfter),
            ],
          ),
        ),
        const SizedBox(height: 16),
        FilledButton.icon(
          onPressed: state.active ? null : _importService.pickAndImport,
          icon: const Icon(Icons.upload_file),
          label: Text(l10n.importStravaArchive),
        ),
        if (state.showUploadProgress) ...[
          const SizedBox(height: 24),
          Text(l10n.uploading),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: state.uploadProgress > 0 ? state.uploadProgress : null,
          ),
        ],
        if (state.showImportProgress) ...[
          const SizedBox(height: 24),
          Text(l10n.importing),
          if (state.importTotal > 0) ...[
            const SizedBox(height: 4),
            Text('${state.importCurrent} / ${state.importTotal}'),
          ],
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: state.importProgress > 0 ? state.importProgress : null,
          ),
        ],
        if (state.completed) ...[
          const SizedBox(height: 24),
          Text(
            l10n.stravaImportCompleted(
              state.resultImported,
              state.resultSkipped,
              state.resultParseSkipped,
              state.resultMediaMissing,
              state.resultErrors,
            ),
            style: theme.textTheme.bodyLarge?.copyWith(
              color: theme.colorScheme.primary,
            ),
          ),
        ],
        if (state.failed && state.message.isNotEmpty) ...[
          const SizedBox(height: 24),
          Text(
            l10n.stravaImportFailed(state.message),
            style: theme.textTheme.bodyLarge?.copyWith(
              color: theme.colorScheme.error,
            ),
          ),
        ],
      ],
    );
  }
}
