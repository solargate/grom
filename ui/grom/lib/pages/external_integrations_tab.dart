import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

import '../services/device_track_import_service.dart';
import '../services/strava_import_service.dart';
import 'about_page_url.dart' if (dart.library.html) 'about_page_url_web.dart';

class ExternalIntegrationsTab extends StatefulWidget {
  const ExternalIntegrationsTab({super.key, this.onWorkoutsImported});

  /// Called when device track import created at least one workout.
  final VoidCallback? onWorkoutsImported;

  @override
  State<ExternalIntegrationsTab> createState() => _ExternalIntegrationsTabState();
}

class _ExternalIntegrationsTabState extends State<ExternalIntegrationsTab> {
  final StravaImportService _importService = StravaImportService.instance;
  final DeviceTrackImportService _trackImportService =
      DeviceTrackImportService.instance;
  late final TapGestureRecognizer _stravaDownloadLinkRecognizer;

  @override
  void initState() {
    super.initState();
    _stravaDownloadLinkRecognizer = TapGestureRecognizer()
      ..onTap = _openStravaDownloadUrl;
    _importService.addListener(_onImportChanged);
    _trackImportService.addListener(_onTrackImportChanged);
    _importService.syncFromServer();
  }

  @override
  void dispose() {
    _stravaDownloadLinkRecognizer.dispose();
    _importService.removeListener(_onImportChanged);
    _trackImportService.removeListener(_onTrackImportChanged);
    super.dispose();
  }

  void _onImportChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  void _onTrackImportChanged() {
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

  Future<void> _runTrackImport() async {
    final result = await _trackImportService.pickAndImport();
    if (!mounted || !result.completed) {
      return;
    }

    final l10n = AppLocalizations.of(context)!;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          l10n.importTracksResult(
            result.created,
            result.skipped,
            result.invalid,
            result.failed,
          ),
        ),
      ),
    );
    if (result.created > 0) {
      widget.onWorkoutsImported?.call();
    }
  }

  Widget _buildImportTracksSection(AppLocalizations l10n, ThemeData theme) {
    final state = _trackImportService.state;
    final busy = state.active || _importService.state.active;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          l10n.importTracksTitle,
          style: theme.textTheme.titleLarge,
        ),
        const SizedBox(height: 16),
        Text(
          l10n.importTracksDescription,
          style: theme.textTheme.bodyMedium,
        ),
        const SizedBox(height: 16),
        FilledButton.icon(
          onPressed: busy ? null : _runTrackImport,
          icon: const Icon(Icons.upload_file),
          label: Text(l10n.importTracksButton),
        ),
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
        const SizedBox(height: 32),
        const Divider(),
        const SizedBox(height: 32),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final state = _importService.state;
    final trackBusy = _trackImportService.state.active;

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _buildImportTracksSection(l10n, theme),
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
                text: l10n.stravaImportDescriptionLink,
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
          onPressed: (state.active || trackBusy)
              ? null
              : _importService.pickAndImport,
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
