import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

class HomePage extends StatelessWidget {
  const HomePage({
    super.key,
    this.nickname,
  });

  final String? nickname;

  @override
  Widget build(BuildContext context) {
    if (nickname == null) {
      return const SizedBox.shrink();
    }

    final l10n = AppLocalizations.of(context)!;

    return Center(
      child: Text(
        l10n.signedInAs(nickname!),
        style: Theme.of(context).textTheme.bodyLarge,
      ),
    );
  }
}
