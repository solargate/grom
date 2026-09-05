import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_auth.dart';
import 'package:grom/services/strava_api_constants.dart';
import 'package:grom/services/strava_api_storage.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  test('buildAuthorizeUrl uses custom scheme redirect and required scopes', () {
    final uri = StravaApiAuth().buildAuthorizeUrl(clientId: '42');
    expect(uri.scheme, 'https');
    expect(uri.host, 'www.strava.com');
    expect(uri.path, '/oauth/authorize');
    expect(uri.queryParameters['client_id'], '42');
    expect(uri.queryParameters['redirect_uri'], kStravaOAuthRedirectUri);
    expect(uri.queryParameters['response_type'], 'code');
    expect(uri.queryParameters['approval_prompt'], 'force');
    expect(uri.queryParameters['scope'], kStravaOAuthScope);
  });

  test('connect returns missingCredentials when id or secret empty', () async {
    final auth = StravaApiAuth(
      authenticate: (_) async => fail('oauth should not run'),
    );
    expect(
      (await auth.connect(clientId: '', clientSecret: 's')).kind,
      StravaConnectResultKind.missingCredentials,
    );
    expect(
      (await auth.connect(clientId: '1', clientSecret: '  ')).kind,
      StravaConnectResultKind.missingCredentials,
    );
  });

  test('connect exchanges code with JSON body and no redirect_uri', () async {
    http.Request? tokenRequest;
    final auth = StravaApiAuth(
      httpClient: MockClient((request) async {
        tokenRequest = request;
        expect(request.url.toString(), kStravaTokenUrl);
        expect(request.headers['content-type'], contains('application/json'));
        expect(request.headers['user-agent'], kStravaHttpUserAgent);
        return http.Response(
          jsonEncode({
            'access_token': 'access',
            'refresh_token': 'refresh',
            'expires_at': 1700000000,
            'athlete': {'id': 7},
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      }),
      authenticate: (_) async =>
          'grom://localhost?code=abc&scope=read,activity:read',
    );

    final result = await auth.connect(clientId: '1', clientSecret: 'secret');
    expect(result.kind, StravaConnectResultKind.connected);

    final body = jsonDecode(tokenRequest!.body) as Map<String, dynamic>;
    expect(body['grant_type'], 'authorization_code');
    expect(body['code'], 'abc');
    expect(body['client_id'], '1');
    expect(body['client_secret'], 'secret');
    expect(body.containsKey('redirect_uri'), isFalse);

    final tokens = await StravaApiStorage.loadTokens();
    expect(tokens.connected, isTrue);
    expect(tokens.accessToken, 'access');
    expect(tokens.refreshToken, 'refresh');
    expect(tokens.athleteId, 7);
  });

  test('connect accepts activity:read_all scope', () async {
    final auth = StravaApiAuth(
      httpClient: MockClient((_) async => http.Response(
            jsonEncode({
              'access_token': 'a',
              'refresh_token': 'r',
              'expires_at': 1700000000,
            }),
            200,
          )),
      authenticate: (_) async =>
          'grom://localhost?code=x&scope=read,activity:read_all',
    );
    expect(
      (await auth.connect(clientId: '1', clientSecret: 's')).kind,
      StravaConnectResultKind.connected,
    );
  });

  test('connect returns missingScope without activity:read', () async {
    final auth = StravaApiAuth(
      httpClient: MockClient((_) async => fail('token exchange must not run')),
      authenticate: (_) async => 'grom://localhost?code=x&scope=read',
    );
    expect(
      (await auth.connect(clientId: '1', clientSecret: 's')).kind,
      StravaConnectResultKind.missingScope,
    );
  });

  test('connect maps access_denied and cancel', () async {
    final denied = StravaApiAuth(
      authenticate: (_) async => 'grom://localhost?error=access_denied',
    );
    expect(
      (await denied.connect(clientId: '1', clientSecret: 's')).kind,
      StravaConnectResultKind.denied,
    );

    final cancelled = StravaApiAuth(
      authenticate: (_) async => throw Exception('User cancelled the flow'),
    );
    expect(
      (await cancelled.connect(clientId: '1', clientSecret: 's')).kind,
      StravaConnectResultKind.cancelled,
    );
  });

  test('connect surfaces VPN hint on token 403 HTML', () async {
    final auth = StravaApiAuth(
      httpClient: MockClient(
        (_) async => http.Response('<html>cloudflare</html>', 403),
      ),
      authenticate: (_) async =>
          'grom://localhost?code=abc&scope=activity:read',
    );
    final result = await auth.connect(clientId: '1', clientSecret: 's');
    expect(result.kind, StravaConnectResultKind.error);
    expect(result.message, contains('disable VPN'));
    expect(result.message, contains('403'));
  });

  test('ensureAccessToken refreshes when near expiry', () async {
    await StravaApiStorage.saveCredentials(
      clientId: '1',
      clientSecret: 'secret',
    );
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
    await StravaApiStorage.saveTokens(
      accessToken: 'old',
      refreshToken: 'refresh',
      expiresAt: now + 30,
      connected: true,
    );

    http.Request? refreshRequest;
    final auth = StravaApiAuth(
      httpClient: MockClient((request) async {
        refreshRequest = request;
        return http.Response(
          jsonEncode({
            'access_token': 'new-access',
            'refresh_token': 'new-refresh',
            'expires_at': now + 3600,
          }),
          200,
        );
      }),
    );

    expect(await auth.ensureAccessToken(), 'new-access');
    final body = jsonDecode(refreshRequest!.body) as Map<String, dynamic>;
    expect(body['grant_type'], 'refresh_token');
    expect(body['refresh_token'], 'refresh');
    expect(body.containsKey('redirect_uri'), isFalse);
    expect(await StravaApiStorage.loadTokens().then((t) => t.accessToken),
        'new-access');
  });

  test('ensureAccessToken reuses valid access token', () async {
    await StravaApiStorage.saveCredentials(
      clientId: '1',
      clientSecret: 'secret',
    );
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
    await StravaApiStorage.saveTokens(
      accessToken: 'still-valid',
      refreshToken: 'refresh',
      expiresAt: now + 3600,
      connected: true,
    );

    final auth = StravaApiAuth(
      httpClient: MockClient((_) async => fail('refresh should not run')),
    );
    expect(await auth.ensureAccessToken(), 'still-valid');
  });
}
