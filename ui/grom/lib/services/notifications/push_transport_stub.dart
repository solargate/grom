import 'push_transport.dart';

class StubPushTransport implements PushTransport {
  @override
  Future<void> initialize({
    required PushEndpointHandler onEndpoint,
    required PushMessageHandler onMessage,
  }) async {}

  @override
  Future<void> register({required String vapidPublicKey}) async {}

  @override
  Future<void> unregister() async {}
}

PushTransport createPushTransport() => StubPushTransport();
