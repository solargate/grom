import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/pages/user_profile_page.dart';
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

Map<String, dynamic> profileJson({
  Map<String, dynamic>? viewerFollow,
}) {
  return {
    'nickname': 'bob',
    'name': 'Bob',
    'handle': 'bob@grom.example',
    'is_local': true,
    'has_avatar': false,
    if (viewerFollow != null) 'viewer_follow': viewerFollow,
  };
}

MockClient userProfileClient({
  Map<String, dynamic>? profile,
  List<Map<String, dynamic>> following = const [],
  List<Map<String, dynamic>> followers = const [],
  int followingStatus = 200,
}) {
  return MockClient((request) async {
    final path = request.url.path;
    if (path == '/api/v1/users/bob%40grom.example' ||
        path == '/api/v1/users/bob@grom.example') {
      return http.Response(
        jsonEncode(profile ?? profileJson()),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    if (path.endsWith('/following')) {
      if (followingStatus != 200) {
        return http.Response('{"error":"boom"}', followingStatus);
      }
      return http.Response(
        jsonEncode(following),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    if (path.endsWith('/followers')) {
      return http.Response(
        jsonEncode(followers),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    if (path.endsWith('/workouts')) {
      return http.Response(
        jsonEncode({'items': [], 'has_more': false}),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    if (path == '/api/v1/social/follow' && request.method == 'POST') {
      return http.Response(
        jsonEncode({
          'id': 'follow-new',
          'target_handle': 'bob@grom.example',
          'target_nickname': 'bob',
          'target_name': 'Bob',
          'target_is_local': true,
          'status': 'active',
        }),
        201,
        headers: {'content-type': 'application/json'},
      );
    }
    if (path.startsWith('/api/v1/social/follow/') &&
        request.method == 'DELETE') {
      return http.Response('', 204);
    }
    return http.Response('not found: $path', 404);
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

  testWidgets('loads identity, counts, and follow button', (tester) async {
    final client = userProfileClient(
      followers: [
        {
          'follower_handle': 'carol@grom.example',
          'follower_nickname': 'carol',
          'follower_name': 'Carol',
          'follower_is_local': true,
        },
      ],
      following: [
        {
          'id': 'f1',
          'target_handle': 'dave@grom.example',
          'target_nickname': 'dave',
          'target_name': 'Dave',
          'target_is_local': true,
          'status': 'active',
        },
      ],
    );

    await tester.pumpWidget(
      wrap(
        UserProfilePage(
          handle: 'bob@grom.example',
          viewerNickname: 'alice',
          api: ApiRequest(client: client),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('bob'), findsOneWidget);
    expect(find.text('Bob'), findsOneWidget);
    expect(find.text('Following: 1'), findsOneWidget);
    expect(find.text('Followers: 1'), findsOneWidget);
    expect(find.byTooltip('Follow'), findsOneWidget);
  });

  testWidgets('follow then unfollow toggles button', (tester) async {
    var profile = profileJson();
    final client = MockClient((request) async {
      final path = request.url.path;
      if (path.contains('/users/') &&
          !path.endsWith('/following') &&
          !path.endsWith('/followers') &&
          !path.endsWith('/workouts')) {
        return http.Response(
          jsonEncode(profile),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (path.endsWith('/following') || path.endsWith('/followers')) {
        return http.Response(
          '[]',
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (path.endsWith('/workouts')) {
        return http.Response(
          jsonEncode({'items': [], 'has_more': false}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (path == '/api/v1/social/follow' && request.method == 'POST') {
        profile = profileJson(
          viewerFollow: {'id': 'follow-1', 'status': 'active'},
        );
        return http.Response(
          jsonEncode({
            'id': 'follow-1',
            'target_handle': 'bob@grom.example',
            'target_nickname': 'bob',
            'target_name': 'Bob',
            'target_is_local': true,
            'status': 'active',
          }),
          201,
          headers: {'content-type': 'application/json'},
        );
      }
      if (path == '/api/v1/social/follow/follow-1' &&
          request.method == 'DELETE') {
        profile = profileJson();
        return http.Response('', 204);
      }
      return http.Response('not found: $path', 404);
    });

    await tester.pumpWidget(
      wrap(
        UserProfilePage(
          handle: 'bob@grom.example',
          viewerNickname: 'alice',
          api: ApiRequest(client: client),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.byTooltip('Follow'), findsOneWidget);
    await tester.tap(find.byTooltip('Follow'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    expect(find.byTooltip('Unfollow'), findsOneWidget);

    await tester.tap(find.byTooltip('Unfollow'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    expect(find.byTooltip('Follow'), findsOneWidget);
  });

  testWidgets('following list failure still shows profile', (tester) async {
    final client = userProfileClient(
      followingStatus: 502,
      followers: [
        {
          'follower_handle': 'carol@grom.example',
          'follower_nickname': 'carol',
          'follower_name': 'Carol',
          'follower_is_local': true,
        },
      ],
    );

    await tester.pumpWidget(
      wrap(
        UserProfilePage(
          handle: 'bob@grom.example',
          viewerNickname: 'alice',
          api: ApiRequest(client: client),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('bob'), findsOneWidget);
    expect(find.text('Following: 0'), findsOneWidget);
    expect(find.text('Followers: 1'), findsOneWidget);
    expect(find.textContaining('failed', findRichText: true), findsNothing);
  });
}
