import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/pages/profile_page.dart';
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

Map<String, dynamic> meJson() {
  return {
    'id': 'u1',
    'nickname': 'alice',
    'name': 'Alice',
    'email': 'alice@example.com',
    'has_avatar': false,
  };
}

Map<String, dynamic> followerJson({
  required String nickname,
  String? name,
}) {
  return {
    'follower_handle': '$nickname@grom.example',
    'follower_nickname': nickname,
    'follower_name': name ?? nickname,
    'follower_is_local': true,
    'follower_has_avatar': false,
  };
}

Map<String, dynamic> followingJson({
  required String nickname,
  required String status,
  String? name,
}) {
  return {
    'id': 'f-$nickname',
    'target_handle': '$nickname@grom.example',
    'target_nickname': nickname,
    'target_name': name ?? nickname,
    'target_is_local': true,
    'target_has_avatar': false,
    'status': status,
  };
}

MockClient profileClient({
  List<Map<String, dynamic>> followers = const [],
  List<Map<String, dynamic>> following = const [],
}) {
  return MockClient((request) async {
    if (request.url.path == '/api/v1/auth/me') {
      return http.Response(
        jsonEncode(meJson()),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    if (request.url.path == '/api/v1/social/followers') {
      return http.Response(
        jsonEncode(followers),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    if (request.url.path == '/api/v1/social/following') {
      return http.Response(
        jsonEncode(following),
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
    await ServerStorage.saveBaseUrl('https://grom.example');
    await AuthStorage.saveToken('session-token');
  });

  tearDown(() async {
    await AuthStorage.clear();
    await ServerStorage.clear();
  });

  testWidgets('shows follower and following count cards', (tester) async {
    final client = profileClient(
      followers: [
        followerJson(nickname: 'bob'),
        followerJson(nickname: 'carol'),
      ],
      following: [
        followingJson(nickname: 'dave', status: 'active'),
        followingJson(nickname: 'erin', status: 'pending'),
        followingJson(nickname: 'frank', status: 'active'),
      ],
    );

    await tester.pumpWidget(
      wrap(
        ProfilePage(
          nickname: 'alice',
          api: ApiRequest(client: client),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('Followers: 2'), findsOneWidget);
    // Pending is excluded from the following count.
    expect(find.text('Following: 2'), findsOneWidget);
    expect(find.text('bob'), findsNothing);
    expect(find.text('dave'), findsNothing);
  });

  testWidgets('followers card opens dialog with list', (tester) async {
    final client = profileClient(
      followers: [
        followerJson(nickname: 'bob', name: 'Bob'),
      ],
    );

    await tester.pumpWidget(
      wrap(
        ProfilePage(
          nickname: 'alice',
          api: ApiRequest(client: client),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    await tester.tap(find.text('Followers: 1'));
    await tester.pumpAndSettle();

    expect(find.text('Followers'), findsWidgets);
    expect(find.text('bob'), findsOneWidget);
    expect(find.text('Bob · bob@grom.example'), findsOneWidget);
  });

  testWidgets('following dialog includes pending with label', (tester) async {
    final client = profileClient(
      following: [
        followingJson(nickname: 'dave', status: 'active', name: 'Dave'),
        followingJson(nickname: 'erin', status: 'pending', name: 'Erin'),
      ],
    );

    await tester.pumpWidget(
      wrap(
        ProfilePage(
          nickname: 'alice',
          api: ApiRequest(client: client),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('Following: 1'), findsOneWidget);

    await tester.tap(find.text('Following: 1'));
    await tester.pumpAndSettle();

    expect(find.text('dave'), findsOneWidget);
    expect(find.text('erin'), findsOneWidget);
    expect(find.text('Pending'), findsOneWidget);
  });

  testWidgets('empty followers dialog shows empty message', (tester) async {
    final client = profileClient();

    await tester.pumpWidget(
      wrap(
        ProfilePage(
          nickname: 'alice',
          api: ApiRequest(client: client),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('Followers: 0'), findsOneWidget);
    expect(find.text('Following: 0'), findsOneWidget);

    await tester.tap(find.text('Followers: 0'));
    await tester.pumpAndSettle();

    expect(find.text('No one is following you yet'), findsOneWidget);
  });
}
