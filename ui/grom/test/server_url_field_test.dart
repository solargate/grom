import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/server_catalog.dart';
import 'package:grom/server_history.dart';
import 'package:grom/widgets/server_url_field.dart';
import 'package:shared_preferences/shared_preferences.dart';

Widget wrap(Widget child) {
  return MaterialApp(
    theme: buildAppTheme(),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    locale: const Locale('en'),
    home: Scaffold(body: child),
  );
}

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ServerHistory.clear();
  });

  testWidgets('dropdown fills the field from an approved server',
      (tester) async {
    final controller = TextEditingController();
    addTearDown(controller.dispose);

    await tester.pumpWidget(
      wrap(
        ServerUrlField(
          controller: controller,
          approvedServers: const [
            CatalogServer(
              url: 'https://approved.example',
              name: 'Approved Grom',
              description: 'A public test instance',
            ),
          ],
        ),
      ),
    );

    await tester.tap(find.byTooltip('Choose a server'));
    await tester.pumpAndSettle();

    expect(find.text('https://approved.example'), findsOneWidget);
    expect(find.text('Approved Grom'), findsOneWidget);
    expect(find.text('A public test instance'), findsOneWidget);

    await tester.tap(find.text('https://approved.example'));
    await tester.pumpAndSettle();

    expect(controller.text, 'https://approved.example');
    expect(find.text('Approved Grom'), findsNothing);
  });

  testWidgets('dropdown lists recent custom servers after approved',
      (tester) async {
    await ServerHistory.remember('https://custom.example');
    final controller = TextEditingController();
    addTearDown(controller.dispose);

    await tester.pumpWidget(
      wrap(
        ServerUrlField(
          controller: controller,
          approvedServers: const [
            CatalogServer(
              url: 'https://approved.example',
              name: 'Approved Grom',
              description: 'Listed',
            ),
          ],
        ),
      ),
    );
    await tester.pump();

    await tester.tap(find.byTooltip('Choose a server'));
    await tester.pumpAndSettle();

    expect(find.text('Approved servers'), findsOneWidget);
    expect(find.text('Recent servers'), findsOneWidget);
    expect(find.text('https://custom.example'), findsOneWidget);

    await tester.tap(find.text('https://custom.example'));
    await tester.pumpAndSettle();
    expect(controller.text, 'https://custom.example');
  });

  testWidgets('empty picker shows a hint', (tester) async {
    final controller = TextEditingController();
    addTearDown(controller.dispose);

    await tester.pumpWidget(
      wrap(
        ServerUrlField(
          controller: controller,
          approvedServers: const [],
        ),
      ),
    );
    await tester.tap(find.byTooltip('Choose a server'));
    await tester.pumpAndSettle();

    expect(
      find.text('No servers yet. Enter a URL, or sign in to remember one.'),
      findsOneWidget,
    );
  });
}
