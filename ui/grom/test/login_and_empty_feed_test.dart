import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/login.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/widgets/altcha_field.dart';
import 'package:grom/widgets/workout_feed_list.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
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

Finder labeledField(String label) {
  return find.ancestor(
    of: find.text(label),
    matching: find.byType(TextFormField),
  );
}

MockClient serverInfoClient({
  bool passwordResetEnabled = false,
  bool captchaEnabled = false,
}) {
  return MockClient((request) async {
    if (request.url.path == '/api/v1/server-info') {
      return http.Response(
        jsonEncode({
          'name': 'Test',
          'password_reset_enabled': passwordResetEnabled,
          'captcha_enabled': captchaEnabled,
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    return http.Response('not found', 404);
  });
}

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await AuthStorage.clear();
    await ServerStorage.clear();
    await ServerStorage.saveBaseUrl('https://grom.example');
  });

  testWidgets('LoginForm shows email password and sign-in controls', (tester) async {
    await tester.pumpWidget(
      wrap(LoginForm(api: ApiRequest(client: serverInfoClient()))),
    );
    await tester.pumpAndSettle();

    expect(find.text('Email *'), findsOneWidget);
    expect(find.text('Password *'), findsOneWidget);
    expect(find.text('Sign in'), findsOneWidget);
  });

  testWidgets('LoginForm shows forgot password when server enables reset', (tester) async {
    await tester.pumpWidget(
      wrap(
        LoginForm(
          api: ApiRequest(client: serverInfoClient(passwordResetEnabled: true)),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Forgot password?'), findsOneWidget);
  });

  testWidgets('LoginForm hides forgot password when reset disabled', (tester) async {
    await tester.pumpWidget(
      wrap(LoginForm(api: ApiRequest(client: serverInfoClient()))),
    );
    await tester.pumpAndSettle();

    expect(find.text('Forgot password?'), findsNothing);
  });

  testWidgets('LoginForm shows captcha snackbar when required and missing', (tester) async {
    await tester.pumpWidget(
      wrap(
        LoginForm(
          api: ApiRequest(client: serverInfoClient(captchaEnabled: true)),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(find.byType(AltchaField), findsOneWidget);

    await tester.enterText(labeledField('Email *'), 'alice@example.com');
    await tester.enterText(labeledField('Password *'), 'password123');
    await tester.tap(find.text('Sign in'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(find.text('Complete the captcha check'), findsOneWidget);
  });

  testWidgets('WorkoutFeedList shows empty state when API returns no items', (tester) async {
    await AuthStorage.saveToken('tok');
    await ServerStorage.saveBaseUrl('https://grom.example');

    final client = MockClient((request) async {
      expect(request.url.path, '/api/v1/workouts');
      return http.Response(
        jsonEncode({'items': [], 'has_more': false}),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    final scroll = ScrollController();
    addTearDown(scroll.dispose);

    await tester.pumpWidget(
      wrap(
        WorkoutFeedList(
          nickname: 'alice',
          scope: 'own',
          scrollController: scroll,
          refreshToken: 0,
          federationEnabled: false,
          onWorkoutTap: (_) {},
          onPhotoTap: (_, __) {},
          api: ApiRequest(client: client),
        ),
      ),
    );

    await tester.pump();
    await tester.pumpAndSettle();

    expect(find.text('You have no workouts yet'), findsOneWidget);
  });
}
