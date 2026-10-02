import 'push_payload.dart';

typedef PushEndpointHandler = Future<void> Function({
  required String endpoint,
  required String p256dh,
  required String auth,
});

typedef PushMessageHandler = void Function(PushNotificationPayload payload);

/// Platform push transport. Android uses UnifiedPush; other platforms are stubs.
abstract class PushTransport {
  Future<void> initialize({
    required PushEndpointHandler onEndpoint,
    required PushMessageHandler onMessage,
  });

  Future<void> register({required String vapidPublicKey});

  Future<void> unregister();
}
