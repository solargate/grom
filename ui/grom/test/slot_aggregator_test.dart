import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/notifications/push_payload.dart';
import 'package:grom/services/notifications/slot_aggregator.dart';

void main() {
  test('aggregates likes in one slot and resets after clear', () {
    final agg = SlotAggregator();
    final first = agg.ingest(
      const PushNotificationPayload(
        type: 'workout.liked',
        actorDisplayName: 'Alice',
        workoutId: 'w1',
        workoutTitle: 'Run',
        owner: 'bob',
        slot: 'liked:bob:w1',
      ),
    );
    expect(first!.count, 1);
    expect(first.lastActorDisplayName, 'Alice');

    final second = agg.ingest(
      const PushNotificationPayload(
        type: 'workout.liked',
        actorDisplayName: 'Carol',
        workoutId: 'w1',
        workoutTitle: 'Run',
        owner: 'bob',
        slot: 'liked:bob:w1',
      ),
    );
    expect(second!.count, 2);
    expect(second.lastActorDisplayName, 'Carol');

    agg.clearWorkout('bob', 'w1');
    expect(agg.get('liked:bob:w1'), isNull);
  });

  test('keeps like and comment slots separate', () {
    final agg = SlotAggregator();
    agg.ingest(
      const PushNotificationPayload(
        type: 'workout.liked',
        actorDisplayName: 'Alice',
        workoutId: 'w1',
        workoutTitle: 'Run',
        owner: 'bob',
        slot: 'liked:bob:w1',
      ),
    );
    agg.ingest(
      const PushNotificationPayload(
        type: 'workout.commented',
        actorDisplayName: 'Alice',
        workoutId: 'w1',
        workoutTitle: 'Run',
        owner: 'bob',
        slot: 'commented:bob:w1',
      ),
    );
    expect(agg.get('liked:bob:w1')!.count, 1);
    expect(agg.get('commented:bob:w1')!.count, 1);
    expect(
      SlotAggregator.notificationIdForSlot('liked:bob:w1'),
      isNot(SlotAggregator.notificationIdForSlot('commented:bob:w1')),
    );
  });
}
