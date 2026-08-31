import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/my_workouts_layout.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/widgets/workout_feed_list.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await AuthStorage.clear();
    await ServerStorage.clear();
  });

  testWidgets('WorkoutFeedList sends sport_types when filter provided', (tester) async {
    await AuthStorage.saveToken('tok');
    await ServerStorage.saveBaseUrl('https://grom.example');

    String? capturedSportTypes;
    final client = MockClient((request) async {
      expect(request.url.path, '/api/v1/workouts');
      capturedSportTypes = request.url.queryParameters['sport_types'];
      return http.Response(
        jsonEncode({'items': <dynamic>[], 'has_more': false}),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    final scroll = ScrollController();
    addTearDown(scroll.dispose);

    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: WorkoutFeedList(
            nickname: 'alice',
            scope: 'own',
            scrollController: scroll,
            refreshToken: 0,
            federationEnabled: false,
            layout: MyWorkoutsLayout.list,
            sportTypes: const ['Run', 'Ride'],
            onWorkoutTap: (_) {},
            onPhotoTap: (_, __) {},
            api: ApiRequest(client: client),
          ),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(capturedSportTypes, 'Run,Ride');
  });

  testWidgets('WorkoutFeedList sends empty sport_types for empty filter', (tester) async {
    await AuthStorage.saveToken('tok');
    await ServerStorage.saveBaseUrl('https://grom.example');

    String? capturedSportTypes;
    final client = MockClient((request) async {
      capturedSportTypes = request.url.queryParameters['sport_types'];
      return http.Response(
        jsonEncode({'items': <dynamic>[], 'has_more': false}),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    final scroll = ScrollController();
    addTearDown(scroll.dispose);

    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: WorkoutFeedList(
            nickname: 'alice',
            scope: 'own',
            scrollController: scroll,
            refreshToken: 0,
            federationEnabled: false,
            layout: MyWorkoutsLayout.list,
            sportTypes: const [],
            onWorkoutTap: (_) {},
            onPhotoTap: (_, __) {},
            api: ApiRequest(client: client),
          ),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(capturedSportTypes, '');
  });
}
