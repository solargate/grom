class PushNotificationPayload {
  const PushNotificationPayload({
    required this.type,
    required this.actorDisplayName,
    required this.workoutId,
    required this.workoutTitle,
    required this.owner,
    required this.slot,
    this.eventId = '',
  });

  final String type;
  final String actorDisplayName;
  final String workoutId;
  final String workoutTitle;
  final String owner;
  final String slot;
  final String eventId;

  bool get isLiked => type == 'workout.liked';
  bool get isCommented => type == 'workout.commented';

  factory PushNotificationPayload.fromJson(Map<String, dynamic> json) {
    return PushNotificationPayload(
      type: json['type'] as String? ?? '',
      actorDisplayName: json['actor_display_name'] as String? ?? '',
      workoutId: json['workout_id'] as String? ?? '',
      workoutTitle: json['workout_title'] as String? ?? '',
      owner: json['owner'] as String? ?? '',
      slot: json['slot'] as String? ?? '',
      eventId: json['event_id'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
        'type': type,
        'actor_display_name': actorDisplayName,
        'workout_id': workoutId,
        'workout_title': workoutTitle,
        'owner': owner,
        'slot': slot,
        'event_id': eventId,
      };
}

class NotificationSlotState {
  NotificationSlotState({
    required this.slot,
    required this.type,
    required this.workoutId,
    required this.workoutTitle,
    required this.owner,
    required this.count,
    required this.lastActorDisplayName,
  });

  final String slot;
  final String type;
  final String workoutId;
  final String workoutTitle;
  final String owner;
  int count;
  String lastActorDisplayName;
}
