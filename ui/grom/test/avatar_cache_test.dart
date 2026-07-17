import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/avatar_cache.dart';

void main() {
  test('withCacheBuster preserves empty URLs and non-positive versions', () {
    expect(AvatarCache.withCacheBuster('', 1), '');
    expect(AvatarCache.withCacheBuster('https://example.test/avatar.png', 0),
        'https://example.test/avatar.png');
  });

  test('withCacheBuster appends version using the appropriate separator', () {
    expect(
      AvatarCache.withCacheBuster('https://example.test/avatar.png', 2),
      'https://example.test/avatar.png?v=2',
    );
    expect(
      AvatarCache.withCacheBuster('https://example.test/avatar.png?size=64', 3),
      'https://example.test/avatar.png?size=64&v=3',
    );
  });

  test('bump increments a nickname cache version', () {
    final cache = AvatarCache.instance;
    final nickname = 'test-user-${DateTime.now().microsecondsSinceEpoch}';

    expect(cache.versionFor(nickname), 0);
    cache.bump(nickname);
    cache.bump(nickname);

    expect(cache.versionFor(nickname), 2);
  });
}
