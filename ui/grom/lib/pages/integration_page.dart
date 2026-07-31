import 'package:flutter/foundation.dart' show kIsWeb, defaultTargetPlatform, TargetPlatform;
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:material_symbols_icons/symbols.dart';
import 'package:url_launcher/url_launcher.dart';

import '../services/health_sync_service.dart';
import '../services/strava_import_service.dart';
import 'about_page_url.dart' if (dart.library.html) 'about_page_url_web.dart';

class IntegrationPage extends StatefulWidget {
  const IntegrationPage({super.key});

  @override
  State<IntegrationPage> createState() => _IntegrationPageState();
}

class _IntegrationPageState extends State<IntegrationPage> {
  final StravaImportService _importService = StravaImportService.instance;
  final HealthSyncService _healthSyncService = HealthSyncService.instance;
  late final TapGestureRecognizer _stravaDownloadLinkRecognizer;
  late final TapGestureRecognizer _healthSyncLinkRecognizer;
  late final TextEditingController _folderNameController;

  bool get _isAndroid =>
      !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

  @override
  void initState() {
    super.initState();
    _folderNameController = TextEditingController();
    _stravaDownloadLinkRecognizer = TapGestureRecognizer()
      ..onTap = _openStravaDownloadUrl;
    _healthSyncLinkRecognizer = TapGestureRecognizer()
      ..onTap = _openHealthSyncPlayStoreUrl;
    _importService.addListener(_onImportChanged);
    _healthSyncService.addListener(_onHealthSyncChanged);
    _importService.syncFromServer();
    _loadHealthSyncState();
  }

  Future<void> _loadHealthSyncState() async {
    await _healthSyncService.loadFromStorage();
    if (!mounted) {
      return;
    }
    _folderNameController.text = _healthSyncService.folderName;
    setState(() {});
  }

  @override
  void dispose() {
    _stravaDownloadLinkRecognizer.dispose();
    _healthSyncLinkRecognizer.dispose();
    _folderNameController.dispose();
    _importService.removeListener(_onImportChanged);
    _healthSyncService.removeListener(_onHealthSyncChanged);
    super.dispose();
  }

  void _onImportChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  void _onHealthSyncChanged() {
    if (!mounted) {
      return;
    }
    if (_folderNameController.text != _healthSyncService.folderName) {
      _folderNameController.text = _healthSyncService.folderName;
    }
    setState(() {});
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

  Future<void> _openHealthSyncPlayStoreUrl() async {
    final l10n = AppLocalizations.of(context)!;
    final uri = Uri.parse(l10n.healthSyncPlayStoreUrl);
    if (kIsWeb) {
      openCopyrightUrlInBrowser(uri.toString());
      return;
    }

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

  void _showHealthSyncResultSnackBar(HealthSyncResult result) {
    final l10n = AppLocalizations.of(context)!;
    final messenger = ScaffoldMessenger.of(context);
    final message = switch (result.kind) {
      HealthSyncResultKind.imported => l10n.healthSyncImported(result.importedCount),
      HealthSyncResultKind.noNewWorkouts => l10n.healthSyncNoNewWorkouts,
      HealthSyncResultKind.folderNotFound => l10n.healthSyncFolderNotFound,
      HealthSyncResultKind.folderEmpty => l10n.healthSyncFolderEmpty,
      HealthSyncResultKind.signInCancelled => l10n.healthSyncGoogleSignInCancelled,
      HealthSyncResultKind.signInFailed => l10n.healthSyncGoogleSignInFailed,
      HealthSyncResultKind.accessDenied => l10n.healthSyncDriveAccessDenied,
      HealthSyncResultKind.error => l10n.healthSyncSyncError(result.message),
    };
    messenger.showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _onHealthSyncToggleChanged(bool enabled) async {
    if (!enabled) {
      await _healthSyncService.setEnabled(false);
      return;
    }

    final result = await _healthSyncService.enableAndSetup();
    if (!mounted) {
      return;
    }

    if (!_healthSyncService.enabled) {
      await _healthSyncService.setEnabled(false);
      _showHealthSyncResultSnackBar(result);
      return;
    }

    _folderNameController.text = _healthSyncService.folderName;
    if (result.kind != HealthSyncResultKind.noNewWorkouts) {
      _showHealthSyncResultSnackBar(result);
    }
  }

  Future<void> _refreshHealthSyncFolder() async {
    final result = await _healthSyncService.refreshFolderFromDrive();
    if (!mounted) {
      return;
    }
    _folderNameController.text = _healthSyncService.folderName;
    if (result.kind != HealthSyncResultKind.noNewWorkouts) {
      _showHealthSyncResultSnackBar(result);
    }
  }

  Future<void> _onFolderNameChanged(String value) async {
    await _healthSyncService.updateFolderName(value);
  }

  Widget _buildHealthSyncSection(AppLocalizations l10n, ThemeData theme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          l10n.healthSyncGoogleDrive,
          style: theme.textTheme.titleLarge,
        ),
        const SizedBox(height: 16),
        Text.rich(
          TextSpan(
            style: theme.textTheme.bodyMedium,
            children: [
              TextSpan(text: l10n.healthSyncImportDescriptionBefore),
              TextSpan(
                text: l10n.healthSyncImportDescriptionLink,
                style: TextStyle(
                  color: theme.colorScheme.primary,
                  decoration: TextDecoration.underline,
                ),
                recognizer: _healthSyncLinkRecognizer,
              ),
              TextSpan(text: l10n.healthSyncImportDescriptionAfter),
            ],
          ),
        ),
        const SizedBox(height: 16),
        SwitchListTile(
          contentPadding: EdgeInsets.zero,
          title: Text(l10n.healthSyncSyncToggle),
          value: _healthSyncService.enabled,
          onChanged: _healthSyncService.syncing ? null : _onHealthSyncToggleChanged,
        ),
        if (_healthSyncService.enabled) ...[
          const SizedBox(height: 8),
          Text(l10n.healthSyncFolderLabel, style: theme.textTheme.titleSmall),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _folderNameController,
                  decoration: InputDecoration(
                    border: const OutlineInputBorder(),
                    hintText: l10n.healthSyncFolderLabel,
                  ),
                  onChanged: _onFolderNameChanged,
                ),
              ),
              IconButton(
                tooltip: l10n.healthSyncFindFolder,
                onPressed: _healthSyncService.syncing ? null : _refreshHealthSyncFolder,
                icon: const Icon(Symbols.folder_eye),
              ),
            ],
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

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        if (_isAndroid) _buildHealthSyncSection(l10n, theme),
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
