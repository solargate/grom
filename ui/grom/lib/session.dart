import 'dart:async';

import 'auth_storage.dart';

enum SessionEndReason { expired, manual }

/// Coordinates JWT clear + a single UI signal when the session JWT is rejected.
class SessionCoordinator {
  SessionCoordinator._();

  static final SessionCoordinator instance = SessionCoordinator._();

  final StreamController<SessionEndReason> _events =
      StreamController<SessionEndReason>.broadcast();

  /// When false (cold start), unauthorized clears the token but does not emit.
  bool emitExpiredEvents = false;

  bool _handlingUnauthorized = false;

  Stream<SessionEndReason> get events => _events.stream;

  /// Clears the JWT. Emits [SessionEndReason.expired] at most once per wave
  /// while [emitExpiredEvents] is true.
  Future<void> notifyUnauthorized() async {
    if (_handlingUnauthorized) {
      return;
    }
    _handlingUnauthorized = true;
    await clearLocalSession();
    if (emitExpiredEvents) {
      _events.add(SessionEndReason.expired);
    } else {
      _handlingUnauthorized = false;
    }
  }

  /// Call after a successful login so a later expiry can notify again.
  void resetUnauthorizedGuard() {
    _handlingUnauthorized = false;
  }

  /// Test helper: reset singleton state between tests.
  void debugReset() {
    _handlingUnauthorized = false;
    emitExpiredEvents = false;
  }
}

/// Clears the Grom JWT from local storage (does not clear last email).
Future<void> clearLocalSession() async {
  await AuthStorage.clear();
}
