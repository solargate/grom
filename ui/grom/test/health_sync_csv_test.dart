import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/health_sync_csv.dart';

void main() {
  final testdataDir = Directory(
    '../../testdata/integrations/health-sync',
  );

  group('parseHealthSyncCsvContent', () {
    test('parses Strava sample CSV by column index', () {
      final file = File('${testdataDir.path}/CYCLING 2026.07.30 16.26 Strava.csv');
      final row = parseHealthSyncCsvContent(file.readAsStringSync());
      expect(row, isNotNull);
      expect(row!.sourceApp, 'strava');
      expect(row.sport, 'CYCLING');
      expect(row.name, 'По району по делам');
      expect(row.startDate, DateTime(2026, 7, 30, 16, 26, 55));
      expect(row.durationTotalSeconds, 3765);
      expect(row.durationMovingSeconds, 1474);
      expect(row.distanceKm, closeTo(5.1734, 0.0001));
      expect(row.externalIdName, 'health-sync/strava');
      expect(row.sportTypeId, 'Ride');
    });

    test('parses Garmin sample CSV by column index', () {
      final file = File('${testdataDir.path}/CYCLING 2026.07.30 14.48 Garmin.csv');
      final row = parseHealthSyncCsvContent(file.readAsStringSync());
      expect(row, isNotNull);
      expect(row!.sourceApp, 'garmin');
      expect(row.sport, 'CYCLING');
      expect(row.name, 'Москва Велосипед');
      expect(row.startDate, DateTime(2026, 7, 30, 14, 48, 16));
      expect(row.durationTotalSeconds, 709);
      expect(row.durationMovingSeconds, 709);
      expect(row.distanceKm, closeTo(1.8371199, 0.000001));
      expect(row.externalIdName, 'health-sync/garmin');
    });
  });

  group('matchHealthSyncTrackFilename', () {
    test('prefers FIT over GPX when both exist', () {
      final track = matchHealthSyncTrackFilename(
        'CYCLING 2026.07.30 16.26 Strava.csv',
        [
          '2026.07.30 16.26-CYCLING.gpx',
          '2026.07.30 16.26-CYCLING.fit',
        ],
      );
      expect(track, '2026.07.30 16.26-CYCLING.fit');
    });

    test('falls back to GPX when FIT is missing', () {
      final track = matchHealthSyncTrackFilename(
        'CYCLING 2026.07.30 14.48 Garmin.csv',
        testdataDir
            .listSync()
            .whereType<File>()
            .map((file) => file.uri.pathSegments.last)
            .where((name) => name.endsWith('.gpx') || name.endsWith('.fit')),
      );
      expect(track, '2026.07.30 14.48-CYCLING.gpx');
    });

    test('returns null when no matching track exists', () {
      final track = matchHealthSyncTrackFilename(
        'CYCLING 2026.07.30 16.26 Strava.csv',
        const ['unrelated.gpx'],
      );
      expect(track, isNull);
    });
  });

  group('isHealthSyncCsvFilename', () {
    test('matches Health Sync CSV naming pattern', () {
      expect(isHealthSyncCsvFilename('CYCLING 2026.07.30 16.26 Strava.csv'), isTrue);
      expect(isHealthSyncCsvFilename('notes.txt'), isFalse);
    });
  });
}
