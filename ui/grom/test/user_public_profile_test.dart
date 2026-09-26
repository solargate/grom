import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/user_public_profile.dart';
import 'package:grom/models/workout.dart';

void main() {
  group('ViewerFollow', () {
    test('fromJson and isActive for active/pending', () {
      expect(
        ViewerFollow.fromJson({'id': 'f1', 'status': 'active'}).isActive,
        isTrue,
      );
      expect(
        ViewerFollow.fromJson({'id': 'f2', 'status': 'pending'}).isActive,
        isTrue,
      );
      expect(
        ViewerFollow.fromJson({'id': 'f3', 'status': 'rejected'}).isActive,
        isFalse,
      );
      final empty = ViewerFollow.fromJson({});
      expect(empty.id, '');
      expect(empty.status, '');
      expect(empty.isActive, isFalse);
    });
  });

  group('UserPublicProfile.fromJson', () {
    test('reads local profile with viewer_follow', () {
      final profile = UserPublicProfile.fromJson({
        'nickname': 'bob',
        'name': 'Bob',
        'handle': 'bob@grom.example',
        'is_local': true,
        'has_avatar': true,
        'avatar_url': '/api/v1/users/bob/avatar',
        'viewer_follow': {'id': 'follow-1', 'status': 'active'},
      });

      expect(profile.nickname, 'bob');
      expect(profile.name, 'Bob');
      expect(profile.handle, 'bob@grom.example');
      expect(profile.isLocal, isTrue);
      expect(profile.hasAvatar, isTrue);
      expect(profile.avatarUrl, '/api/v1/users/bob/avatar');
      expect(profile.viewerFollow?.id, 'follow-1');
      expect(profile.viewerFollow?.status, 'active');
    });

    test('reads remote profile without viewer_follow', () {
      final profile = UserPublicProfile.fromJson({
        'nickname': 'carol',
        'name': '',
        'handle': 'carol@remote.test',
        'is_local': false,
        'has_avatar': false,
      });

      expect(profile.isLocal, isFalse);
      expect(profile.name, '');
      expect(profile.hasAvatar, isFalse);
      expect(profile.avatarUrl, isNull);
      expect(profile.viewerFollow, isNull);
    });

    test('defaults is_local when missing', () {
      final profile = UserPublicProfile.fromJson({
        'nickname': 'dave',
        'handle': 'dave@grom.example',
      });
      expect(profile.isLocal, isTrue);
      expect(profile.name, '');
    });
  });

  group('Workout.objectId', () {
    test('fromJson reads object_id', () {
      final workout = Workout.fromJson({
        'id': 'abcd1234',
        'name': 'Remote run',
        'sport_type': 'Run',
        'start_date': '2026-07-08T10:00:00Z',
        'duration_seconds': 1800,
        'distance': 5000,
        'object_id': 'https://remote.test/users/bob/workouts/abcd1234',
        'owner': 'bob',
      });

      expect(
        workout.objectId,
        'https://remote.test/users/bob/workouts/abcd1234',
      );
      expect(
        workout.workoutResourceQuery()['object_id'],
        'https://remote.test/users/bob/workouts/abcd1234',
      );
      expect(workout.workoutResourceQuery()['owner'], 'bob');
    });

    test('fromJson omits object_id when absent', () {
      final workout = Workout.fromJson({
        'id': 'local1',
        'name': 'Local',
        'sport_type': 'Run',
        'start_date': '2026-07-08T10:00:00Z',
        'duration_seconds': 60,
        'distance': 100,
      });
      expect(workout.objectId, isNull);
      expect(workout.workoutResourceQuery().containsKey('object_id'), isFalse);
    });
  });
}
