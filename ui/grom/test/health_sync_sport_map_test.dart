import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/health_sync_sport_map.dart';

void main() {
  group('mapHealthSyncSportType', () {
    test('maps common Health Sync sports to Grom sport types', () {
      expect(mapHealthSyncSportType('CYCLING'), 'Ride');
      expect(mapHealthSyncSportType('RUNNING'), 'Run');
      expect(mapHealthSyncSportType('WALKING'), 'Walk');
      expect(mapHealthSyncSportType('SWIMMING'), 'Swim');
      expect(mapHealthSyncSportType('HIKING'), 'Hike');
    });

    test('normalizes case and spaces', () {
      expect(mapHealthSyncSportType('cycling'), 'Ride');
      expect(mapHealthSyncSportType('Trail Running'), 'TrailRun');
    });

    test('returns Workout for unknown sports', () {
      expect(mapHealthSyncSportType('UNKNOWN_SPORT'), defaultHealthSyncSportTypeId);
      expect(mapHealthSyncSportType(''), defaultHealthSyncSportTypeId);
    });
  });
}
