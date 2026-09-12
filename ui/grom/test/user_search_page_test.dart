import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/pages/user_search_page.dart';
import 'package:grom/server_storage.dart';
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

Map<String, dynamic> userJson({
  required String nickname,
  String? name,
  bool hasAvatar = false,
}) {
  return {
    'nickname': nickname,
    'name': name ?? nickname,
    'handle': '$nickname@grom.example',
    'is_local': true,
    'has_avatar': hasAvatar,
  };
}

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ServerStorage.saveBaseUrl('https://grom.example');
    await AuthStorage.saveToken('session-token');
  });

  tearDown(() async {
    await AuthStorage.clear();
    await ServerStorage.clear();
  });

  testWidgets('opens with local user catalog', (tester) async {
    var listCalls = 0;
    final client = MockClient((request) async {
      if (request.url.path == '/api/v1/users' && request.method == 'GET') {
        listCalls++;
        return http.Response(
          jsonEncode([
            userJson(nickname: 'bob', name: 'Bob'),
            userJson(nickname: 'carol', name: 'Carol'),
          ]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.url.path == '/api/v1/social/following') {
        return http.Response(
          jsonEncode([]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      return http.Response('not found', 404);
    });

    await tester.pumpWidget(
      wrap(UserSearchPage(api: ApiRequest(client: client))),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(listCalls, 1);
    expect(find.text('bob'), findsOneWidget);
    expect(find.text('carol'), findsOneWidget);
    expect(find.text('Bob · bob@grom.example'), findsOneWidget);
    expect(find.byIcon(Icons.person_add), findsNWidgets(2));
  });

  testWidgets('empty catalog shows no-other-users message', (tester) async {
    final client = MockClient((request) async {
      if (request.url.path == '/api/v1/users') {
        return http.Response(
          jsonEncode([]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.url.path == '/api/v1/social/following') {
        return http.Response(
          jsonEncode([]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      return http.Response('not found', 404);
    });

    await tester.pumpWidget(
      wrap(UserSearchPage(api: ApiRequest(client: client))),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('No other users on this server yet'), findsOneWidget);
    expect(find.text('No users found'), findsNothing);
  });

  testWidgets('search then empty query restores catalog', (tester) async {
    var listCalls = 0;
    var searchCalls = 0;
    final client = MockClient((request) async {
      if (request.url.path == '/api/v1/users' && request.method == 'GET') {
        listCalls++;
        return http.Response(
          jsonEncode([
            userJson(nickname: 'bob'),
            userJson(nickname: 'carol'),
          ]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.url.path == '/api/v1/users/search') {
        searchCalls++;
        expect(request.url.queryParameters['q'], 'zzz');
        return http.Response(
          jsonEncode([]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.url.path == '/api/v1/social/following') {
        return http.Response(
          jsonEncode([]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      return http.Response('not found', 404);
    });

    await tester.pumpWidget(
      wrap(UserSearchPage(api: ApiRequest(client: client))),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    expect(find.text('bob'), findsOneWidget);
    expect(listCalls, 1);

    await tester.enterText(find.byType(TextField), 'zzz');
    await tester.tap(find.widgetWithText(FilledButton, 'Search'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(searchCalls, 1);
    expect(find.text('No users found'), findsOneWidget);
    expect(find.text('bob'), findsNothing);

    await tester.enterText(find.byType(TextField), '');
    await tester.tap(find.widgetWithText(FilledButton, 'Search'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(listCalls, 2);
    expect(find.text('bob'), findsOneWidget);
    expect(find.text('carol'), findsOneWidget);
    expect(find.text('No users found'), findsNothing);
  });
}
