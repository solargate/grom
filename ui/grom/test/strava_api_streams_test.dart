import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_streams.dart';

void main() {
  test('parseStravaStreamsByType returns empty when length under 2', () {
    expect(
      parseStravaStreamsByType({
        'time': {
          'data': [0],
        },
        'heartrate': {
          'data': [120],
        },
      }),
      isEmpty,
    );
    expect(parseStravaStreamsByType(const {}), isEmpty);
  });

  test('parseStravaStreamsByType drops zero and out-of-bounds latlng', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 1, 2, 3],
      },
      'latlng': {
        'data': [
          [0, 0],
          [55.75, 37.61],
          [91.0, 10.0],
          [10.0, 181.0],
        ],
      },
    });
    expect(samples, hasLength(4));
    expect(samples[0].hasGps, isFalse);
    expect(samples[0].lat, isNull);
    expect(samples[0].lon, isNull);
    expect(samples[1].hasGps, isTrue);
    expect(samples[1].lat, 55.75);
    expect(samples[1].lon, 37.61);
    expect(samples[2].hasGps, isFalse);
    expect(samples[3].hasGps, isFalse);
  });

  test('parseStravaStreamsByType ignores negative and NaN velocity', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 1, 2],
      },
      'velocity_smooth': {
        'data': [-1.0, double.nan, 2.5],
      },
    });
    expect(samples, hasLength(3));
    expect(samples[0].speedMps, isNull);
    expect(samples[1].speedMps, isNull);
    expect(samples[2].speedMps, 2.5);
  });

  test('parseStravaStreamsByType ignores negative HR cadence watts', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 1],
      },
      'heartrate': {
        'data': [-5, 140],
      },
      'cadence': {
        'data': [-1, 90],
      },
      'watts': {
        'data': [-10, 200],
      },
    });
    expect(samples.first.heartRateBpm, isNull);
    expect(samples.first.cadenceRpm, isNull);
    expect(samples.first.watts, isNull);
    expect(samples.last.heartRateBpm, 140);
    expect(samples.last.cadenceRpm, 90);
    expect(samples.last.watts, 200);
  });

  test('parseStravaStreamsByType reads altitude temp grade heartrate', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 30],
      },
      'altitude': {
        'data': [100.5, 102.0],
      },
      'temp': {
        'data': [18, 19],
      },
      'grade_smooth': {
        'data': [1.5, -0.5],
      },
      'heartrate': {
        'data': [120, 135],
      },
    });
    expect(samples, hasLength(2));
    expect(samples.first.elevation, 100.5);
    expect(samples.first.temperatureC, 18);
    expect(samples.first.gradePercent, 1.5);
    expect(samples.first.heartRateBpm, 120);
    expect(samples.last.elevation, 102.0);
    expect(samples.last.temperatureC, 19);
    expect(samples.last.gradePercent, -0.5);
    expect(samples.last.heartRateBpm, 135);
  });

  test('parseStravaStreamsByType caps length to time stream when present', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 1],
      },
      'heartrate': {
        'data': [100, 110, 120, 130],
      },
      'distance': {
        'data': [0.0, 5.0, 10.0, 15.0],
      },
    });
    expect(samples, hasLength(2));
    expect(samples.last.heartRateBpm, 110);
    expect(samples.last.distanceMeters, 5.0);
  });

  test('parseStravaStreamsByType tolerates missing stream keys', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 10, 20],
      },
      'heartrate': {
        'data': [110, 120, 130],
      },
    });
    expect(samples, hasLength(3));
    expect(samples.every((s) => s.lat == null && s.lon == null), isTrue);
    expect(samples.every((s) => s.speedMps == null), isTrue);
    expect(samples.map((s) => s.heartRateBpm).toList(), [110, 120, 130]);
  });

  test('parseStravaStreamsByType skips non-numeric stream values', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 1],
      },
      'heartrate': {
        'data': ['bad', 140],
      },
      'distance': {
        'data': [null, 12.0],
      },
    });
    expect(samples.first.heartRateBpm, isNull);
    expect(samples.first.distanceMeters, isNull);
    expect(samples.last.heartRateBpm, 140);
    expect(samples.last.distanceMeters, 12.0);
  });
}
