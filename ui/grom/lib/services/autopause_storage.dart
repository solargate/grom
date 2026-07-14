import 'package:shared_preferences/shared_preferences.dart';

const autoPauseEnabledStorageKey = 'auto_pause_enabled';

class AutopauseStorage {
  static Future<bool> getEnabled({bool defaultValue = true}) async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getBool(autoPauseEnabledStorageKey) ?? defaultValue;
  }

  static Future<void> setEnabled(bool enabled) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(autoPauseEnabledStorageKey, enabled);
  }
}
