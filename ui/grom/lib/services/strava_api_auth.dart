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
  StravaApiAuth({
    http.Client? httpClient,
    Future<String> Function(String authorizeUrl)? authenticate,
  })  : _http = httpClient ?? http.Client(),
        _authenticate = authenticate ?? authenticateStravaOAuth;

  final http.Client _http;
  final Future<String> Function(String authorizeUrl) _authenticate;

  static Map<String, String> get _stravaHeaders => {
        'User-Agent': kStravaHttpUserAgent,
        'Accept': 'application/json',
      };

  Uri buildAuthorizeUrl({required String clientId}) {
    return Uri.parse(kStravaAuthorizeUrl).replace(
      queryParameters: {
        'client_id': clientId,
        'redirect_uri': kStravaOAuthRedirectUri,
        'response_type': 'code',
        'approval_prompt': 'force',
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
      final redirectUrl = await _authenticate(
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
    // JSON body; do not send redirect_uri on token exchange (Strava docs omit it).
    return _parseTokenResponse(
      await _tokenRequest({
        'client_id': clientId,
        'client_secret': clientSecret,
        'code': code,
        'grant_type': 'authorization_code',
      }),
    );
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
    return _parseTokenResponse(
      await _tokenRequest({
        'client_id': clientId,
        'client_secret': clientSecret,
        'grant_type': 'refresh_token',
        'refresh_token': refreshToken,
      }),
    );
  }

  Future<String> _tokenRequest(Map<String, String> params) async {
    final response = await _http.post(
      Uri.parse(kStravaTokenUrl),
      headers: {
        ..._stravaHeaders,
        'Content-Type': 'application/json',
      },
      body: jsonEncode(params),
    );
    if (response.statusCode != 200) {
      throw Exception(_tokenError(response));
    }
    return response.body;
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
    final body = response.body.trim();
    try {
      final json = jsonDecode(body);
      if (json is Map<String, dynamic>) {
        final message = json['message'];
        final errors = json['errors'];
        if (errors is List && errors.isNotEmpty) {
          final details = errors
              .whereType<Map>()
              .map((e) {
                final field = e['field'];
                final code = e['code'];
                final resource = e['resource'];
                return [
                  if (resource != null) '$resource',
                  if (field != null && '$field'.isNotEmpty) 'field=$field',
                  if (code != null) 'code=$code',
                ].join(' ');
              })
              .where((s) => s.isNotEmpty)
              .join('; ');
          if (message is String && message.isNotEmpty && details.isNotEmpty) {
            return '$message ($details)';
          }
          if (details.isNotEmpty) {
            return details;
          }
        }
        if (message is String && message.isNotEmpty) {
          return message;
        }
      }
    } catch (_) {
      // Non-JSON (often Cloudflare HTML on 403).
    }
    if (response.statusCode == 403) {
      return 'token request forbidden (403). '
          'Check Client ID/Secret, disable VPN if enabled, then retry Connect';
    }
    if (body.isNotEmpty && body.length < 200 && !body.startsWith('<')) {
      return 'token request failed (${response.statusCode}): $body';
    }
    return 'token request failed (${response.statusCode})';
  }
}
