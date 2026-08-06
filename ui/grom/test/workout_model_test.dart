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

  test('Workout.fromJson reads optional stats fields', () {
    final workout = Workout.fromJson({
      'id': 'workout-2',
      'name': 'Track run',
      'sport_type': 'Run',
      'start_date': '2026-07-17T06:30:00Z',
      'duration_seconds': 3600,
      'duration_total_seconds': 3900,
      'distance': 10000,
      'temp_avg_kmm': '6:00',
      'speed_max_kmh': 32.4,
      'speed_avg_kmh': 10,
      'elevation_gain': 80,
      'heart_rate_avg': 145,
      'heart_rate_max': 187,
      'steps_total': 9000,
      'calories': 500,
    });

    expect(workout.durationTotalSeconds, 3900);
    expect(workout.tempAvgKmm, '6:00');
    expect(workout.speedMaxKmh, 32.4);
    expect(workout.speedAvgKmh, 10);
    expect(workout.elevationGain, 80);
    expect(workout.heartRateAvg, 145);
    expect(workout.heartRateMax, 187);
    expect(workout.stepsTotal, 9000);
    expect(workout.calories, 500);
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

  test('CreateWorkoutDraft.toJson always includes equipment_ids', () {
    final draft = CreateWorkoutDraft(
      name: 'Easy run',
      description: '',
      sportType: 'Run',
      startDate: DateTime.parse('2026-07-17T21:15:00Z'),
      durationSeconds: 1800,
      distanceKm: 5,
    );

    final json = draft.toJson();
    expect(json.containsKey('equipment_ids'), isTrue);
    expect(json['equipment_ids'], isEmpty);
  });

  test('Workout.fromJson reads likes summary fields', () {
    final withLikes = Workout.fromJson({
      'id': 'workout-3',
      'name': 'Liked run',
      'sport_type': 'Run',
      'start_date': '2026-07-17T06:30:00Z',
      'duration_seconds': 1800,
      'distance': 5000,
      'likes_count': 3,
      'liked_by_me': true,
      'can_like': true,
    });
    expect(withLikes.likesCount, 3);
    expect(withLikes.likedByMe, isTrue);
    expect(withLikes.canLike, isTrue);

    final defaults = Workout.fromJson({
      'id': 'workout-4',
      'name': 'Plain run',
      'sport_type': 'Run',
      'start_date': '2026-07-17T06:30:00Z',
      'duration_seconds': 1800,
      'distance': 5000,
    });
    expect(defaults.likesCount, 0);
    expect(defaults.likedByMe, isFalse);
    expect(defaults.canLike, isFalse);
  });

  test('WorkoutLikeState and WorkoutLikesResponse parse API payloads', () {
    final state = WorkoutLikeState.fromJson({
      'count': 2,
      'liked_by_me': true,
    });
    expect(state.count, 2);
    expect(state.likedByMe, isTrue);

    final emptyState = WorkoutLikeState.fromJson(<String, dynamic>{});
    expect(emptyState.count, 0);
    expect(emptyState.likedByMe, isFalse);

    final likes = WorkoutLikesResponse.fromJson({
      'count': 1,
      'users': [
        {
          'handle': 'alice@localhost',
          'nickname': 'alice',
          'name': 'Alice',
          'is_local': true,
          'has_avatar': true,
          'avatar_url': '/api/v1/users/alice/avatar',
        },
      ],
    });
    expect(likes.count, 1);
    expect(likes.users, hasLength(1));
    expect(likes.users.single.handle, 'alice@localhost');
    expect(likes.users.single.hasAvatar, isTrue);
    expect(likes.users.single.avatarUrl, '/api/v1/users/alice/avatar');

    final noUsers = WorkoutLikesResponse.fromJson({'count': 0});
    expect(noUsers.count, 0);
    expect(noUsers.users, isEmpty);
  });
}
