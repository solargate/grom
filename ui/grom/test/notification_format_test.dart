import 'package:flutter_test/flutter_test.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/services/notifications/notification_format.dart';
import 'package:grom/services/notifications/push_payload.dart';

void main() {
  final l10n = AppLocalizationsEn();

  test('formatPushBody uses count and type', () {
    final likedOne = NotificationSlotState(
      slot: 'liked:bob:w1',
      type: 'workout.liked',
      workoutId: 'w1',
      workoutTitle: 'Run',
      owner: 'bob',
      count: 1,
      lastActorDisplayName: 'Alice',
    );
    expect(formatPushBody(l10n, likedOne), l10n.pushWorkoutLikedBody('Alice'));
    likedOne.count = 3;
    expect(formatPushBody(l10n, likedOne), l10n.pushWorkoutLikedCount(3));

    final commented = NotificationSlotState(
      slot: 'commented:bob:w1',
      type: 'workout.commented',
      workoutId: 'w1',
      workoutTitle: 'Run',
      owner: 'bob',
      count: 1,
      lastActorDisplayName: 'Carol',
    );
    expect(
      formatPushBody(l10n, commented),
      l10n.pushWorkoutCommentedBody('Carol'),
    );
    commented.count = 2;
    expect(formatPushBody(l10n, commented), l10n.pushWorkoutCommentedCount(2));

    expect(formatPushBody(null, likedOne), 'Alice');
  });

  test('formatFollowTitle', () {
    const payload = PushNotificationPayload(
      type: 'user.followed',
      actorDisplayName: 'Alice',
      actorHandle: 'alice@grom.test',
      workoutId: '',
      workoutTitle: '',
      owner: '',
      slot: 'followed:1',
      eventId: '1',
    );
    expect(formatFollowTitle(l10n, payload), l10n.pushNewFollowerTitle('Alice'));
    expect(formatFollowTitle(null, payload), 'Alice');
    expect(
      formatFollowTitle(
        l10n,
        const PushNotificationPayload(
          type: 'user.followed',
          actorDisplayName: '  ',
          workoutId: '',
          workoutTitle: '',
          owner: '',
          slot: 's',
        ),
      ),
      '',
    );
  });

  test('resolvePushTap routes workout and profile', () {
    final workout = resolvePushTap(
      const PushNotificationPayload(
        type: 'workout.liked',
        actorDisplayName: 'Alice',
        workoutId: 'w1',
        workoutTitle: 'Run',
        owner: 'bob',
        slot: 'liked:bob:w1',
      ),
    );
    expect(workout.target, PushTapTarget.workout);
    expect(workout.workoutId, 'w1');
    expect(workout.owner, 'bob');
    expect(workout.cancelSlot, 'liked:bob:w1');

    final follow = resolvePushTap(
      const PushNotificationPayload(
        type: 'user.followed',
        actorDisplayName: 'Alice',
        actorHandle: 'alice@grom.test',
        workoutId: '',
        workoutTitle: '',
        owner: '',
        slot: 'followed:evt-1',
        eventId: 'evt-1',
      ),
    );
    expect(follow.target, PushTapTarget.userProfile);
    expect(follow.handle, 'alice@grom.test');
    expect(follow.nickname, 'alice');
    expect(follow.cancelSlot, 'followed:evt-1');

    final followNoHandle = resolvePushTap(
      const PushNotificationPayload(
        type: 'user.followed',
        actorDisplayName: 'Alice',
        workoutId: '',
        workoutTitle: '',
        owner: '',
        slot: '',
        eventId: 'evt-2',
      ),
    );
    expect(followNoHandle.target, PushTapTarget.userProfile);
    expect(followNoHandle.handle, '');
    expect(followNoHandle.cancelSlot, 'followed:evt-2');
  });

  test('shouldResetSlotCounter', () {
    expect(
      shouldResetSlotCounter(
        hadExistingSlot: true,
        notificationStillActive: false,
      ),
      isTrue,
    );
    expect(
      shouldResetSlotCounter(
        hadExistingSlot: true,
        notificationStillActive: true,
      ),
      isFalse,
    );
    expect(
      shouldResetSlotCounter(
        hadExistingSlot: false,
        notificationStillActive: false,
      ),
      isFalse,
    );
  });
}
