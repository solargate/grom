import 'package:flutter/foundation.dart';

import '../../api_request.dart';
import '../../auth_storage.dart';
import '../../l10n/app_localizations.dart';
import 'installation_id_store.dart';
import 'local_notifier.dart';
import 'push_payload.dart';
import 'push_transport.dart';
import 'push_transport_factory.dart';
import 'slot_aggregator.dart';

typedef OpenWorkoutFromPush = Future<void> Function({
  required String workoutId,
  required String owner,
});

/// Coordinates push registration, local display, collapse, and deeplinks.
class NotificationService {
  NotificationService({
    ApiRequest? api,
    PushTransport? transport,
    LocalNotifier? localNotifier,
    InstallationIdStore? installationIdStore,
    SlotAggregator? aggregator,
  })  : _api = api ?? ApiRequest(),
        _transport = transport ?? createPushTransport(),
        _local = localNotifier ?? LocalNotifier(),
        _installationIds = installationIdStore ?? InstallationIdStore(),
        _aggregator = aggregator ?? SlotAggregator();

  final ApiRequest _api;
  final PushTransport _transport;
  final LocalNotifier _local;
  final InstallationIdStore _installationIds;
  final SlotAggregator _aggregator;

  OpenWorkoutFromPush? onOpenWorkout;
  AppLocalizations? _l10n;
  bool _started = false;
  String? _vapidPublicKey;

  Future<void> start({
    required AppLocalizations l10n,
    OpenWorkoutFromPush? onOpenWorkout,
  }) async {
    _l10n = l10n;
    this.onOpenWorkout = onOpenWorkout;
    if (_started) {
      return;
    }
    _local.onTap = _handleTap;
    _local.bodyBuilder = _formatBody;
    await _local.initialize();
    await _transport.initialize(
      onEndpoint: _registerEndpoint,
      onMessage: _handleMessage,
    );
    _started = true;
  }

  void updateLocalizations(AppLocalizations l10n) {
    _l10n = l10n;
  }

  Future<void> enableForSession() async {
    if (!_started) {
      return;
    }
    final token = await AuthStorage.getToken();
    if (token == null || token.isEmpty) {
      return;
    }
    await _local.requestPermission();
    try {
      final info = await _api.getServerInfo();
      _vapidPublicKey = info.vapidPublicKey;
    } catch (err) {
      debugPrint('notifications: server-info failed: $err');
      return;
    }
    if ((_vapidPublicKey ?? '').isEmpty) {
      debugPrint('notifications: missing vapid public key');
      return;
    }
    await _transport.register(vapidPublicKey: _vapidPublicKey!);
  }

  Future<void> disableForSession() async {
    final token = await AuthStorage.getToken();
    final installationId = await _installationIds.getOrCreate();
    try {
      await _transport.unregister();
    } catch (err) {
      debugPrint('notifications: unregister transport failed: $err');
    }
    if (token != null && token.isNotEmpty) {
      try {
        await _api.deletePushSubscription(
          token: token,
          installationId: installationId,
        );
      } catch (err) {
        debugPrint('notifications: delete subscription failed: $err');
      }
    }
    _aggregator.clearAll();
    await _local.cancelAll();
  }

  void clearWorkoutSlots({required String owner, required String workoutId}) {
    final before = List<NotificationSlotState>.from(
      [
        for (final slot in [
          'liked:$owner:$workoutId',
          'commented:$owner:$workoutId',
        ])
          if (_aggregator.get(slot) != null) _aggregator.get(slot)!,
      ],
    );
    _aggregator.clearWorkout(owner, workoutId);
    for (final state in before) {
      _local.cancelSlot(state.slot);
    }
  }

  void _handleMessage(PushNotificationPayload payload) {
    // If the OS notification was dismissed, restart the collapse counter.
    final existing = _aggregator.get(payload.slot);
    if (existing != null) {
      final id = SlotAggregator.notificationIdForSlot(payload.slot);
      _local.activeNotificationIds().then((active) {
        if (!active.contains(id)) {
          _aggregator.clearSlot(payload.slot);
        }
        final state = _aggregator.ingest(payload);
        if (state != null) {
          _local.showOrUpdate(state);
        }
      });
      return;
    }
    final state = _aggregator.ingest(payload);
    if (state == null) {
      return;
    }
    _local.showOrUpdate(state);
  }

  Future<void> _registerEndpoint({
    required String endpoint,
    required String p256dh,
    required String auth,
  }) async {
    final token = await AuthStorage.getToken();
    if (token == null || token.isEmpty) {
      return;
    }
    final installationId = await _installationIds.getOrCreate();
    try {
      await _api.registerPushSubscription(
        token: token,
        installationId: installationId,
        endpoint: endpoint,
        p256dh: p256dh,
        auth: auth,
        platform: 'android',
      );
    } catch (err) {
      debugPrint('notifications: register subscription failed: $err');
    }
  }

  void _handleTap(PushNotificationPayload payload) {
    _aggregator.clearSlot(payload.slot);
    _local.cancelSlot(payload.slot);
    final open = onOpenWorkout;
    if (open == null) {
      return;
    }
    open(workoutId: payload.workoutId, owner: payload.owner);
  }

  String _formatBody(NotificationSlotState state) {
    final l10n = _l10n;
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
}
