import 'package:flutter_test/flutter_test.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/models/workout_cadence.dart';

void main() {
  group('supportsCadenceDisplay', () {
    test('true for foot and cycle except Wheelchair', () {
      expect(supportsCadenceDisplay('Run'), isTrue);
      expect(supportsCadenceDisplay('Walk'), isTrue);
      expect(supportsCadenceDisplay('Ride'), isTrue);
      expect(supportsCadenceDisplay('GravelRide'), isTrue);
      expect(supportsCadenceDisplay('Wheelchair'), isFalse);
      expect(supportsCadenceDisplay('Swim'), isFalse);
      expect(supportsCadenceDisplay('Workout'), isFalse);
    });
  });

  group('displayCadence', () {
    test('doubles foot sports like Strava', () {
      expect(displayCadence(90, 'Run'), 180);
      expect(displayCadence(85.5, 'TrailRun'), 171);
    });

    test('keeps cycle cadence as-is', () {
      expect(displayCadence(90, 'Ride'), 90);
    });

    test('returns null for missing or non-positive', () {
      expect(displayCadence(null, 'Run'), isNull);
      expect(displayCadence(0, 'Run'), isNull);
    });
  });

  group('hasCadenceChart', () {
    final samples = [
      WorkoutCadenceSample(
        time: DateTime.utc(2026, 7, 1),
        cadence: 80,
      ),
      WorkoutCadenceSample(
        time: DateTime.utc(2026, 7, 1, 0, 1),
        cadence: 90,
      ),
    ];

    test('requires supported sport and ≥2 samples', () {
      expect(
        hasCadenceChart(sportType: 'Run', samples: samples),
        isTrue,
      );
      expect(
        hasCadenceChart(sportType: 'Wheelchair', samples: samples),
        isFalse,
      );
      expect(
        hasCadenceChart(sportType: 'Swim', samples: samples),
        isFalse,
      );
      expect(
        hasCadenceChart(sportType: 'Run', samples: samples.take(1).toList()),
        isFalse,
      );
      expect(
        hasCadenceChart(sportType: 'Run', samples: null),
        isFalse,
      );
    });
  });

  group('formatCadence', () {
    final l10n = AppLocalizationsEn();

    test('uses rpm for cycle and spm for foot', () {
      expect(formatCadence(l10n, 90, footUnits: false), '90 rpm');
      expect(formatCadence(l10n, 180, footUnits: true), '180 spm');
    });
  });

  group('WorkoutCadenceSeries.fromJson', () {
    test('parses samples with distance and metadata', () {
      final series = WorkoutCadenceSeries.fromJson({
        'samples': [
          {
            't': '2026-07-08T10:00:01Z',
            'cadence': 80,
            'distance_m': 5.2,
          },
          {
            't': '2026-07-08T10:00:02Z',
            'cadence': 90,
            'distance_m': 12.5,
          },
        ],
        'cadence_avg': 85,
        'cadence_max': 95,
        'has_gps': true,
      });

      expect(series.samples, hasLength(2));
      expect(series.samples.first.cadence, 80);
      expect(series.samples.first.distanceKm, closeTo(0.0052, 1e-9));
      expect(series.cadenceAvg, 85);
      expect(series.cadenceMax, 95);
      expect(series.hasGps, isTrue);
    });

    test('parses samples without distance when no GPS', () {
      final series = WorkoutCadenceSeries.fromJson({
        'samples': [
          {
            't': '2026-07-08T10:00:01Z',
            'cadence': 70,
          },
          {
            't': '2026-07-08T10:01:01Z',
            'cadence': 90,
          },
        ],
        'has_gps': false,
      });

      expect(series.samples, hasLength(2));
      expect(series.samples.first.distanceM, isNull);
      expect(series.hasGps, isFalse);
      expect(
        cadenceMinutesFromSeriesStart(
          series.samples,
          series.samples.last.time,
        ),
        closeTo(1.0, 1e-9),
      );
    });
  });

  group('resolveCadenceAvg / resolveCadenceMax', () {
    final samples = [
      WorkoutCadenceSample(
        time: DateTime.utc(2026, 7, 8, 10),
        cadence: 70,
      ),
      WorkoutCadenceSample(
        time: DateTime.utc(2026, 7, 8, 10, 0, 1),
        cadence: 90,
      ),
    ];

    test('prefers positive metadata', () {
      expect(resolveCadenceAvg(85, samples), 85);
      expect(resolveCadenceMax(95, samples), 95);
    });

    test('falls back to samples when metadata missing', () {
      expect(resolveCadenceAvg(null, samples), 80);
      expect(resolveCadenceMax(null, samples), 90);
    });

    test('falls back when metadata is non-positive', () {
      expect(resolveCadenceAvg(0, samples), 80);
      expect(resolveCadenceMax(-1, samples), 90);
    });
  });

  group('cadenceMinutesFromSeriesStart', () {
    test('returns 0 for empty series', () {
      expect(cadenceMinutesFromSeriesStart([], DateTime.utc(2026, 7, 8, 10)), 0);
    });

    test('returns fractional minutes from first sample', () {
      final samples = [
        WorkoutCadenceSample(
          time: DateTime.utc(2026, 7, 8, 10),
          cadence: 70,
        ),
        WorkoutCadenceSample(
          time: DateTime.utc(2026, 7, 8, 10, 0, 30),
          cadence: 80,
        ),
      ];
      expect(
        cadenceMinutesFromSeriesStart(samples, samples.last.time),
        closeTo(0.5, 1e-9),
      );
    });
  });
}
