import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/personal_access_token.dart';
import 'package:intl/intl.dart';

import '../api_request.dart';
import '../auth_storage.dart';

enum _PatExpiryMode { days90, days180, custom, none }

class GromApiTab extends StatefulWidget {
  const GromApiTab({super.key});

  @override
  State<GromApiTab> createState() => _GromApiTabState();
}

class _GromApiTabState extends State<GromApiTab> {
  final ApiRequest _api = ApiRequest();

  List<PersonalAccessToken> _tokens = [];
  bool _isLoading = true;
  String? _error;

  static const _scopeValues = [
    'workouts:read',
    'workouts:write',
    'equipment:read',
    'equipment:write',
  ];

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }
      final items = await _api.listPersonalAccessTokens(token);
      if (!mounted) return;
      setState(() {
        _tokens = items;
        _isLoading = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _error = l10n.patFailedToLoad;
        _isLoading = false;
      });
    }
  }

  String _scopeLabel(AppLocalizations l10n, String scope) {
    switch (scope) {
      case 'workouts:read':
        return l10n.patScopeWorkoutsRead;
      case 'workouts:write':
        return l10n.patScopeWorkoutsWrite;
      case 'equipment:read':
        return l10n.patScopeEquipmentRead;
      case 'equipment:write':
        return l10n.patScopeEquipmentWrite;
      default:
        return scope;
    }
  }

  String _formatDate(DateTime date) {
    return DateFormat.yMMMd().format(date.toLocal());
  }

  Future<void> _showCreateDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final nameController = TextEditingController();
    final customDaysController = TextEditingController(text: '90');
    final selectedScopes = <String>{};
    var expiryMode = _PatExpiryMode.days90;

    final created = await showDialog<CreatePersonalAccessTokenResult?>(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            return AlertDialog(
              title: Text(l10n.patCreateToken),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    TextField(
                      controller: nameController,
                      decoration: InputDecoration(
                        labelText: l10n.patNameLabel,
                        border: const OutlineInputBorder(),
                      ),
                      textInputAction: TextInputAction.done,
                    ),
                    const SizedBox(height: 16),
                    Text(l10n.patScopesLabel, style: Theme.of(context).textTheme.titleSmall),
                    for (final scope in _scopeValues)
                      CheckboxListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text(_scopeLabel(l10n, scope)),
                        value: selectedScopes.contains(scope),
                        onChanged: (checked) {
                          setDialogState(() {
                            if (checked == true) {
                              selectedScopes.add(scope);
                            } else {
                              selectedScopes.remove(scope);
                            }
                          });
                        },
                      ),
                    const SizedBox(height: 8),
                    Text(l10n.patExpiryLabel, style: Theme.of(context).textTheme.titleSmall),
                    RadioListTile<_PatExpiryMode>(
                      contentPadding: EdgeInsets.zero,
                      title: Text(l10n.patExpiry90Days),
                      value: _PatExpiryMode.days90,
                      groupValue: expiryMode,
                      onChanged: (value) => setDialogState(() => expiryMode = value!),
                    ),
                    RadioListTile<_PatExpiryMode>(
                      contentPadding: EdgeInsets.zero,
                      title: Text(l10n.patExpiry180Days),
                      value: _PatExpiryMode.days180,
                      groupValue: expiryMode,
                      onChanged: (value) => setDialogState(() => expiryMode = value!),
                    ),
                    RadioListTile<_PatExpiryMode>(
                      contentPadding: EdgeInsets.zero,
                      title: Text(l10n.patExpiryCustomDays),
                      value: _PatExpiryMode.custom,
                      groupValue: expiryMode,
                      onChanged: (value) => setDialogState(() => expiryMode = value!),
                    ),
                    if (expiryMode == _PatExpiryMode.custom)
                      Padding(
                        padding: const EdgeInsets.only(left: 16, bottom: 8),
                        child: TextField(
                          controller: customDaysController,
                          keyboardType: TextInputType.number,
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                          ),
                        ),
                      ),
                    RadioListTile<_PatExpiryMode>(
                      contentPadding: EdgeInsets.zero,
                      title: Text(l10n.patExpiryNone),
                      value: _PatExpiryMode.none,
                      groupValue: expiryMode,
                      onChanged: (value) => setDialogState(() => expiryMode = value!),
                    ),
                    if (expiryMode == _PatExpiryMode.none)
                      Text(
                        l10n.patNoExpiryWarning,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Theme.of(context).colorScheme.error,
                            ),
                      ),
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: Text(l10n.cancel),
                ),
                FilledButton(
                  onPressed: () async {
                    if (nameController.text.trim().isEmpty) {
                      return;
                    }
                    if (selectedScopes.isEmpty) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(content: Text(l10n.patSelectScope)),
                      );
                      return;
                    }
                    final session = await AuthStorage.getToken();
                    if (session == null) {
                      return;
                    }
                    int? expiresInDays;
                    var noExpiration = false;
                    switch (expiryMode) {
                      case _PatExpiryMode.days90:
                        expiresInDays = 90;
                      case _PatExpiryMode.days180:
                        expiresInDays = 180;
                      case _PatExpiryMode.custom:
                        expiresInDays = int.tryParse(customDaysController.text.trim());
                        if (expiresInDays == null || expiresInDays < 1) {
                          return;
                        }
                      case _PatExpiryMode.none:
                        noExpiration = true;
                    }
                    try {
                      final result = await _api.createPersonalAccessToken(
                        token: session,
                        name: nameController.text.trim(),
                        scopes: selectedScopes.toList(),
                        expiresInDays: expiresInDays,
                        noExpiration: noExpiration,
                      );
                      if (context.mounted) {
                        Navigator.pop(context, result);
                      }
                    } on ApiException catch (e) {
                      if (context.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text(e.message)),
                        );
                      }
                    } catch (_) {
                      if (context.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text(l10n.patFailedToCreate)),
                        );
                      }
                    }
                  },
                  child: Text(l10n.patCreateToken),
                ),
              ],
            );
          },
        );
      },
    );

    nameController.dispose();
    customDaysController.dispose();

    if (!mounted || created == null) {
      return;
    }

    await _showTokenRevealDialog(created.token);
    await _load();
  }

  Future<void> _showTokenRevealDialog(String token) async {
    final l10n = AppLocalizations.of(context)!;
    await showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (context) {
        return AlertDialog(
          title: Text(l10n.patTokenCreatedTitle),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                l10n.patTokenCreatedWarning,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
              const SizedBox(height: 16),
              SelectableText(
                token,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      fontFamily: 'monospace',
                    ),
              ),
            ],
          ),
          actions: [
            FilledButton(
              onPressed: () async {
                await Clipboard.setData(ClipboardData(text: token));
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text(l10n.patTokenCopied)),
                  );
                }
              },
              child: Text(l10n.patCopyToken),
            ),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(l10n.patClose),
            ),
          ],
        );
      },
    );
  }

  Future<void> _confirmRevoke(PersonalAccessToken item) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.patRevokeConfirmTitle),
        content: Text(l10n.patRevokeConfirmMessage(item.name)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.patRevoke),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) {
      return;
    }
    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        return;
      }
      await _api.revokePersonalAccessToken(token: token, id: item.id);
      await _load();
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.patFailedToRevoke)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(_error!, textAlign: TextAlign.center),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _load,
                child: Text(l10n.retry),
              ),
            ],
          ),
        ),
      );
    }

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text(l10n.gromApiTitle, style: theme.textTheme.titleLarge),
        const SizedBox(height: 16),
        Text(l10n.gromApiDescription, style: theme.textTheme.bodyMedium),
        const SizedBox(height: 24),
        FilledButton.icon(
          onPressed: _showCreateDialog,
          icon: const Icon(Icons.add),
          label: Text(l10n.patCreateToken),
        ),
        const SizedBox(height: 24),
        if (_tokens.isEmpty)
          Text(l10n.patNoTokens, style: theme.textTheme.bodyLarge)
        else
          ..._tokens.map((item) {
            final scopeText =
                item.scopes.map((scope) => _scopeLabel(l10n, scope)).join(', ');
            final expiresText = item.expiresAt == null
                ? l10n.patExpiresNever
                : l10n.patExpiresAt(_formatDate(item.expiresAt!));
            final lastUsedText = item.lastUsedAt == null
                ? l10n.patLastUsedNever
                : l10n.patLastUsedAt(_formatDate(item.lastUsedAt!));
            return Card(
              margin: const EdgeInsets.only(bottom: 12),
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(item.name, style: theme.textTheme.titleMedium),
                    if (item.tokenPrefix.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        item.tokenPrefix,
                        style: theme.textTheme.bodySmall?.copyWith(
                          fontFamily: 'monospace',
                        ),
                      ),
                    ],
                    const SizedBox(height: 8),
                    Text(scopeText, style: theme.textTheme.bodyMedium),
                    const SizedBox(height: 4),
                    Text(
                      l10n.patCreatedAt(_formatDate(item.createdAt)),
                      style: theme.textTheme.bodySmall,
                    ),
                    Text(expiresText, style: theme.textTheme.bodySmall),
                    Text(lastUsedText, style: theme.textTheme.bodySmall),
                    const SizedBox(height: 12),
                    Align(
                      alignment: Alignment.centerRight,
                      child: TextButton(
                        onPressed: () => _confirmRevoke(item),
                        child: Text(l10n.patRevoke),
                      ),
                    ),
                  ],
                ),
              ),
            );
          }),
      ],
    );
  }
}
