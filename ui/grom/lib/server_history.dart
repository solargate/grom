import 'package:shared_preferences/shared_preferences.dart';

import 'server_catalog.dart';
import 'server_storage.dart';

const serverUrlHistoryStorageKey = 'server_url_history';
const serverUrlHistoryLimit = 20;

class ServerHistory {
  static Future<List<String>> getRecent() async {
    final prefs = await SharedPreferences.getInstance();
    return List<String>.from(
      prefs.getStringList(serverUrlHistoryStorageKey) ?? const <String>[],
    );
  }

  static Future<void> remember(String url) async {
    final normalized = ServerStorage.normalizeBaseUrl(url);
    if (normalized.isEmpty) {
      return;
    }

    final prefs = await SharedPreferences.getInstance();
    final existing = List<String>.from(
      prefs.getStringList(serverUrlHistoryStorageKey) ?? const <String>[],
    );
    existing.remove(normalized);
    existing.insert(0, normalized);
    if (existing.length > serverUrlHistoryLimit) {
      existing.removeRange(serverUrlHistoryLimit, existing.length);
    }
    await prefs.setStringList(serverUrlHistoryStorageKey, existing);
  }

  static Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(serverUrlHistoryStorageKey);
  }
}

List<String> recentServersNotInCatalog(
  List<String> recent, [
  List<CatalogServer> catalog = kApprovedServers,
]) {
  final approved = <String>{
    for (final server in catalog) ServerStorage.normalizeBaseUrl(server.url),
  };
  return [
    for (final url in recent)
      if (!approved.contains(ServerStorage.normalizeBaseUrl(url))) url,
  ];
}
