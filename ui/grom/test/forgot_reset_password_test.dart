import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/forgot_password.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/reset_password.dart';
import 'package:grom/server_storage.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:shared_preferences/shared_preferences.dart';

Finder labeledField(String label) {
  return find.ancestor(
    of: find.text(label),
    matching: find.byType(TextFormField),
  );
}

Finder forgotEmailField() => labeledField('Email *');

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
    await ServerStorage.saveBaseUrl('https://grom.example');
  });

  group('ForgotPasswordForm', () {
    testWidgets('submits email and shows success snackbar', (tester) async {
      var forgotCalled = false;
      final client = MockClient((request) async {
        if (request.url.path == '/api/v1/server-info') {
          return http.Response(jsonEncode({'name': 'Test'}), 200);
        }
        if (request.url.path == '/api/v1/auth/password/forgot') {
          forgotCalled = true;
          expect(jsonDecode(request.body)['email'], 'alice@example.com');
          return http.Response('', 204);
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(
        wrap(ForgotPasswordForm(api: ApiRequest(client: client))),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      await tester.enterText(forgotEmailField(), 'alice@example.com');
      await tester.tap(find.text('Send reset link'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(forgotCalled, isTrue);
    });

    testWidgets('shows captcha required snackbar when enabled without payload', (tester) async {
      final client = MockClient((request) async {
        if (request.url.path == '/api/v1/server-info') {
          return http.Response(
            jsonEncode({'captcha_enabled': true}),
            200,
            headers: {'content-type': 'application/json'},
          );
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(
        wrap(ForgotPasswordForm(api: ApiRequest(client: client))),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      await tester.enterText(forgotEmailField(), 'alice@example.com');
      await tester.tap(find.text('Send reset link'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Complete the captcha check'), findsOneWidget);
    });

    testWidgets('maps API errors to snackbar', (tester) async {
      final client = MockClient((request) async {
        if (request.url.path == '/api/v1/server-info') {
          return http.Response(jsonEncode({'name': 'Test'}), 200);
        }
        if (request.url.path == '/api/v1/auth/password/forgot') {
          return http.Response(
            jsonEncode({'error': 'too many requests'}),
            429,
            headers: {'content-type': 'application/json'},
          );
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(
        wrap(ForgotPasswordForm(api: ApiRequest(client: client))),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      await tester.enterText(forgotEmailField(), 'alice@example.com');
      await tester.tap(find.text('Send reset link'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('too many requests'), findsOneWidget);
    });
  });

  group('ResetPasswordPage', () {
    testWidgets('validates password length and confirmation', (tester) async {
      await tester.pumpWidget(
        wrap(const ResetPasswordPage(token: 'reset-token')),
      );
      await tester.pumpAndSettle();

      final fields = find.byType(TextFormField);
      await tester.enterText(fields.at(0), 'short');
      await tester.enterText(fields.at(1), 'mismatch');
      await tester.tap(find.text('Update password'));
      await tester.pumpAndSettle();

      expect(find.text('Password must be at least 8 characters'), findsOneWidget);

      await tester.enterText(fields.at(0), 'longenough');
      await tester.enterText(fields.at(1), 'different');
      await tester.tap(find.text('Update password'));
      await tester.pumpAndSettle();

      expect(find.text('Passwords do not match'), findsOneWidget);
    });

    testWidgets('successful reset shows dialog and calls API', (tester) async {
      var resetCalled = false;
      final client = MockClient((request) async {
        expect(request.url.path, '/api/v1/auth/password/reset');
        resetCalled = true;
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['token'], 'reset-token');
        expect(body['password'], 'newpassword1');
        return http.Response('', 204);
      });

      await tester.pumpWidget(
        MaterialApp(
          theme: buildAppTheme(),
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          locale: const Locale('en'),
          initialRoute: '/',
          routes: {
            '/': (context) => Scaffold(
                  body: ResetPasswordPage(
                    token: 'reset-token',
                    api: ApiRequest(client: client),
                  ),
                ),
          },
        ),
      );
      await tester.pumpAndSettle();

      final fields = find.byType(TextFormField);
      await tester.enterText(fields.at(0), 'newpassword1');
      await tester.enterText(fields.at(1), 'newpassword1');
      await tester.tap(find.text('Update password'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(resetCalled, isTrue);
      expect(find.text('Password updated. Please sign in.'), findsOneWidget);
      await tester.tap(find.text('Sign in'));
      await tester.pump();
    });
  });
}
