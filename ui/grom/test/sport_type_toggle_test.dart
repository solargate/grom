import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/widgets/sport_type_toggle.dart';

void main() {
  Widget wrap(Widget child) {
    return MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(body: child),
    );
  }

  testWidgets('SportTypeToggleButton toggles selection', (tester) async {
    var selected = true;
    await tester.pumpWidget(
      wrap(
        SportTypeToggleButton(
          sportTypeId: 'Run',
          selected: selected,
          onSelected: (value) => selected = value,
        ),
      ),
    );

    expect(find.text('Run'), findsOneWidget);
    await tester.tap(find.text('Run'));
    expect(selected, isFalse);
  });

  testWidgets('SportTypeToggleWrap renders toggles for ids', (tester) async {
    final selected = {'Run', 'Ride'};
    await tester.pumpWidget(
      wrap(
        SportTypeToggleWrap(
          sportTypeIds: const ['Run', 'Ride', 'Walk'],
          selectedIds: selected,
          onToggle: (_) {},
        ),
      ),
    );

    expect(find.text('Run'), findsOneWidget);
    expect(find.text('Ride'), findsOneWidget);
    expect(find.text('Walk'), findsOneWidget);
  });

  testWidgets('SportTypeToggleWrap hides when ids empty', (tester) async {
    await tester.pumpWidget(
      wrap(
        SportTypeToggleWrap(
          sportTypeIds: const [],
          selectedIds: const {},
          onToggle: (_) {},
        ),
      ),
    );
    expect(find.byType(SportTypeToggleButton), findsNothing);
  });
}
