import 'package:flutter_test/flutter_test.dart';
import 'package:grom/widgets/create_workout_name_sync.dart';

void main() {
  test('follows sport label while synced', () {
    final sync = CreateWorkoutNameSync();
    expect(sync.nameForSportChange('Run'), 'Run');
    expect(sync.nameForSportChange('Ride'), 'Ride');
    expect(sync.synced, isTrue);
  });

  test('user edit locks name from sport changes', () {
    final sync = CreateWorkoutNameSync();
    sync.onUserEdited();
    expect(sync.synced, isFalse);
    expect(sync.nameForSportChange('Ride'), isNull);
  });

  test('empty user edit still locks', () {
    final sync = CreateWorkoutNameSync();
    sync.onUserEdited();
    expect(sync.nameForSportChange('Walk'), isNull);
  });
}
