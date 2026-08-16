import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/widgets/welcome_guest_view.dart';

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
  testWidgets('WelcomeGuestView shows branding copy and auth actions',
      (tester) async {
    var signedIn = false;
    var registered = false;

    await tester.pumpWidget(
      wrap(
        WelcomeGuestView(
          onSignIn: () => signedIn = true,
          onRegister: () => registered = true,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Grom'), findsWidgets);
    expect(
      find.text('Workouts, equipment, and a friends feed on your own server.'),
      findsOneWidget,
    );
    expect(find.text('To get started, sign in or register.'), findsOneWidget);
    expect(
      find.text('On a mobile phone, enter the Grom server address.'),
      findsNothing,
    );

    await tester.tap(find.text('Sign in'));
    await tester.pump();
    expect(signedIn, isTrue);

    await tester.tap(find.text('Register'));
    await tester.pump();
    expect(registered, isTrue);
  });

  testWidgets('WelcomeGuestView shows mobile server hint when requested',
      (tester) async {
    await tester.pumpWidget(
      wrap(
        WelcomeGuestView(
          onSignIn: () {},
          onRegister: () {},
          showMobileServerHint: true,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      find.text('On a mobile phone, enter the Grom server address.'),
      findsOneWidget,
    );
  });
}
