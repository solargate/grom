import 'package:flutter_test/flutter_test.dart';
import 'package:grom/server_catalog.dart';
import 'package:grom/server_history.dart';
import 'package:grom/server_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ServerHistory.clear();
  });

  test('remember stores LRU unique URLs newest first', () async {
    await ServerHistory.remember('https://one.example/');
    await ServerHistory.remember('https://two.example');
    await ServerHistory.remember('https://one.example');

    expect(
      await ServerHistory.getRecent(),
      ['https://one.example', 'https://two.example'],
    );
  });

  test('remember caps history at the limit', () async {
    for (var i = 0; i < serverUrlHistoryLimit + 5; i++) {
      await ServerHistory.remember('https://s$i.example');
    }
    final recent = await ServerHistory.getRecent();
    expect(recent, hasLength(serverUrlHistoryLimit));
    expect(recent.first, 'https://s${serverUrlHistoryLimit + 4}.example');
    expect(recent.last, 'https://s5.example');
  });

  test('recentServersNotInCatalog skips approved URLs', () {
    const catalog = [
      CatalogServer(
        url: 'https://approved.example',
        name: 'Approved',
        description: 'Listed',
      ),
    ];
    expect(
      recentServersNotInCatalog(
        [
          'https://approved.example/',
          'https://custom.example',
        ],
        catalog,
      ),
      ['https://custom.example'],
    );
  });

  test('recentServersNotInCatalog uses generated catalog by default', () {
    expect(
      recentServersNotInCatalog(['https://custom.example']),
      ['https://custom.example'],
    );
  });

  test('normalize matches catalog path stripping', () {
    expect(
      ServerStorage.normalizeBaseUrl('https://example.org/grom/'),
      'https://example.org/grom',
    );
  });
}
