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
}
