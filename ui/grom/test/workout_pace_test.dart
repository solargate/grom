import 'package:flutter_test/flutter_test.dart';
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
    });

    test('false for other categories', () {
      expect(isFootSport('Ride'), isFalse);
      expect(isFootSport('Swim'), isFalse);
      expect(isFootSport('unknown'), isFalse);
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
  });

  group('parsePaceMmSs', () {
    test('parses m:ss', () {
      expect(parsePaceMmSs('6:00'), 360);
      expect(parsePaceMmSs('12:22'), 742);
    });

    test('rejects invalid', () {
      expect(parsePaceMmSs(null), isNull);
      expect(parsePaceMmSs(''), isNull);
      expect(parsePaceMmSs('6'), isNull);
      expect(parsePaceMmSs('6:60'), isNull);
    });
  });

  group('paceSamplesFromSpeed', () {
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

    test('best prefers pace from speed max', () {
      expect(
        resolveBestPaceSec(speedMaxKmh: 15, samples: samples),
        240,
      );
    });

    test('best falls back to min sample pace', () {
      expect(resolveBestPaceSec(samples: samples), 300);
    });
  });
}
