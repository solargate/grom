import 'package:flutter_test/flutter_test.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/models/workout.dart';
import 'package:grom/models/workout_stats.dart';

void main() {
  final l10n = AppLocalizationsEn();

  Workout workout({
    String sportType = 'Run',
    double distance = 10000,
    int durationSeconds = 3600,
    int? durationTotalSeconds = 3900,
    String? tempAvgKmm = '6:00',
    double? speedAvgKmh = 10,
    double? elevationGain = 80,
    double? heartRateAvg = 145,
    int? stepsTotal = 9000,
    double? calories = 500,
  }) {
    return Workout(
      id: 'w1',
      name: 'Test',
      description: '',
      sportType: sportType,
      startDate: DateTime.utc(2026, 7, 17),
      durationSeconds: durationSeconds,
      distance: distance,
      durationTotalSeconds: durationTotalSeconds,
      tempAvgKmm: tempAvgKmm,
      speedAvgKmh: speedAvgKmh,
      elevationGain: elevationGain,
      heartRateAvg: heartRateAvg,
      stepsTotal: stepsTotal,
      calories: calories,
    );
  }

  test('buildWorkoutStats orders fields and skips empty values', () {
    final stats = buildWorkoutStats(
      l10n,
      workout(
        distance: 0,
        tempAvgKmm: null,
        durationSeconds: 1800,
        elevationGain: 0,
        speedAvgKmh: 12,
        durationTotalSeconds: null,
        heartRateAvg: 120,
        stepsTotal: 0,
        calories: 200,
      ),
    );

    expect(stats.map((s) => s.label).toList(), [
      'Time',
      'Avg. speed',
      'Avg. heart rate',
      'Calories',
    ]);
  });

  test('pace is omitted for non-foot sports', () {
    final stats = buildWorkoutStats(
      l10n,
      workout(sportType: 'Ride', tempAvgKmm: '3:00'),
    );

    expect(stats.any((s) => s.label == 'Pace'), isFalse);
  });

  test('pace is included for foot sports', () {
    final stats = buildWorkoutStats(l10n, workout(sportType: 'Hike'));
    expect(stats.map((s) => s.label).first, 'Distance');
    expect(stats[1].label, 'Pace');
    expect(stats[1].value, '6:00');
  });

  test('chunkWorkoutStats respects maxRows', () {
    final stats = buildWorkoutStats(l10n, workout());
    expect(stats.length, greaterThan(3));

    final feedRows = chunkWorkoutStats(stats, maxRows: 1);
    expect(feedRows, hasLength(1));
    expect(feedRows.single, hasLength(3));

    final detailRows = chunkWorkoutStats(stats);
    expect(detailRows.length, greaterThan(1));
  });
}
