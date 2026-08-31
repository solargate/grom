import 'package:flutter_test/flutter_test.dart';
import 'package:grom/platform/google_drive.dart';
import 'package:grom/services/health_sync_service.dart';
import 'package:grom/services/health_sync_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  test('resetForLogout clears enabled state and storage', () async {
    await HealthSyncStorage.saveEnabled(true);
    await HealthSyncStorage.saveFolder(name: 'Health Sync', id: 'folder-1');

    final service = HealthSyncService.instance;
    await service.loadFromStorage();
    expect(service.enabled, isTrue);
    expect(service.folderId, 'folder-1');

    await service.resetForLogout();
    expect(service.enabled, isFalse);
    expect(service.folderId, isEmpty);
    expect(await HealthSyncStorage.loadEnabled(), isFalse);
    expect((await HealthSyncStorage.loadFolder()).id, isEmpty);
  });

  test('healthSyncResultFromDriveError maps cancelled sign-in', () {
    final result = healthSyncResultFromDriveError(
      GoogleDriveException(GoogleDriveError.cancelled, message: 'cancelled'),
    );
    expect(result.kind, HealthSyncResultKind.signInCancelled);
  });

  test('healthSyncResultFromDriveError maps access denied', () {
    final result = healthSyncResultFromDriveError(
      GoogleDriveException(GoogleDriveError.accessDenied, message: 'denied'),
    );
    expect(result.kind, HealthSyncResultKind.accessDenied);
  });
}
