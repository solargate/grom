import 'dart:convert';

import 'package:flutter_local_notifications/flutter_local_notifications.dart';

import 'push_payload.dart';
import 'slot_aggregator.dart';

typedef NotificationTapCallback = void Function(PushNotificationPayload payload);
typedef NotificationDismissCallback = void Function(String slot);
typedef NotificationBodyBuilder = String Function(NotificationSlotState state);

class LocalNotifier {
  LocalNotifier({
    FlutterLocalNotificationsPlugin? plugin,
  }) : _plugin = plugin ?? FlutterLocalNotificationsPlugin();

  static const channelId = 'grom_social';
  static const channelName = 'Social';
  static const channelDescription =
      'Likes, comments, and new followers';

  final FlutterLocalNotificationsPlugin _plugin;
  NotificationTapCallback? onTap;
  NotificationDismissCallback? onDismiss;
  NotificationBodyBuilder? bodyBuilder;
  bool _initialized = false;

  Future<void> initialize() async {
    if (_initialized) {
      return;
    }
    const android = AndroidInitializationSettings('@mipmap/ic_launcher');
    const initSettings = InitializationSettings(android: android);
    await _plugin.initialize(
      initSettings,
      onDidReceiveNotificationResponse: _onResponse,
    );
    final androidPlugin = _plugin.resolvePlatformSpecificImplementation<
        AndroidFlutterLocalNotificationsPlugin>();
    await androidPlugin?.createNotificationChannel(
      const AndroidNotificationChannel(
        channelId,
        channelName,
        description: channelDescription,
        importance: Importance.high,
      ),
    );
    _initialized = true;
  }

  Future<bool> requestPermission() async {
    final androidPlugin = _plugin.resolvePlatformSpecificImplementation<
        AndroidFlutterLocalNotificationsPlugin>();
    final granted = await androidPlugin?.requestNotificationsPermission();
    return granted ?? true;
  }

  Future<void> showOrUpdate(NotificationSlotState state) async {
    await initialize();
    final title = state.workoutTitle.isNotEmpty
        ? state.workoutTitle
        : state.workoutId;
    final body = bodyBuilder?.call(state) ?? state.lastActorDisplayName;
    final payload = jsonEncode(
      PushNotificationPayload(
        type: state.type,
        actorDisplayName: state.lastActorDisplayName,
        workoutId: state.workoutId,
        workoutTitle: state.workoutTitle,
        owner: state.owner,
        slot: state.slot,
      ).toJson(),
    );
    await _plugin.show(
      SlotAggregator.notificationIdForSlot(state.slot),
      title,
      body,
      const NotificationDetails(
        android: AndroidNotificationDetails(
          channelId,
          channelName,
          channelDescription: channelDescription,
          importance: Importance.high,
          priority: Priority.high,
          category: AndroidNotificationCategory.social,
        ),
      ),
      payload: payload,
    );
  }

  /// Shows a one-shot follower notification (no body, no collapse).
  Future<void> showFollow({
    required String title,
    required PushNotificationPayload payload,
  }) async {
    await initialize();
    final idKey =
        payload.slot.isNotEmpty ? payload.slot : 'followed:${payload.eventId}';
    await _plugin.show(
      SlotAggregator.notificationIdForSlot(idKey),
      title,
      null,
      const NotificationDetails(
        android: AndroidNotificationDetails(
          channelId,
          channelName,
          channelDescription: channelDescription,
          importance: Importance.high,
          priority: Priority.high,
          category: AndroidNotificationCategory.social,
        ),
      ),
      payload: jsonEncode(payload.toJson()),
    );
  }

  Future<void> cancelSlot(String slot) async {
    await _plugin.cancel(SlotAggregator.notificationIdForSlot(slot));
  }

  Future<void> cancelAll() async {
    await _plugin.cancelAll();
  }

  Future<Set<int>> activeNotificationIds() async {
    try {
      final active = await _plugin.getActiveNotifications();
      return active.map((n) => n.id).whereType<int>().toSet();
    } catch (_) {
      return {};
    }
  }

  void _onResponse(NotificationResponse response) {
    final raw = response.payload;
    if (raw == null || raw.isEmpty) {
      return;
    }
    try {
      final json = jsonDecode(raw) as Map<String, dynamic>;
      final payload = PushNotificationPayload.fromJson(json);
      if (response.actionId == null ||
          response.notificationResponseType ==
              NotificationResponseType.selectedNotification) {
        onTap?.call(payload);
      }
    } catch (_) {
      // Ignore malformed payloads.
    }
  }
}
