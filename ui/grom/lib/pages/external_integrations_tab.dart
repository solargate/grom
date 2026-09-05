import 'package:flutter/foundation.dart'
    show TargetPlatform, defaultTargetPlatform, kIsWeb;
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

import '../services/device_track_import_service.dart';
import '../services/strava_api_auth.dart';
import '../services/strava_api_constants.dart';
import '../services/strava_api_sync_service.dart';
import '../services/strava_import_service.dart';
import 'about_page_url.dart' if (dart.library.html) 'about_page_url_web.dart';

class ExternalIntegrationsTab extends StatefulWidget {
  const ExternalIntegrationsTab({super.key, this.onWorkoutsImported});

  /// Called when device track import created at least one workout.
  final VoidCallback? onWorkoutsImported;

  @override
  State<ExternalIntegrationsTab> createState() =>
      _ExternalIntegrationsTabState();
}

class _ExternalIntegrationsTabState extends State<ExternalIntegrationsTab> {
  final StravaImportService _importService = StravaImportService.instance;
  final DeviceTrackImportService _trackImportService =
      DeviceTrackImportService.instance;
  final StravaApiSyncService _stravaApi = StravaApiSyncService.instance;
  late final TapGestureRecognizer _stravaDownloadLinkRecognizer;
  late final TextEditingController _clientIdController;
  late final TextEditingController _clientSecretController;

  bool _stravaConnectFailed = false;
  bool _connecting = false;

  bool get _isAndroid =>
      !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

  @override
  void initState() {
    super.initState();
    _clientIdController = TextEditingController();
    _clientSecretController = TextEditingController();
    _stravaDownloadLinkRecognizer = TapGestureRecognizer()
      ..onTap = _openStravaDownloadUrl;
    _importService.addListener(_onImportChanged);
    _trackImportService.addListener(_onTrackImportChanged);
    _stravaApi.addListener(_onStravaApiChanged);
    _importService.syncFromServer();
    _loadStravaApiState();
  }

  Future<void> _loadStravaApiState() async {
    await _stravaApi.loadFromStorage();
    if (!mounted) {
      return;
    }
    _clientIdController.text = _stravaApi.clientId;
    _clientSecretController.text = _stravaApi.clientSecret;
    setState(() {});
  }

  @override
  void dispose() {
    _stravaDownloadLinkRecognizer.dispose();
    _clientIdController.dispose();
    _clientSecretController.dispose();
    _importService.removeListener(_onImportChanged);
    _trackImportService.removeListener(_onTrackImportChanged);
    _stravaApi.removeListener(_onStravaApiChanged);
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

  void _onStravaApiChanged() {
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

  Future<void> _onStravaApiToggle(bool enabled) async {
    await _stravaApi.setEnabled(enabled);
    if (!enabled) {
      setState(() => _stravaConnectFailed = false);
    }
  }

  Future<void> _persistCredentialsDraft() async {
    await _stravaApi.saveCredentialsDraft(
      clientId: _clientIdController.text,
      clientSecret: _clientSecretController.text,
    );
  }

  Future<void> _connectStrava() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      _connecting = true;
      _stravaConnectFailed = false;
    });
    await _persistCredentialsDraft();
    final result = await _stravaApi.connect();
    if (!mounted) {
      return;
    }
    setState(() {
      _connecting = false;
      _stravaConnectFailed = result.kind != StravaConnectResultKind.connected &&
          result.kind != StravaConnectResultKind.cancelled;
    });

    final message = switch (result.kind) {
      StravaConnectResultKind.connected => l10n.stravaApiConnectStatusConnected,
      StravaConnectResultKind.cancelled => l10n.stravaApiConnectCancelled,
      StravaConnectResultKind.missingCredentials =>
        l10n.stravaApiConnectMissingCredentials,
      StravaConnectResultKind.missingScope =>
        l10n.stravaApiConnectMissingScope,
      StravaConnectResultKind.denied => l10n.stravaApiConnectDenied,
      StravaConnectResultKind.error =>
        l10n.stravaApiConnectError(result.message),
    };
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  Widget _buildImportTracksSection(AppLocalizations l10n, ThemeData theme) {
    final state = _trackImportService.state;
    final busy = state.active || _importService.state.active || _stravaApi.syncing;

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

  Widget _buildStravaApiSection(AppLocalizations l10n, ThemeData theme) {
    final busy = _connecting || _stravaApi.syncing || _importService.state.active;

    final Color statusColor;
    final String statusText;
    if (_stravaApi.connected) {
      statusColor = theme.colorScheme.primary;
      statusText = l10n.stravaApiConnectStatusConnected;
    } else if (_stravaConnectFailed) {
      statusColor = theme.colorScheme.error;
      statusText = l10n.stravaApiConnectStatusFailed;
    } else {
      statusColor = theme.colorScheme.onSurfaceVariant;
      statusText = l10n.stravaApiConnectStatusDisconnected;
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        SwitchListTile(
          contentPadding: EdgeInsets.zero,
          title: Text(l10n.stravaApiImportToggle),
          value: _stravaApi.enabled,
          onChanged: busy ? null : _onStravaApiToggle,
        ),
        if (_stravaApi.enabled) ...[
          const SizedBox(height: 8),
          TextField(
            controller: _clientIdController,
            enabled: !busy,
            decoration: InputDecoration(
              labelText: l10n.stravaApiClientIdLabel,
              border: const OutlineInputBorder(),
            ),
            keyboardType: TextInputType.number,
            onChanged: (_) => _persistCredentialsDraft(),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _clientSecretController,
            enabled: !busy,
            obscureText: true,
            decoration: InputDecoration(
              labelText: l10n.stravaApiClientSecretLabel,
              border: const OutlineInputBorder(),
            ),
            onChanged: (_) => _persistCredentialsDraft(),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Icon(Icons.circle, size: 12, color: statusColor),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  statusText,
                  style: theme.textTheme.bodyMedium?.copyWith(color: statusColor),
                ),
              ),
              const SizedBox(width: 8),
              InkWell(
                onTap: busy ? null : _connectStrava,
                borderRadius: BorderRadius.circular(4),
                child: Opacity(
                  opacity: busy ? 0.5 : 1,
                  child: SvgPicture.asset(
                    kStravaConnectButtonAsset,
                    height: 40,
                    semanticsLabel: 'Connect with Strava',
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),
        ],
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
        if (_isAndroid) _buildStravaApiSection(l10n, theme),
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
          onPressed: (state.active || trackBusy || _stravaApi.syncing)
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
