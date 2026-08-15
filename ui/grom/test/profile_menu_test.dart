import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/widgets/profile_menu.dart';

Widget wrap(Widget child) {
  return MaterialApp(
    theme: buildAppTheme(),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(
      appBar: AppBar(actions: [child]),
    ),
  );
}

void main() {
  testWidgets('ProfileMenu exposes edit and delete account actions', (tester) async {
    ProfileMenuAction? selected;

    await tester.pumpWidget(
      wrap(
        ProfileMenu(
          onSelected: (action) => selected = action,
        ),
      ),
    );

    await tester.tap(find.byType(PopupMenuButton<ProfileMenuAction>));
    await tester.pumpAndSettle();

    expect(find.text('Edit profile'), findsOneWidget);
    expect(find.text('Delete account'), findsOneWidget);

    await tester.tap(find.text('Delete account'));
    await tester.pumpAndSettle();
    expect(selected, ProfileMenuAction.deleteAccount);
  });
}
