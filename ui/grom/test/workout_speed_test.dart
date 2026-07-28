import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/workout_speed.dart';

void main() {
  group('WorkoutSpeedSeries.fromJson', () {
    test('parses samples and optional metadata', () {
      final series = WorkoutSpeedSeries.fromJson({
        'samples': [
          {
            't': '2026-07-08T10:00:01Z',
            'speed_kmh': 18.4,
            'distance_m': 5.2,
          },
          {
            't': '2026-07-08T10:00:02Z',
            'speed_kmh': 19.1,
            'distance_m': 12.5,
          },
        ],
        'speed_avg_kmh': 17.5,
        'speed_max_kmh': 32.4,
      });

      expect(series.samples, hasLength(2));
      expect(series.samples.first.speedKmh, 18.4);
      expect(series.samples.first.distanceKm, closeTo(0.0052, 1e-9));
      expect(series.speedAvgKmh, 17.5);
      expect(series.speedMaxKmh, 32.4);
    });

    test('handles empty samples', () {
      final series = WorkoutSpeedSeries.fromJson({'samples': []});
      expect(series.samples, isEmpty);
      expect(series.speedAvgKmh, isNull);
    });
  });

  group('downsampleSpeedSamples', () {
    test('keeps short series unchanged', () {
      final samples = List.generate(
        10,
        (i) => WorkoutSpeedSample(
          time: DateTime.utc(2026, 7, 8, 10, 0, i),
          speedKmh: 10 + i.toDouble(),
          distanceM: i * 10.0,
        ),
      );
      expect(downsampleSpeedSamples(samples, maxPoints: 1000), same(samples));
    });

    test('reduces to at most maxPoints keeping ends', () {
      final samples = List.generate(
        5000,
        (i) => WorkoutSpeedSample(
          time: DateTime.utc(2026, 7, 8, 10, 0, 0).add(Duration(seconds: i)),
          speedKmh: 10 + (i % 20),
          distanceM: i.toDouble(),
        ),
      );
      const maxPoints = 1000;
      final drawn = downsampleSpeedSamples(samples, maxPoints: maxPoints);
      expect(drawn.length, lessThanOrEqualTo(maxPoints));
      expect(drawn.length, greaterThan(maxPoints ~/ 2));
      expect(drawn.first.distanceM, samples.first.distanceM);
      expect(drawn.last.distanceM, samples.last.distanceM);
    });
  });

  group('resolveSpeedAvgKmh / resolveSpeedMaxKmh', () {
    final samples = [
      WorkoutSpeedSample(
        time: DateTime.utc(2026, 7, 8, 10),
        speedKmh: 10,
        distanceM: 0,
      ),
      WorkoutSpeedSample(
        time: DateTime.utc(2026, 7, 8, 10, 0, 1),
        speedKmh: 20,
        distanceM: 10,
      ),
    ];

    test('prefers positive metadata', () {
      expect(resolveSpeedAvgKmh(17.5, samples), 17.5);
      expect(resolveSpeedMaxKmh(32.4, samples), 32.4);
    });

    test('falls back to samples when metadata missing', () {
      expect(resolveSpeedAvgKmh(null, samples), 15);
      expect(resolveSpeedMaxKmh(null, samples), 20);
    });

    test('falls back when metadata is non-positive', () {
      expect(resolveSpeedAvgKmh(0, samples), 15);
      expect(resolveSpeedMaxKmh(-1, samples), 20);
    });
  });
}
