import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_constants.dart';
import 'package:grom/services/strava_api_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  test('persists sync limit with default and clamp', () async {
    expect(await StravaApiStorage.loadSyncLimit(), kStravaApiSyncLimitDefault);

    await StravaApiStorage.saveSyncLimit(25);
    expect(await StravaApiStorage.loadSyncLimit(), 25);

    await StravaApiStorage.saveSyncLimit(0);
    expect(await StravaApiStorage.loadSyncLimit(), kStravaApiSyncLimitDefault);

    await StravaApiStorage.saveSyncLimit(500);
    expect(await StravaApiStorage.loadSyncLimit(), kStravaApiSyncLimitMax);
  });

  test('persists tokens and connected state', () async {
    await StravaApiStorage.saveTokens(
      accessToken: 'access',
      refreshToken: 'refresh',
      expiresAt: 1700000000,
      connected: true,
      athleteId: 99,
      scope: 'activity:read',
    );

    final tokens = await StravaApiStorage.loadTokens();
    expect(tokens.accessToken, 'access');
    expect(tokens.refreshToken, 'refresh');
    expect(tokens.expiresAt, 1700000000);
    expect(tokens.connected, isTrue);
    expect(tokens.athleteId, 99);
    expect(tokens.scope, 'activity:read');
  });

  test('clearTokens keeps credentials and enabled', () async {
    await StravaApiStorage.saveEnabled(true);
    await StravaApiStorage.saveCredentials(
      clientId: '1',
      clientSecret: 's',
    );
    await StravaApiStorage.saveTokens(
      accessToken: 'a',
      refreshToken: 'r',
      expiresAt: 1,
      connected: true,
    );

    await StravaApiStorage.clearTokens();

    expect(await StravaApiStorage.loadEnabled(), isTrue);
    expect(await StravaApiStorage.loadClientId(), '1');
    final tokens = await StravaApiStorage.loadTokens();
    expect(tokens.connected, isFalse);
    expect(tokens.accessToken, isEmpty);
    expect(tokens.refreshToken, isEmpty);
  });
}
