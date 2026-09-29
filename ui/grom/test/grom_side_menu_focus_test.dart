import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/navigation/grom_destination.dart';
import 'package:grom/navigation/grom_side_menu.dart';

void main() {
  testWidgets('Tab skips side menu and stays in content fields', (tester) async {
    final firstFocus = FocusNode();
    final secondFocus = FocusNode();
    addTearDown(firstFocus.dispose);
    addTearDown(secondFocus.dispose);

    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: Row(
            children: [
              const SizedBox(
                width: kSideMenuWidth,
                child: GromSideMenu(
                  selectedDestination: GromDestination.login,
                  onDestinationSelected: _noopDestination,
                  serverTitle: 'Grom',
                  isLoggedIn: false,
                  onLogout: _noop,
                ),
              ),
              Expanded(
                child: Column(
                  children: [
                    TextField(
                      focusNode: firstFocus,
                      decoration: const InputDecoration(labelText: 'First'),
                    ),
                    TextField(
                      focusNode: secondFocus,
                      decoration: const InputDecoration(labelText: 'Second'),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byType(ExcludeFocusTraversal), findsOneWidget);

    firstFocus.requestFocus();
    await tester.pump();
    expect(firstFocus.hasFocus, isTrue);

    await tester.sendKeyEvent(LogicalKeyboardKey.tab);
    await tester.pump();

    expect(secondFocus.hasFocus, isTrue);
    expect(firstFocus.hasFocus, isFalse);

    final focusedContext = FocusManager.instance.primaryFocus?.context;
    expect(focusedContext, isNotNull);
    expect(
      focusedContext!.findAncestorWidgetOfExactType<GromSideMenu>(),
      isNull,
    );
  });
}

void _noop() {}

void _noopDestination(GromDestination _) {}
