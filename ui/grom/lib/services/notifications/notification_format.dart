import '../../l10n/app_localizations.dart';
import 'push_payload.dart';

/// Localized body for a collapsed like/comment notification slot.
String formatPushBody(AppLocalizations? l10n, NotificationSlotState state) {
  if (l10n == null) {
    return state.lastActorDisplayName;
  }
  if (state.type == 'workout.liked') {
    if (state.count <= 1) {
      return l10n.pushWorkoutLikedBody(state.lastActorDisplayName);
    }
    return l10n.pushWorkoutLikedCount(state.count);
  }
  if (state.type == 'workout.commented') {
    if (state.count <= 1) {
      return l10n.pushWorkoutCommentedBody(state.lastActorDisplayName);
    }
    return l10n.pushWorkoutCommentedCount(state.count);
  }
  return state.lastActorDisplayName;
}

/// Localized title for a new-follower notification, or empty when no actor name.
String formatFollowTitle(AppLocalizations? l10n, PushNotificationPayload payload) {
  final name = payload.actorDisplayName.trim();
  if (name.isEmpty) {
    return '';
  }
  if (l10n == null) {
    return name;
  }
  return l10n.pushNewFollowerTitle(name);
}

enum PushTapTarget { workout, userProfile, none }

class PushTapAction {
  const PushTapAction.none()
      : target = PushTapTarget.none,
        workoutId = '',
        owner = '',
        handle = '',
        nickname = '',
        cancelSlot = '';

  const PushTapAction.workout({
    required this.workoutId,
    required this.owner,
    required this.cancelSlot,
  })  : target = PushTapTarget.workout,
        handle = '',
        nickname = '';

  const PushTapAction.userProfile({
    required this.handle,
    required this.nickname,
    required this.cancelSlot,
  })  : target = PushTapTarget.userProfile,
        workoutId = '',
        owner = '';

  final PushTapTarget target;
  final String workoutId;
  final String owner;
  final String handle;
  final String nickname;
  final String cancelSlot;
}

/// Resolves deeplink target and which local notification slot to cancel on tap.
PushTapAction resolvePushTap(PushNotificationPayload payload) {
  if (payload.isFollowed) {
    final idKey =
        payload.slot.isNotEmpty ? payload.slot : 'followed:${payload.eventId}';
    return PushTapAction.userProfile(
      handle: payload.actorHandle.trim(),
      nickname: payload.actorNickname,
      cancelSlot: idKey,
    );
  }
  if (payload.workoutId.isEmpty || payload.owner.isEmpty) {
    return const PushTapAction.none();
  }
  return PushTapAction.workout(
    workoutId: payload.workoutId,
    owner: payload.owner,
    cancelSlot: payload.slot,
  );
}

/// Whether an existing collapse slot should be reset before ingesting a new event
/// (OS notification was dismissed while slot state remained in memory).
bool shouldResetSlotCounter({
  required bool hadExistingSlot,
  required bool notificationStillActive,
}) {
  return hadExistingSlot && !notificationStillActive;
}
