import 'dart:convert';

import 'package:http/http.dart' as http;

import '../platform/strava_oauth.dart';
import 'strava_api_constants.dart';
import 'strava_api_storage.dart';

enum StravaConnectResultKind {
  connected,
  cancelled,
  missingCredentials,
  missingScope,
  denied,
  error,
}

class StravaConnectResult {
  const StravaConnectResult({
    required this.kind,
    this.message = '',
  });

  final StravaConnectResultKind kind;
  final String message;
}

class StravaApiAuth {
  StravaApiAuth({http.Client? httpClient})
      : _http = httpClient ?? http.Client();

  final http.Client _http;

  Uri buildAuthorizeUrl({required String clientId}) {
    return Uri.parse(kStravaAuthorizeUrl).replace(
      queryParameters: {
        'client_id': clientId,
        'redirect_uri': kStravaOAuthRedirectUri,
        'response_type': 'code',
        'approval_prompt': 'auto',
        'scope': kStravaOAuthScope,
      },
    );
  }

  Future<StravaConnectResult> connect({
    required String clientId,
    required String clientSecret,
  }) async {
    final id = clientId.trim();
    final secret = clientSecret.trim();
    if (id.isEmpty || secret.isEmpty) {
      return const StravaConnectResult(
        kind: StravaConnectResultKind.missingCredentials,
      );
    }

    await StravaApiStorage.saveCredentials(
      clientId: id,
      clientSecret: secret,
    );

    try {
      final redirectUrl = await authenticateStravaOAuth(
        buildAuthorizeUrl(clientId: id).toString(),
      );
      final redirect = Uri.parse(redirectUrl);
      final error = redirect.queryParameters['error'];
      if (error != null && error.isNotEmpty) {
        if (error == 'access_denied') {
          return const StravaConnectResult(kind: StravaConnectResultKind.denied);
        }
        return StravaConnectResult(
          kind: StravaConnectResultKind.error,
          message: error,
        );
      }

      final code = redirect.queryParameters['code'];
      if (code == null || code.isEmpty) {
        return const StravaConnectResult(
          kind: StravaConnectResultKind.error,
          message: 'missing authorization code',
        );
      }

      final grantedScope = redirect.queryParameters['scope'] ?? '';
      if (!_scopeIncludesActivityRead(grantedScope)) {
        return const StravaConnectResult(
          kind: StravaConnectResultKind.missingScope,
        );
      }

      final tokens = await _exchangeCode(
        clientId: id,
        clientSecret: secret,
        code: code,
      );
      await StravaApiStorage.saveTokens(
        accessToken: tokens.accessToken,
        refreshToken: tokens.refreshToken,
        expiresAt: tokens.expiresAt,
        connected: true,
        athleteId: tokens.athleteId,
        scope: grantedScope.isNotEmpty ? grantedScope : kStravaOAuthScope,
      );
      return const StravaConnectResult(kind: StravaConnectResultKind.connected);
    } on Exception catch (error) {
      final text = error.toString().toLowerCase();
      if (text.contains('cancel') || text.contains('cancell')) {
        return const StravaConnectResult(
          kind: StravaConnectResultKind.cancelled,
        );
      }
      return StravaConnectResult(
        kind: StravaConnectResultKind.error,
        message: error.toString(),
      );
    }
  }

  /// Returns a valid access token, refreshing when near expiry.
  Future<String?> ensureAccessToken() async {
    final clientId = (await StravaApiStorage.loadClientId()).trim();
    final clientSecret = (await StravaApiStorage.loadClientSecret()).trim();
    final tokens = await StravaApiStorage.loadTokens();
    if (!tokens.connected ||
        tokens.refreshToken.isEmpty ||
        clientId.isEmpty ||
        clientSecret.isEmpty) {
      return null;
    }

    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
    // Refresh one minute early.
    if (tokens.accessToken.isNotEmpty && tokens.expiresAt > now + 60) {
      return tokens.accessToken;
    }

    final refreshed = await _refreshToken(
      clientId: clientId,
      clientSecret: clientSecret,
      refreshToken: tokens.refreshToken,
    );
    await StravaApiStorage.saveTokens(
      accessToken: refreshed.accessToken,
      refreshToken: refreshed.refreshToken,
      expiresAt: refreshed.expiresAt,
      connected: true,
      athleteId: tokens.athleteId,
      scope: tokens.scope,
    );
    return refreshed.accessToken;
  }

  bool _scopeIncludesActivityRead(String scope) {
    final parts = scope
        .split(RegExp(r'[\s,]+'))
        .map((s) => s.trim())
        .where((s) => s.isNotEmpty)
        .toSet();
    // activity:read_all also covers public/followers activities.
    return parts.contains('activity:read') ||
        parts.contains('activity:read_all');
  }

  Future<
      ({
        String accessToken,
        String refreshToken,
        int expiresAt,
        int? athleteId,
      })> _exchangeCode({
    required String clientId,
    required String clientSecret,
    required String code,
  }) async {
    final response = await _http.post(
      Uri.parse(kStravaTokenUrl),
      body: {
        'client_id': clientId,
        'client_secret': clientSecret,
        'code': code,
        'grant_type': 'authorization_code',
      },
    );
    if (response.statusCode != 200) {
      throw Exception(_tokenError(response));
    }
    return _parseTokenResponse(response.body);
  }

  Future<
      ({
        String accessToken,
        String refreshToken,
        int expiresAt,
        int? athleteId,
      })> _refreshToken({
    required String clientId,
    required String clientSecret,
    required String refreshToken,
  }) async {
    final response = await _http.post(
      Uri.parse(kStravaTokenUrl),
      body: {
        'client_id': clientId,
        'client_secret': clientSecret,
        'grant_type': 'refresh_token',
        'refresh_token': refreshToken,
      },
    );
    if (response.statusCode != 200) {
      throw Exception(_tokenError(response));
    }
    return _parseTokenResponse(response.body);
  }

  ({
    String accessToken,
    String refreshToken,
    int expiresAt,
    int? athleteId,
  }) _parseTokenResponse(String body) {
    final json = jsonDecode(body) as Map<String, dynamic>;
    final accessToken = json['access_token'] as String? ?? '';
    final refreshToken = json['refresh_token'] as String? ?? '';
    final expiresAt = (json['expires_at'] as num?)?.toInt() ?? 0;
    if (accessToken.isEmpty || refreshToken.isEmpty) {
      throw Exception('invalid token response');
    }
    int? athleteId;
    final athlete = json['athlete'];
    if (athlete is Map<String, dynamic>) {
      athleteId = (athlete['id'] as num?)?.toInt();
    }
    return (
      accessToken: accessToken,
      refreshToken: refreshToken,
      expiresAt: expiresAt,
      athleteId: athleteId,
    );
  }

  String _tokenError(http.Response response) {
    try {
      final json = jsonDecode(response.body);
      if (json is Map<String, dynamic>) {
        final message = json['message'];
        if (message is String && message.isNotEmpty) {
          return message;
        }
      }
    } catch (_) {
      // Fall through.
    }
    return 'token request failed (${response.statusCode})';
  }
}
