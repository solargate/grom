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
}
