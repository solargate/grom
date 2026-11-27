import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'models/personal_access_token.dart';
import 'models/downloaded_track.dart';
import 'models/equipment.dart';
import 'models/parsed_track_metadata.dart';
import 'models/social.dart';
import 'models/user_public_profile.dart';
import 'models/workout.dart';
import 'models/workout_heartrate.dart';
import 'models/workout_cadence.dart';
import 'models/workout_speed.dart';
import 'server_storage.dart';
import 'session.dart';

part 'api_request_account.dart';
part 'api_request_workouts.dart';
part 'api_request_social.dart';
part 'api_request_integrations.dart';

class ApiException implements Exception {
  ApiException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() => message;
}

/// Thrown when an authenticated API call receives HTTP 401.
///
/// [SessionCoordinator] has already cleared the JWT and (when enabled) emitted
/// [SessionEndReason.expired]. UI catch sites should stop loading and skip
/// showing a raw error snackbar.
class SessionExpiredException implements Exception {
  @override
  String toString() => 'session expired';
}

class ServerInfo {
  ServerInfo({
    required this.name,
    this.federationEnabled = false,
    this.passwordResetEnabled = false,
    this.captchaEnabled = false,
    this.vapidPublicKey = '',
  });

  final String name;
  final bool federationEnabled;
  final bool passwordResetEnabled;
  final bool captchaEnabled;
  final String vapidPublicKey;

  factory ServerInfo.fromJson(Map<String, dynamic> json) {
    return ServerInfo(
      name: json['name'] as String? ?? 'Grom Home',
      federationEnabled: json['federation_enabled'] as bool? ?? false,
      passwordResetEnabled: json['password_reset_enabled'] as bool? ?? false,
      captchaEnabled: json['captcha_enabled'] as bool? ?? false,
      vapidPublicKey: json['vapid_public_key'] as String? ?? '',
    );
  }
}

class UserInfo {
  UserInfo({
    required this.id,
    required this.nickname,
    required this.name,
    required this.email,
    this.hasAvatar = false,
    this.avatarUrl,
  });

  final String id;
  final String nickname;
  final String name;
  final String email;
  final bool hasAvatar;
  final String? avatarUrl;

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['id'] as String,
      nickname: json['nickname'] as String,
      name: json['name'] as String? ?? '',
      email: json['email'] as String,
      hasAvatar: json['has_avatar'] as bool? ?? false,
      avatarUrl: json['avatar_url'] as String?,
    );
  }
}

class UserProfile {
  UserProfile({
    this.totalWorkouts,
    this.lastSportType,
    this.lastEquipmentBySport = const {},
    this.usedSportTypes = const [],
  });

  final int? totalWorkouts;
  final String? lastSportType;
  final Map<String, List<String>> lastEquipmentBySport;
  final List<String> usedSportTypes;

  factory UserProfile.fromJson(Map<String, dynamic> json) {
    final lastEquipmentRaw = json['last_equipment_by_sport'];
    final lastEquipment = <String, List<String>>{};
    if (lastEquipmentRaw is Map<String, dynamic>) {
      for (final entry in lastEquipmentRaw.entries) {
        final value = entry.value;
        if (value is List) {
          lastEquipment[entry.key] =
              value.map((item) => item.toString()).toList();
        }
      }
    }
    final usedRaw = json['used_sport_types'];
    final usedSports = <String>[];
    if (usedRaw is List) {
      for (final item in usedRaw) {
        final s = item.toString();
        if (s.isNotEmpty) {
          usedSports.add(s);
        }
      }
    }
    final sport = json['last_sport_type'] as String?;
    return UserProfile(
      totalWorkouts: json['total_workouts'] as int?,
      lastSportType: (sport == null || sport.isEmpty) ? null : sport,
      lastEquipmentBySport: lastEquipment,
      usedSportTypes: usedSports,
    );
  }
}

class LoginResult {
  LoginResult({
    required this.token,
    required this.expiresAt,
    required this.user,
  });

  final String token;
  final String expiresAt;
  final UserInfo user;
}

abstract class _ApiTransport {
  _ApiTransport({http.Client? client}) : _client = client ?? http.Client();

  final http.Client _client;

  final _requestTimeout = const Duration(seconds: 30);
  final _uploadTimeout = const Duration(minutes: 2);
  final _archiveTimeout = const Duration(minutes: 30);

  Map<String, String> _bearerHeaders(String token, {String? contentType}) => {
        'Authorization': 'Bearer $token',
        if (contentType != null) 'Content-Type': contentType,
      };

  T _decode<T>(http.Response response) => jsonDecode(response.body) as T;

  ({http.MultipartRequest request, void Function() abort}) _multipart(
    String method,
    Uri uri,
  ) {
    final abort = Completer<void>();
    return (
      request: http.AbortableMultipartRequest(method, uri,
          abortTrigger: abort.future),
      abort: abort.complete,
    );
  }

  Future<http.Response> _request(
    String method,
    Uri uri, {
    Map<String, String>? headers,
    String? body,
  }) {
    final abort = Completer<void>();
    final request =
        http.AbortableRequest(method, uri, abortTrigger: abort.future);
    if (headers != null) request.headers.addAll(headers);
    if (body != null) request.body = body;
    return _send(request, timeout: _requestTimeout, onTimeout: abort.complete);
  }

  Future<http.Response> _send(
    http.BaseRequest request, {
    required Duration timeout,
    void Function()? onHeaders,
    void Function()? onTimeout,
  }) {
    Future<http.Response> readResponse() async {
      final streamed = await _client.send(request);
      onHeaders?.call();
      return http.Response.fromStream(streamed);
    }

    return readResponse().timeout(timeout, onTimeout: () {
      onTimeout?.call();
      throw TimeoutException('Grom request timed out', timeout);
    });
  }

  Uri resolveUri(String path) {
    final base = ServerStorage.cachedBaseUrl;
    if (base == null || base.isEmpty) {
      return Uri.parse(path);
    }
    return Uri.parse(base)
        .resolve(path.startsWith('/') ? path.substring(1) : path);
  }

  Uri _uri(String path) => resolveUri(path);

  Never _parseError(
    http.Response response, {
    bool authenticated = false,
  }) {
    if (authenticated && response.statusCode == 401) {
      // Fire-and-forget clear; callers still receive SessionExpiredException.
      unawaited(SessionCoordinator.instance.notifyUnauthorized());
      throw SessionExpiredException();
    }
    try {
      final json = _decode<Map<String, dynamic>>(response);
      final message = json['error'] as String? ?? 'Unknown error';
      throw ApiException(message, statusCode: response.statusCode);
    } on ApiException {
      rethrow;
    } on SessionExpiredException {
      rethrow;
    } catch (_) {
      throw ApiException(
        'Request failed (${response.statusCode})',
        statusCode: response.statusCode,
      );
    }
  }
}

class ApiRequest extends _ApiTransport
    with _AccountApi, _WorkoutsApi, _SocialApi, _IntegrationsApi {
  ApiRequest({super.client});

  Future<ServerInfo> getServerInfo() async {
    final response = await _request('GET', _uri('/api/v1/server-info'));
    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return ServerInfo.fromJson(json);
    }
    return ServerInfo(name: 'Grom Home');
  }

  Future<void> registerPushSubscription({
    required String token,
    required String installationId,
    required String endpoint,
    required String p256dh,
    required String auth,
    String platform = 'android',
    String? deviceName,
  }) async {
    final body = <String, dynamic>{
      'installation_id': installationId,
      'endpoint': endpoint,
      'keys': {
        'p256dh': p256dh,
        'auth': auth,
      },
      'platform': platform,
    };
    if (deviceName != null && deviceName.isNotEmpty) {
      body['device_name'] = deviceName;
    }
    final response = await _request(
      'POST',
      _uri('/api/v1/notifications/push'),
      headers: _bearerHeaders(token, contentType: 'application/json'),
      body: jsonEncode(body),
    );
    if (response.statusCode == 200) {
      return;
    }
    throw _parseError(response, authenticated: true);
  }

  Future<void> deletePushSubscription({
    required String token,
    required String installationId,
  }) async {
    final response = await _request(
      'DELETE',
      _uri('/api/v1/notifications/push/$installationId'),
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 204 || response.statusCode == 404) {
      return;
    }
    throw _parseError(response, authenticated: true);
  }

  static String resolveAvatarUrl({
    required String nickname,
    bool hasAvatar = false,
    String? avatarUrl,
  }) {
    if (avatarUrl != null && avatarUrl.isNotEmpty) {
      if (avatarUrl.startsWith('http://') || avatarUrl.startsWith('https://')) {
        return avatarUrl;
      }
      final base = ServerStorage.cachedBaseUrl;
      if (base == null || base.isEmpty) {
        return avatarUrl;
      }
      return Uri.parse(base)
          .resolve(
            avatarUrl.startsWith('/') ? avatarUrl.substring(1) : avatarUrl,
          )
          .toString();
    }
    if (!hasAvatar) {
      return '';
    }
    final api = ApiRequest();
    return api._uri('/api/v1/users/$nickname/avatar').toString();
  }

  Future<List<Equipment>> listEquipment(String token) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/equipment'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map((item) => Equipment.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response, authenticated: true);
  }

  Future<Equipment> createEquipment({
    required String token,
    required EquipmentDraft body,
  }) async {
    final response = await _request(
      'POST',
      _uri('/api/v1/equipment'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body.toJson()),
    );

    if (response.statusCode == 201) {
      final json = _decode<Map<String, dynamic>>(response);
      return Equipment.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<Equipment> updateEquipment({
    required String token,
    required String id,
    required EquipmentDraft body,
  }) async {
    final response = await _request(
      'PUT',
      _uri('/api/v1/equipment/$id'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body.toJson()),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return Equipment.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<void> deleteEquipment({
    required String token,
    required String id,
  }) async {
    final response = await _request(
      'DELETE',
      _uri('/api/v1/equipment/$id'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response, authenticated: true);
  }
}
