import 'package:shared_preferences/shared_preferences.dart';

const tokenStorageKey = 'auth_token';
const lastEmailStorageKey = 'auth_last_email';
const _legacyNicknameStorageKey = 'auth_nickname';

class AuthStorage {
  static Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(tokenStorageKey, token);
  }

  static Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(tokenStorageKey);
  }

  static Future<void> saveLastEmail(String email) async {
    final trimmed = email.trim();
    if (trimmed.isEmpty) {
      return;
    }
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(lastEmailStorageKey, trimmed);
  }

  static Future<String?> getLastEmail() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(lastEmailStorageKey);
  }

  /// Clears the JWT (and legacy nickname). Keeps [lastEmailStorageKey].
  static Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(tokenStorageKey);
    await prefs.remove(_legacyNicknameStorageKey);
  }
}
