import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:unifiedpush/unifiedpush.dart';

import 'push_payload.dart';
import 'push_transport.dart';

const _gromPushInstance = 'grom';

class UnifiedPushTransport implements PushTransport {
  PushEndpointHandler? _onEndpoint;
  PushMessageHandler? _onMessage;
  bool _initialized = false;

  @override
  Future<void> initialize({
    required PushEndpointHandler onEndpoint,
    required PushMessageHandler onMessage,
  }) async {
    _onEndpoint = onEndpoint;
    _onMessage = onMessage;
    if (_initialized) {
      return;
    }
    await UnifiedPush.initialize(
      onNewEndpoint: _handleNewEndpoint,
      onRegistrationFailed: (reason, instance) {
        debugPrint('UnifiedPush registration failed: $reason ($instance)');
      },
      onUnregistered: (instance) {
        debugPrint('UnifiedPush unregistered: $instance');
      },
      onMessage: _handleMessage,
    );
    _initialized = true;
  }

  @override
  Future<void> register({required String vapidPublicKey}) async {
    final vapid = vapidPublicKey.trim();
    final ok = await UnifiedPush.tryUseCurrentOrDefaultDistributor();
    if (!ok) {
      final distributors = await UnifiedPush.getDistributors();
      if (distributors.isEmpty) {
        debugPrint('UnifiedPush: no distributor available');
        return;
      }
      await UnifiedPush.saveDistributor(distributors.first);
    }
    await UnifiedPush.register(
      instance: _gromPushInstance,
      vapid: vapid.isEmpty ? null : vapid,
    );
  }

  @override
  Future<void> unregister() async {
    await UnifiedPush.unregister(_gromPushInstance);
  }

  Future<void> _handleNewEndpoint(PushEndpoint endpoint, String instance) async {
    final handler = _onEndpoint;
    if (handler == null) {
      return;
    }
    final keys = endpoint.pubKeySet;
    if (keys == null) {
      debugPrint('UnifiedPush endpoint missing pubKeySet');
      return;
    }
    await handler(
      endpoint: endpoint.url,
      p256dh: keys.pubKey,
      auth: keys.auth,
    );
  }

  void _handleMessage(PushMessage message, String instance) {
    final handler = _onMessage;
    if (handler == null) {
      return;
    }
    try {
      final raw = utf8.decode(message.content);
      final json = jsonDecode(raw) as Map<String, dynamic>;
      handler(PushNotificationPayload.fromJson(json));
    } catch (err) {
      debugPrint('UnifiedPush message parse failed: $err');
    }
  }
}
