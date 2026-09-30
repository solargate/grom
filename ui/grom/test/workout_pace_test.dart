import 'package:flutter_test/flutter_test.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/l10n/app_localizations_ru.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/models/workout_pace.dart';
import 'package:grom/models/workout_speed.dart';

void main() {
  group('isFootSport', () {
    test('true for foot category', () {
      expect(isFootSport('Run'), isTrue);
      expect(isFootSport('Walk'), isTrue);
      expect(isFootSport('Hike'), isTrue);
      expect(isFootSport('Wheelchair'), isTrue);
      expect(isFootSport('TrailRun'), isTrue);
      expect(isFootSport('NordicWalk'), isTrue);
    });

    test('false for other categories', () {
      expect(isFootSport('Ride'), isFalse);
      expect(isFootSport('Swim'), isFalse);
      expect(isFootSport('unknown'), isFalse);
    });
  });

  group('hasPaceChart / hasSpeedChart', () {
    final paceSamples = [
      WorkoutPaceSample(
        time: DateTime.utc(2026, 7, 8, 10),
        paceSecPerKm: 360,
        distanceM: 0,
      ),
      WorkoutPaceSample(
        time: DateTime.utc(2026, 7, 8, 10, 0, 1),
        paceSecPerKm: 300,
        distanceM: 10,
      ),
    ];
    final speedSamples = [
      WorkoutSpeedSample(
        time: DateTime.utc(2026, 7, 8, 10),
        speedKmh: 10,
        distanceM: 0,
      ),
      WorkoutSpeedSample(
        time: DateTime.utc(2026, 7, 8, 10, 0, 1),
        speedKmh: 12,
        distanceM: 10,
      ),
    ];

    test('foot sport with ≥2 pace samples shows pace chart only', () {
      expect(
        hasPaceChart(sportType: 'Run', paceSamples: paceSamples),
        isTrue,
      );
      expect(
        hasSpeedChart(sportType: 'Run', speedSamples: speedSamples),
        isFalse,
      );
    });

    test('non-foot sport with ≥2 speed samples shows speed chart only', () {
      expect(
        hasPaceChart(sportType: 'Ride', paceSamples: paceSamples),
        isFalse,
      );
      expect(
        hasSpeedChart(sportType: 'Ride', speedSamples: speedSamples),
        isTrue,
      );
    });

    test('foot sport with fewer than 2 pace samples shows neither', () {
      expect(
        hasPaceChart(sportType: 'Run', paceSamples: paceSamples.take(1).toList()),
        isFalse,
      );
      expect(
        hasSpeedChart(sportType: 'Run', speedSamples: speedSamples),
        isFalse,
      );
    });

    test('non-foot sport with null or short speed series shows neither', () {
      expect(
        hasSpeedChart(sportType: 'Ride', speedSamples: null),
        isFalse,
      );
      expect(
        hasSpeedChart(
          sportType: 'Ride',
          speedSamples: speedSamples.take(1).toList(),
        ),
        isFalse,
      );
      expect(
        hasPaceChart(sportType: 'Ride', paceSamples: paceSamples),
        isFalse,
      );
    });
  });

  group('speedKmhToPaceSec / formatPaceMmSs', () {
    test('converts 10 km/h to 6:00', () {
      expect(speedKmhToPaceSec(10), 360);
      expect(formatPaceMmSs(360), '6:00');
    });

    test('converts 12 km/h to 5:00', () {
      expect(speedKmhToPaceSec(12), 300);
      expect(formatPaceMmSs(300), '5:00');
    });

    test('omits non-positive speed', () {
      expect(speedKmhToPaceSec(0), isNull);
      expect(speedKmhToPaceSec(-1), isNull);
    });

    test('formats minutes beyond 59', () {
      expect(formatPaceMmSs(4500), '75:00');
    });

    test('returns placeholder for non-positive or non-finite', () {
      expect(formatPaceMmSs(0), '--');
      expect(formatPaceMmSs(-1), '--');
      expect(formatPaceMmSs(double.nan), '--');
      expect(formatPaceMmSs(double.infinity), '--');
    });
  });

  group('formatPaceWithUnit', () {
    test('appends localized unit', () {
      final en = AppLocalizationsEn();
      final ru = AppLocalizationsRu();
      expect(formatPaceWithUnit(en, 360), '6:00 /km');
      expect(formatPaceWithUnit(ru, 360), '6:00 /км');
      expect(formatPaceStringWithUnit(en, '5:30'), '5:30 /km');
    });
  });

  group('parsePaceMmSs', () {
    test('parses m:ss', () {
      expect(parsePaceMmSs('6:00'), 360);
      expect(parsePaceMmSs('12:22'), 742);
    });

    test('trims whitespace and accepts single-digit seconds', () {
      expect(parsePaceMmSs(' 5:30 '), 330);
      expect(parsePaceMmSs('5:3'), 303);
    });

    test('rejects invalid', () {
      expect(parsePaceMmSs(null), isNull);
      expect(parsePaceMmSs(''), isNull);
      expect(parsePaceMmSs('6'), isNull);
      expect(parsePaceMmSs('6:60'), isNull);
    });
  });

  group('paceSamplesFromSpeed', () {
    test('returns empty for empty input', () {
      expect(paceSamplesFromSpeed([]), isEmpty);
    });

    test('omits zero speeds', () {
      final samples = paceSamplesFromSpeed([
        WorkoutSpeedSample(
          time: DateTime.utc(2026, 7, 8, 10),
          speedKmh: 0,
          distanceM: 0,
        ),
        WorkoutSpeedSample(
          time: DateTime.utc(2026, 7, 8, 10, 0, 1),
          speedKmh: 10,
          distanceM: 5,
        ),
        WorkoutSpeedSample(
          time: DateTime.utc(2026, 7, 8, 10, 0, 2),
          speedKmh: 12,
          distanceM: 12,
        ),
      ]);
      expect(samples, hasLength(2));
      expect(samples.first.paceSecPerKm, 360);
      expect(samples.last.paceSecPerKm, 300);
    });
  });

  group('resolveAvgPaceSec / resolveBestPaceSec', () {
    final samples = [
      WorkoutPaceSample(
        time: DateTime.utc(2026, 7, 8, 10),
        paceSecPerKm: 360,
        distanceM: 0,
      ),
      WorkoutPaceSample(
        time: DateTime.utc(2026, 7, 8, 10, 0, 1),
        paceSecPerKm: 300,
        distanceM: 10,
      ),
    ];

    test('avg prefers temp_avg_kmm', () {
      expect(
        resolveAvgPaceSec(
          tempAvgKmm: '5:30',
          speedAvgKmh: 10,
          samples: samples,
        ),
        330,
      );
    });

    test('avg falls back to speed avg then samples', () {
      expect(
        resolveAvgPaceSec(speedAvgKmh: 10, samples: samples),
        360,
      );
      expect(
        resolveAvgPaceSec(samples: samples),
        330,
      );
    });

    test('avg falls through invalid meta to speed then samples', () {
      expect(
        resolveAvgPaceSec(
          tempAvgKmm: 'bad',
          speedAvgKmh: 12,
          samples: samples,
        ),
        300,
      );
      expect(
        resolveAvgPaceSec(tempAvgKmm: 'bad', samples: samples),
        330,
      );
    });

    test('avg returns null without meta or samples', () {
      expect(
        resolveAvgPaceSec(tempAvgKmm: 'bad', samples: []),
        isNull,
      );
      expect(resolveAvgPaceSec(samples: []), isNull);
    });

    test('best prefers pace from speed max', () {
      expect(
        resolveBestPaceSec(speedMaxKmh: 15, samples: samples),
        240,
      );
    });

    test('best falls back to min sample pace', () {
      expect(resolveBestPaceSec(samples: samples), 300);
    });

    test('best returns null with empty samples and no max', () {
      expect(resolveBestPaceSec(samples: []), isNull);
      expect(resolveBestPaceSec(speedMaxKmh: 0, samples: []), isNull);
    });
  });
}
