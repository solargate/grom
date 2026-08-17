import 'auth_storage.dart';
import 'services/health_sync_service.dart';

/// Clears the Grom JWT and device-local Health Sync / Google Drive session.
Future<void> clearLocalSession() async {
  await AuthStorage.clear();
  await HealthSyncService.instance.resetForLogout();
}
