import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/models/my_workouts_layout.dart';
import 'package:grom/models/workout.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/widgets/workout_card.dart';
import 'package:grom/widgets/workout_feed_list.dart';
import 'package:grom/widgets/workout_list_row.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  final l10n = AppLocalizationsEn();

  setUpAll(() async {
    await initializeDateFormatting('en');
  });

  Workout workout({
    String sportType = 'Run',
    String name = 'Morning run',
    double distance = 5240,
    int durationSeconds = 1800,
  }) {
    return Workout(
      id: 'w1',
      name: name,
      description: '',
      sportType: sportType,
      startDate: DateTime.utc(2026, 8, 23, 11, 30),
      durationSeconds: durationSeconds,
      distance: distance,
    );
  }

  group('formatWorkoutListDate', () {
    final now = DateTime(2026, 8, 23, 12);

    test('web keeps long date with time', () {
      final text = formatWorkoutListDate(
        l10n,
        DateTime(2026, 8, 23, 14, 30),
        shortDate: false,
      );
      expect(text, contains('2026'));
      expect(text, contains('14:30'));
    });

    test('mobile omits year for current year and has no time', () {
      final text = formatWorkoutListDate(
        l10n,
        DateTime(2026, 8, 23, 14, 30),
        now: now,
        shortDate: true,
      );
      expect(text, '8/23');
      expect(text.contains(':'), isFalse);
    });

    test('mobile includes year for other years', () {
      final text = formatWorkoutListDate(
        l10n,
        DateTime(2025, 8, 23, 14, 30),
        now: now,
        shortDate: true,
      );
      expect(text, '8/23/2025');
    });
  });

  group('primaryMetricForWorkout', () {
    test('distance sports return formatted distance', () {
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'Run')),
        '5.24 km',
      );
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'Ride')),
        '5.24 km',
      );
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'Swim')),
        '5.24 km',
      );
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'NordicSki')),
        '5.24 km',
      );
    });

    test('duration sports return formatted duration', () {
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'WeightTraining')),
        '30m',
      );
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'Soccer')),
        '30m',
      );
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'Tennis')),
        '30m',
      );
      expect(
        primaryMetricForWorkout(l10n, workout(sportType: 'Yoga')),
        '30m',
      );
    });

    test('zero values yield null', () {
      expect(
        primaryMetricForWorkout(
          l10n,
          workout(sportType: 'Run', distance: 0),
        ),
        isNull,
      );
      expect(
        primaryMetricForWorkout(
          l10n,
          workout(sportType: 'Yoga', durationSeconds: 0),
        ),
        isNull,
      );
    });

    test('unknown sport type uses duration', () {
      expect(
        primaryMetricForWorkout(
          l10n,
          workout(sportType: 'NotARealSport', durationSeconds: 90),
        ),
        '1m',
      );
    });
  });

  testWidgets('WorkoutListRow shows date, name, metric and handles tap',
      (tester) async {
    var tapped = false;
    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: WorkoutListRow(
            workout: workout(),
            onTap: () => tapped = true,
          ),
        ),
      ),
    );

    expect(find.text('Morning run'), findsOneWidget);
    expect(find.text('5.24 km'), findsOneWidget);
    expect(find.byIcon(Icons.directions_run), findsOneWidget);

    await tester.tap(find.byType(WorkoutListRow));
    await tester.pump();
    expect(tapped, isTrue);
  });

  testWidgets('WorkoutListRow omits zero metric', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: WorkoutListRow(
            workout: workout(distance: 0),
          ),
        ),
      ),
    );

    expect(find.text('Morning run'), findsOneWidget);
    expect(find.textContaining('km'), findsNothing);
  });

  group('WorkoutFeedList mobile separators', () {
    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      await AuthStorage.clear();
      await ServerStorage.clear();
    });

    Future<void> pumpFeedList(
      WidgetTester tester, {
      required MyWorkoutsLayout layout,
    }) async {
      await AuthStorage.saveToken('tok');
      await ServerStorage.saveBaseUrl('https://grom.example');

      final client = MockClient((request) async {
        expect(request.url.path, '/api/v1/workouts');
        return http.Response(
          jsonEncode({
            'items': [
              {
                'id': 'w1',
                'name': 'Morning run',
                'sport_type': 'Run',
                'start_date': '2026-08-23T11:30:00Z',
                'duration_seconds': 1800,
                'distance': 5240,
              },
              {
                'id': 'w2',
                'name': 'Evening ride',
                'sport_type': 'Ride',
                'start_date': '2026-08-24T18:00:00Z',
                'duration_seconds': 3600,
                'distance': 12000,
              },
            ],
            'has_more': false,
          }),
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
              layout: layout,
              onWorkoutTap: (_) {},
              onPhotoTap: (_, __) {},
              api: ApiRequest(client: client),
            ),
          ),
        ),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
    }

    testWidgets('list layout uses thin dividers between rows', (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      await pumpFeedList(tester, layout: MyWorkoutsLayout.list);

      expect(find.byType(WorkoutListRow), findsNWidgets(2));
      expect(find.byType(Divider), findsOneWidget);

      debugDefaultTargetPlatformOverride = null;
    });

    testWidgets('cards layout keeps gap separators, not dividers',
        (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      await pumpFeedList(tester, layout: MyWorkoutsLayout.cards);

      expect(find.byType(WorkoutCard), findsNWidgets(2));
      expect(find.byType(Divider), findsNothing);
      expect(
        find.byWidgetPredicate(
          (w) => w is Container && w.constraints?.maxHeight == 8,
        ),
        findsOneWidget,
      );

      debugDefaultTargetPlatformOverride = null;
    });
  });
}
