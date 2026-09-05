import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_sport_map.dart';

void main() {
  test('maps Strava sport_type identifiers', () {
    expect(mapStravaSportType('Run'), 'Run');
    expect(mapStravaSportType('TrailRun'), 'TrailRun');
    expect(mapStravaSportType('EBikeRide'), 'EBikeRide');
    expect(mapStravaSportType('WeightTraining'), 'WeightTraining');
    expect(mapStravaSportType('VirtualRide'), 'Ride');
  });

  test('infers from workout name when type is generic Workout', () {
    expect(
      mapStravaSportType('Workout', activityName: 'Morning Pilates'),
      'Pilates',
    );
    expect(
      mapStravaSportType('Workout', activityName: 'Yoga flow'),
      'Yoga',
    );
  });

  test('falls back to Workout for unknown types', () {
    expect(mapStravaSportType('SomeFutureSport'), 'Workout');
  });

  test('empty raw uses default or name inference', () {
    expect(mapStravaSportType(null), 'Run');
    expect(mapStravaSportType('', activityName: 'Packraft day'), 'Packraft');
  });
}
