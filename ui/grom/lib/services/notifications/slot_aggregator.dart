import 'push_payload.dart';

/// Aggregates incoming push events into per-slot collapsed notification state.
class SlotAggregator {
  final Map<String, NotificationSlotState> _slots = {};

  NotificationSlotState? get(String slot) => _slots[slot];

  /// Applies [payload] and returns the updated slot state, or null if invalid.
  NotificationSlotState? ingest(PushNotificationPayload payload) {
    if (payload.slot.isEmpty || payload.workoutId.isEmpty) {
      return null;
    }
    if (!payload.isLiked && !payload.isCommented) {
      return null;
    }
    final existing = _slots[payload.slot];
    if (existing == null) {
      final created = NotificationSlotState(
        slot: payload.slot,
        type: payload.type,
        workoutId: payload.workoutId,
        workoutTitle: payload.workoutTitle,
        owner: payload.owner,
        count: 1,
        lastActorDisplayName: payload.actorDisplayName,
      );
      _slots[payload.slot] = created;
      return created;
    }
    existing.count += 1;
    if (payload.actorDisplayName.isNotEmpty) {
      existing.lastActorDisplayName = payload.actorDisplayName;
    }
    if (payload.workoutTitle.isNotEmpty) {
      // Keep latest title from server.
      _slots[payload.slot] = NotificationSlotState(
        slot: existing.slot,
        type: existing.type,
        workoutId: existing.workoutId,
        workoutTitle: payload.workoutTitle,
        owner: existing.owner,
        count: existing.count,
        lastActorDisplayName: existing.lastActorDisplayName,
      );
      return _slots[payload.slot];
    }
    return existing;
  }

  void clearSlot(String slot) {
    _slots.remove(slot);
  }

  void clearWorkout(String owner, String workoutId) {
    final toRemove = <String>[];
    for (final entry in _slots.entries) {
      if (entry.value.owner == owner && entry.value.workoutId == workoutId) {
        toRemove.add(entry.key);
      }
    }
    for (final key in toRemove) {
      _slots.remove(key);
    }
  }

  void clearAll() {
    _slots.clear();
  }

  /// Stable Android notification id derived from slot.
  static int notificationIdForSlot(String slot) {
    var hash = 0;
    for (final unit in slot.codeUnits) {
      hash = (hash * 31 + unit) & 0x7fffffff;
    }
    // Avoid 0 which some platforms treat specially.
    return hash == 0 ? 1 : hash;
  }
}
