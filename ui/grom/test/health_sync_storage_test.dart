import 'package:flutter_test/flutter_test.dart';
import 'package:grom/platform/google_drive.dart';
import 'package:grom/services/health_sync_service.dart';
import 'package:grom/services/health_sync_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
  });

  group('HealthSyncStorage', () {
    test('round-trips enabled flag and folder', () async {
      expect(await HealthSyncStorage.loadEnabled(), isFalse);

      await HealthSyncStorage.saveEnabled(true);
      expect(await HealthSyncStorage.loadEnabled(), isTrue);

      await HealthSyncStorage.saveFolder(name: 'Health Sync', id: 'folder-1');
      final folder = await HealthSyncStorage.loadFolder();
      expect(folder.name, 'Health Sync');
      expect(folder.id, 'folder-1');

      await HealthSyncStorage.saveFolderName('Renamed');
      await HealthSyncStorage.clearFolderId();
      final cleared = await HealthSyncStorage.loadFolder();
      expect(cleared.name, 'Renamed');
      expect(cleared.id, isEmpty);
    });
  });

  group('healthSyncResultFromDriveError', () {
    test('maps known Google Drive errors', () {
      expect(
        healthSyncResultFromDriveError(
          GoogleDriveException(GoogleDriveError.cancelled),
        ).kind,
        HealthSyncResultKind.signInCancelled,
      );
      expect(
        healthSyncResultFromDriveError(
          GoogleDriveException(GoogleDriveError.signInFailed),
        ).kind,
        HealthSyncResultKind.signInFailed,
      );
      expect(
        healthSyncResultFromDriveError(
          GoogleDriveException(GoogleDriveError.accessDenied),
        ).kind,
        HealthSyncResultKind.accessDenied,
      );
    });

    test('preserves error message for sign-in failures', () {
      final result = healthSyncResultFromDriveError(
        GoogleDriveException(
          GoogleDriveError.signInFailed,
          message: 'authenticatedClient returned null',
        ),
      );
      expect(result.kind, HealthSyncResultKind.signInFailed);
      expect(result.message, 'authenticatedClient returned null');
    });

    test('preserves error message for access denied', () {
      final result = healthSyncResultFromDriveError(
        GoogleDriveException(
          GoogleDriveError.accessDenied,
          message: 'scope denied',
        ),
      );
      expect(result.kind, HealthSyncResultKind.accessDenied);
      expect(result.message, 'scope denied');
    });

    test('maps unsupported to error with message', () {
      final result = healthSyncResultFromDriveError(
        GoogleDriveException(
          GoogleDriveError.unsupported,
          message: 'no drive',
        ),
      );
      expect(result.kind, HealthSyncResultKind.error);
      expect(result.message, 'no drive');
    });
  });

  group('healthSyncResultSnackBarMessage', () {
    test('appends detail for sign-in failed', () {
      final message = healthSyncResultSnackBarMessage(
        const HealthSyncResult(
          kind: HealthSyncResultKind.signInFailed,
          message: 'ApiException: 10',
        ),
        imported: 'imported',
        noNewWorkouts: 'none',
        folderNotFound: 'missing',
        folderEmpty: 'empty',
        signInCancelled: 'cancelled',
        signInFailed: 'sign-in failed',
        accessDenied: 'denied',
        syncError: (detail) => 'error: $detail',
      );
      expect(message, 'sign-in failed: ApiException: 10');
    });

    test('omits empty detail', () {
      final message = healthSyncResultSnackBarMessage(
        const HealthSyncResult(kind: HealthSyncResultKind.signInFailed),
        imported: 'imported',
        noNewWorkouts: 'none',
        folderNotFound: 'missing',
        folderEmpty: 'empty',
        signInCancelled: 'cancelled',
        signInFailed: 'sign-in failed',
        accessDenied: 'denied',
        syncError: (detail) => 'error: $detail',
      );
      expect(message, 'sign-in failed');
    });
  });
}
