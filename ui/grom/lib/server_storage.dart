import 'package:shared_preferences/shared_preferences.dart';

import 'session.dart';

const serverBaseUrlStorageKey = 'server_base_url';

class ServerStorage {
  static String? _cachedBaseUrl;

  static String? get cachedBaseUrl => _cachedBaseUrl;

  static Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    _cachedBaseUrl = prefs.getString(serverBaseUrlStorageKey);
  }

  static Future<String?> getBaseUrl() async {
    if (_cachedBaseUrl != null) {
      return _cachedBaseUrl;
    }
    final prefs = await SharedPreferences.getInstance();
    _cachedBaseUrl = prefs.getString(serverBaseUrlStorageKey);
    return _cachedBaseUrl;
  }

  static Future<void> saveBaseUrl(String url) async {
    final normalized = normalizeBaseUrl(url);
    if (normalized == _cachedBaseUrl) {
      return;
    }

    if (_cachedBaseUrl != null) {
      await clearLocalSession();
    }

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(serverBaseUrlStorageKey, normalized);
    _cachedBaseUrl = normalized;
  }

  static Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(serverBaseUrlStorageKey);
    _cachedBaseUrl = null;
  }

  static String normalizeBaseUrl(String url) {
    var trimmed = url.trim();
    if (trimmed.isEmpty) {
      return trimmed;
    }

    if (!trimmed.contains('://')) {
      trimmed = 'https://$trimmed';
    }

    final uri = Uri.parse(trimmed);
    final scheme = uri.scheme;
    final host = uri.host;
    if (host.isEmpty) {
      return trimmed;
    }

    final buffer = StringBuffer('$scheme://$host');
    if (uri.hasPort) {
      buffer.write(':${uri.port}');
    }

    var path = uri.path;
    while (path.endsWith('/')) {
      path = path.substring(0, path.length - 1);
    }
    if (path.isNotEmpty && path != '/') {
      buffer.write(path);
    }

    return buffer.toString();
  }

  static bool isValidBaseUrl(String url) {
    final normalized = normalizeBaseUrl(url);
    if (normalized.isEmpty) {
      return false;
    }

    final uri = Uri.tryParse(normalized);
    if (uri == null || uri.host.isEmpty) {
      return false;
    }

    return uri.scheme == 'http' || uri.scheme == 'https';
  }
}
