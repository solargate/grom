import 'package:shared_preferences/shared_preferences.dart';

const stravaApiEnabledStorageKey = 'strava_api_enabled';
const stravaApiClientIdStorageKey = 'strava_api_client_id';
const stravaApiClientSecretStorageKey = 'strava_api_client_secret';
const stravaApiAccessTokenStorageKey = 'strava_api_access_token';
const stravaApiRefreshTokenStorageKey = 'strava_api_refresh_token';
const stravaApiExpiresAtStorageKey = 'strava_api_expires_at';
const stravaApiConnectedStorageKey = 'strava_api_connected';
const stravaApiAthleteIdStorageKey = 'strava_api_athlete_id';
const stravaApiScopeStorageKey = 'strava_api_scope';

/// Local prefs for Strava API BYO credentials and OAuth tokens.
///
/// Intentionally not cleared on Grom logout so re-login keeps Connect state.
class StravaApiStorage {
  static Future<bool> loadEnabled({bool defaultValue = false}) async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getBool(stravaApiEnabledStorageKey) ?? defaultValue;
  }

  static Future<void> saveEnabled(bool enabled) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(stravaApiEnabledStorageKey, enabled);
  }

  static Future<String> loadClientId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(stravaApiClientIdStorageKey) ?? '';
  }

  static Future<String> loadClientSecret() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(stravaApiClientSecretStorageKey) ?? '';
  }

  static Future<void> saveCredentials({
    required String clientId,
    required String clientSecret,
  }) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(stravaApiClientIdStorageKey, clientId);
    await prefs.setString(stravaApiClientSecretStorageKey, clientSecret);
  }

  static Future<({
    String accessToken,
    String refreshToken,
    int expiresAt,
    bool connected,
    int? athleteId,
    String scope,
  })> loadTokens() async {
    final prefs = await SharedPreferences.getInstance();
    return (
      accessToken: prefs.getString(stravaApiAccessTokenStorageKey) ?? '',
      refreshToken: prefs.getString(stravaApiRefreshTokenStorageKey) ?? '',
      expiresAt: prefs.getInt(stravaApiExpiresAtStorageKey) ?? 0,
      connected: prefs.getBool(stravaApiConnectedStorageKey) ?? false,
      athleteId: prefs.getInt(stravaApiAthleteIdStorageKey),
      scope: prefs.getString(stravaApiScopeStorageKey) ?? '',
    );
  }

  static Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
    required int expiresAt,
    required bool connected,
    int? athleteId,
    String scope = '',
  }) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(stravaApiAccessTokenStorageKey, accessToken);
    await prefs.setString(stravaApiRefreshTokenStorageKey, refreshToken);
    await prefs.setInt(stravaApiExpiresAtStorageKey, expiresAt);
    await prefs.setBool(stravaApiConnectedStorageKey, connected);
    if (athleteId != null) {
      await prefs.setInt(stravaApiAthleteIdStorageKey, athleteId);
    }
    if (scope.isNotEmpty) {
      await prefs.setString(stravaApiScopeStorageKey, scope);
    }
  }

  static Future<void> clearTokens() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(stravaApiAccessTokenStorageKey);
    await prefs.remove(stravaApiRefreshTokenStorageKey);
    await prefs.remove(stravaApiExpiresAtStorageKey);
    await prefs.setBool(stravaApiConnectedStorageKey, false);
    await prefs.remove(stravaApiAthleteIdStorageKey);
    await prefs.remove(stravaApiScopeStorageKey);
  }
}
