import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../server_catalog.dart';
import '../server_history.dart';
import '../server_storage.dart';

class ServerUrlField extends StatefulWidget {
  const ServerUrlField({
    super.key,
    required this.controller,
    this.approvedServers = kApprovedServers,
  });

  final TextEditingController controller;
  final List<CatalogServer> approvedServers;

  @override
  State<ServerUrlField> createState() => _ServerUrlFieldState();
}

class _ServerUrlFieldState extends State<ServerUrlField> {
  List<String> _recent = const [];

  @override
  void initState() {
    super.initState();
    _loadRecent();
  }

  Future<void> _loadRecent() async {
    final recent = await ServerHistory.getRecent();
    if (!mounted) {
      return;
    }
    setState(() => _recent = recent);
  }

  Future<void> _openPicker() async {
    final l10n = AppLocalizations.of(context)!;
    await _loadRecent();
    if (!mounted) {
      return;
    }

    final approved = widget.approvedServers;
    final recent = recentServersNotInCatalog(_recent, approved);
    final theme = Theme.of(context);
    final smallStyle = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );

    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      builder: (sheetContext) {
        return SafeArea(
          child: ListView(
            shrinkWrap: true,
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: Text(
                  l10n.chooseServerTitle,
                  style: theme.textTheme.titleMedium,
                ),
              ),
              if (approved.isEmpty && recent.isEmpty)
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
                  child: Text(l10n.serverPickerEmpty),
                ),
              if (approved.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
                  child: Text(
                    l10n.approvedServersSection,
                    style: theme.textTheme.labelLarge,
                  ),
                ),
              for (final server in approved)
                ListTile(
                  title: Text(
                    server.url,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                  subtitle: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(server.name),
                      Text(server.description, style: smallStyle),
                    ],
                  ),
                  isThreeLine: true,
                  onTap: () {
                    widget.controller.text = server.url;
                    Navigator.pop(sheetContext);
                  },
                ),
              if (recent.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
                  child: Text(
                    l10n.recentServersSection,
                    style: theme.textTheme.labelLarge,
                  ),
                ),
              for (final url in recent)
                ListTile(
                  title: Text(
                    url,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                  onTap: () {
                    widget.controller.text = url;
                    Navigator.pop(sheetContext);
                  },
                ),
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return TextFormField(
      controller: widget.controller,
      decoration: InputDecoration(
        labelText: l10n.serverUrlLabel,
        border: const OutlineInputBorder(),
        hintText: l10n.serverUrlHint,
        suffixIcon: IconButton(
          icon: const Icon(Icons.arrow_drop_down),
          tooltip: l10n.chooseServerTooltip,
          onPressed: _openPicker,
        ),
      ),
      keyboardType: TextInputType.url,
      textInputAction: TextInputAction.next,
      autocorrect: false,
      validator: (value) {
        if (value == null || value.trim().isEmpty) {
          return l10n.enterServerUrl;
        }
        if (!ServerStorage.isValidBaseUrl(value)) {
          return l10n.enterValidServerUrl;
        }
        return null;
      },
    );
  }
}
