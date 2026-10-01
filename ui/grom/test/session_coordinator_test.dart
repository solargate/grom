import 'package:flutter_test/flutter_test.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/session.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    SessionCoordinator.instance.debugReset();
  });

  tearDown(() async {
    SessionCoordinator.instance.debugReset();
    await AuthStorage.clear();
  });

  test('notifyUnauthorized clears token and emits once when enabled', () async {
    await AuthStorage.saveToken('jwt-token');
    SessionCoordinator.instance.emitExpiredEvents = true;

    final events = <SessionEndReason>[];
    final sub = SessionCoordinator.instance.events.listen(events.add);

    await SessionCoordinator.instance.notifyUnauthorized();
    await SessionCoordinator.instance.notifyUnauthorized();
    await Future<void>.delayed(Duration.zero);

    expect(await AuthStorage.getToken(), isNull);
    expect(events, [SessionEndReason.expired]);

    await sub.cancel();
  });

  test('notifyUnauthorized clears token without emit when disabled', () async {
    await AuthStorage.saveToken('jwt-token');
    SessionCoordinator.instance.emitExpiredEvents = false;

    final events = <SessionEndReason>[];
    final sub = SessionCoordinator.instance.events.listen(events.add);

    await SessionCoordinator.instance.notifyUnauthorized();
    await Future<void>.delayed(Duration.zero);

    expect(await AuthStorage.getToken(), isNull);
    expect(events, isEmpty);

    await sub.cancel();
  });

  test('clearLocalSession keeps last email', () async {
    await AuthStorage.saveToken('jwt-token');
    await AuthStorage.saveLastEmail('alice@example.com');

    await clearLocalSession();

    expect(await AuthStorage.getToken(), isNull);
    expect(await AuthStorage.getLastEmail(), 'alice@example.com');
  });

  test('resetUnauthorizedGuard allows another expiry wave', () async {
    await AuthStorage.saveToken('jwt-1');
    SessionCoordinator.instance.emitExpiredEvents = true;

    final events = <SessionEndReason>[];
    final sub = SessionCoordinator.instance.events.listen(events.add);

    await SessionCoordinator.instance.notifyUnauthorized();
    SessionCoordinator.instance.resetUnauthorizedGuard();
    await AuthStorage.saveToken('jwt-2');
    await SessionCoordinator.instance.notifyUnauthorized();
    await Future<void>.delayed(Duration.zero);

    expect(events, [SessionEndReason.expired, SessionEndReason.expired]);

    await sub.cancel();
  });
}
