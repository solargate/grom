import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/workout_heartrate.dart';

void main() {
  group('WorkoutHeartRateSeries.fromJson', () {
    test('parses samples with distance and metadata', () {
      final series = WorkoutHeartRateSeries.fromJson({
        'samples': [
          {
            't': '2026-07-08T10:00:01Z',
            'heart_rate_bpm': 120,
            'distance_m': 5.2,
          },
          {
            't': '2026-07-08T10:00:02Z',
            'heart_rate_bpm': 130,
            'distance_m': 12.5,
          },
        ],
        'heart_rate_avg': 125,
        'heart_rate_max': 180,
        'has_gps': true,
      });

      expect(series.samples, hasLength(2));
      expect(series.samples.first.heartRateBpm, 120);
      expect(series.samples.first.distanceKm, closeTo(0.0052, 1e-9));
      expect(series.heartRateAvg, 125);
      expect(series.heartRateMax, 180);
      expect(series.hasGps, isTrue);
    });

    test('parses samples without distance when no GPS', () {
      final series = WorkoutHeartRateSeries.fromJson({
        'samples': [
          {
            't': '2026-07-08T10:00:01Z',
            'heart_rate_bpm': 110,
          },
          {
            't': '2026-07-08T10:01:01Z',
            'heart_rate_bpm': 140,
          },
        ],
        'has_gps': false,
      });

      expect(series.samples, hasLength(2));
      expect(series.samples.first.distanceM, isNull);
      expect(series.hasGps, isFalse);
      expect(
        minutesFromSeriesStart(
          series.samples,
          series.samples.last.time,
        ),
        closeTo(1.0, 1e-9),
      );
    });
  });

  group('resolveHeartRateAvg / resolveHeartRateMax', () {
    final samples = [
      WorkoutHeartRateSample(
        time: DateTime.utc(2026, 7, 8, 10),
        heartRateBpm: 100,
      ),
      WorkoutHeartRateSample(
        time: DateTime.utc(2026, 7, 8, 10, 0, 1),
        heartRateBpm: 140,
      ),
    ];

    test('prefers positive metadata', () {
      expect(resolveHeartRateAvg(125, samples), 125);
      expect(resolveHeartRateMax(180, samples), 180);
    });

    test('falls back to samples when metadata missing', () {
      expect(resolveHeartRateAvg(null, samples), 120);
      expect(resolveHeartRateMax(null, samples), 140);
    });

    test('falls back when metadata is non-positive', () {
      expect(resolveHeartRateAvg(0, samples), 120);
      expect(resolveHeartRateMax(-1, samples), 140);
    });
  });
}
