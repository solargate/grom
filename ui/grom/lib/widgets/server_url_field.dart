import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../server_storage.dart';

class ServerUrlField extends StatelessWidget {
  const ServerUrlField({
    super.key,
    required this.controller,
  });

  final TextEditingController controller;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        labelText: l10n.serverUrlLabel,
        border: const OutlineInputBorder(),
        hintText: 'https://example.com',
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
