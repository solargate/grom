import 'package:shared_preferences/shared_preferences.dart';

const tokenStorageKey = 'auth_token';
const nicknameStorageKey = 'auth_nickname';

class AuthStorage {
  static Future<void> saveSession({
    required String token,
    required String nickname,
  }) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(tokenStorageKey, token);
    await prefs.setString(nicknameStorageKey, nickname);
  }

  static Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(tokenStorageKey);
  }

  static Future<String?> getNickname() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(nicknameStorageKey);
  }

  static Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(tokenStorageKey);
    await prefs.remove(nicknameStorageKey);
  }
}
