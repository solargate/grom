import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/workout.dart';

void main() {
  test('Workout.fromJson reads optional social, media, and equipment fields',
      () {
    final workout = Workout.fromJson({
      'id': 'workout-1',
      'name': 'Morning run',
      'sport_type': 'running',
      'start_date': '2026-07-17T06:30:00Z',
      'duration_seconds': 3600,
      'distance': 1234,
      'has_media': true,
      'media_files': ['first.jpg', 2],
      'author': {
        'nickname': 'runner',
        'name': 'Runner Name',
        'handle': '@runner@example.test',
        'is_local': false,
      },
      'equipment': [
        {'id': 'shoe-1', 'name': 'Road shoes', 'type': 'shoes'},
      ],
    });

    expect(workout.owner, 'runner');
    expect(workout.author?.handle, '@runner@example.test');
    expect(workout.distance, 1234.0);
    expect(workout.distanceKm, 1.234);
    expect(workout.hasMedia, isTrue);
    expect(workout.mediaFiles, ['first.jpg', '2']);
    expect(workout.equipment.single.id, 'shoe-1');
  });

  test('CreateWorkoutDraft.toJson converts kilometres and serializes UTC', () {
    final draft = CreateWorkoutDraft(
      name: 'Evening ride',
      description: 'Easy spin',
      sportType: 'cycling',
      startDate: DateTime.parse('2026-07-17T21:15:00+03:00'),
      durationSeconds: 1800,
      durationTotalSeconds: 1900,
      distanceKm: 12.5,
      speedMaxKmh: 32.5,
      speedAvgKmh: 25,
      equipmentIds: ['bike-1', 'helmet-1'],
    );

    final json = draft.toJson();

    expect(json['start_date'], '2026-07-17T18:15:00.000Z');
    expect(json['distance'], 12500.0);
    expect(json['duration_total_seconds'], 1900);
    expect(json['speed_max_kmh'], 32.5);
    expect(json['speed_avg_kmh'], 25);
    expect(json['equipment_ids'], ['bike-1', 'helmet-1']);
  });
}
