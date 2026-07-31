import 'package:shared_preferences/shared_preferences.dart';

const healthSyncEnabledStorageKey = 'health_sync_enabled';
const healthSyncFolderNameStorageKey = 'health_sync_folder_name';
const healthSyncFolderIdStorageKey = 'health_sync_folder_id';

class HealthSyncStorage {
  static Future<bool> loadEnabled({bool defaultValue = false}) async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getBool(healthSyncEnabledStorageKey) ?? defaultValue;
  }

  static Future<void> saveEnabled(bool enabled) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(healthSyncEnabledStorageKey, enabled);
  }

  static Future<({String name, String id})> loadFolder() async {
    final prefs = await SharedPreferences.getInstance();
    return (
      name: prefs.getString(healthSyncFolderNameStorageKey) ?? '',
      id: prefs.getString(healthSyncFolderIdStorageKey) ?? '',
    );
  }

  static Future<void> saveFolder({required String name, required String id}) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(healthSyncFolderNameStorageKey, name);
    await prefs.setString(healthSyncFolderIdStorageKey, id);
  }

  static Future<void> saveFolderName(String name) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(healthSyncFolderNameStorageKey, name);
  }

  static Future<void> clearFolderId() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(healthSyncFolderIdStorageKey);
  }
}
