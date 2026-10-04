import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/notifications/push_payload.dart';

void main() {
  test('parses follower payload and nickname from handle', () {
    final payload = PushNotificationPayload.fromJson({
      'type': 'user.followed',
      'actor_display_name': 'Alice',
      'actor_handle': 'alice@grom.test',
      'workout_id': '',
      'workout_title': '',
      'owner': '',
      'slot': 'followed:evt-1',
      'event_id': 'evt-1',
    });
    expect(payload.isFollowed, isTrue);
    expect(payload.actorDisplayName, 'Alice');
    expect(payload.actorHandle, 'alice@grom.test');
    expect(payload.actorNickname, 'alice');
    expect(payload.toJson()['actor_handle'], 'alice@grom.test');
  });

  test('parses liked and commented payloads with round-trip', () {
    final liked = PushNotificationPayload.fromJson({
      'type': 'workout.liked',
      'actor_display_name': 'Alice',
      'actor_handle': 'alice@grom.test',
      'workout_id': 'w1',
      'workout_title': 'Run',
      'owner': 'bob',
      'slot': 'liked:bob:w1',
      'event_id': 'e1',
    });
    expect(liked.isLiked, isTrue);
    expect(liked.isCommented, isFalse);
    expect(liked.toJson()['workout_id'], 'w1');

    final commented = PushNotificationPayload.fromJson({
      'type': 'workout.commented',
      'actor_display_name': 'Carol',
      'workout_id': 'w1',
      'workout_title': 'Run',
      'owner': 'bob',
      'slot': 'commented:bob:w1',
    });
    expect(commented.isCommented, isTrue);
    expect(commented.eventId, '');
    expect(commented.actorHandle, '');
  });

  test('actorNickname handles edge cases', () {
    expect(
      const PushNotificationPayload(
        type: 'user.followed',
        actorDisplayName: 'A',
        actorHandle: 'alice@x',
        workoutId: '',
        workoutTitle: '',
        owner: '',
        slot: 's',
      ).actorNickname,
      'alice',
    );
    expect(
      const PushNotificationPayload(
        type: 'user.followed',
        actorDisplayName: 'A',
        actorHandle: 'solo',
        workoutId: '',
        workoutTitle: '',
        owner: '',
        slot: 's',
      ).actorNickname,
      'solo',
    );
    expect(
      const PushNotificationPayload(
        type: 'user.followed',
        actorDisplayName: 'A',
        workoutId: '',
        workoutTitle: '',
        owner: '',
        slot: 's',
      ).actorNickname,
      '',
    );
  });
}
