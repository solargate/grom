import 'package:flutter/foundation.dart';

class AvatarCache extends ChangeNotifier {
  AvatarCache._();

  static final AvatarCache instance = AvatarCache._();

  final Map<String, int> _versions = {};

  int versionFor(String nickname) => _versions[nickname] ?? 0;

  void bump(String nickname) {
    _versions[nickname] = versionFor(nickname) + 1;
    notifyListeners();
  }

  static String withCacheBuster(String url, int version) {
    if (url.isEmpty || version <= 0) {
      return url;
    }
    final separator = url.contains('?') ? '&' : '?';
    return '$url${separator}v=$version';
  }
}
