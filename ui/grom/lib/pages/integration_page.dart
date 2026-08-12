import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import 'external_integrations_tab.dart';
import 'grom_api_tab.dart';

class IntegrationPage extends StatelessWidget {
  const IntegrationPage({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return DefaultTabController(
      length: 2,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          TabBar(
            tabs: [
              Tab(text: l10n.integrationTabGrom),
              Tab(text: l10n.integrationTabExternal),
            ],
          ),
          const Expanded(
            child: TabBarView(
              children: [
                GromApiTab(),
                ExternalIntegrationsTab(),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
