import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/pages/grom_api_tab.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await AuthStorage.saveToken('session-token');
  });

  testWidgets('GromApiTab lists personal access tokens', (tester) async {
    final client = MockClient((request) async {
      if (request.url.path == '/api/v1/auth/pat' && request.method == 'GET') {
        return http.Response(
          jsonEncode([
            {
              'id': 'pat-1',
              'name': 'Script',
              'token_prefix': 'grom_pat_ab',
              'scopes': ['workouts:read'],
              'created_at': '2026-08-12T10:00:00Z',
            },
          ]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      return http.Response('not found', 404);
    });

    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: GromApiTab(api: ApiRequest(client: client)),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('Script'), findsOneWidget);
    expect(find.textContaining('grom_pat_ab'), findsOneWidget);
  });
}
