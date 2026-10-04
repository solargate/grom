import 'dart:io' show Platform;

import 'push_transport.dart';
import 'push_transport_android.dart';
import 'push_transport_stub.dart';

PushTransport createPushTransport() {
  if (Platform.isAndroid) {
    return UnifiedPushTransport();
  }
  return StubPushTransport();
}
