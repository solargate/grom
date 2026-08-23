import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/models/my_workouts_layout.dart';
import 'package:grom/models/workout.dart';
import 'package:grom/widgets/workout_list_row.dart';
import 'package:intl/date_symbol_data_local.dart';

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
}
