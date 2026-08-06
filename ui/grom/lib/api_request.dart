import 'dart:convert';

import 'package:http/http.dart' as http;

import 'models/downloaded_track.dart';
import 'models/equipment.dart';
import 'models/parsed_track_metadata.dart';
import 'models/social.dart';
import 'models/workout.dart';
import 'models/workout_heartrate.dart';
import 'models/workout_speed.dart';
import 'server_storage.dart';

class ApiException implements Exception {
  ApiException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() => message;
}

class ServerInfo {
  ServerInfo({
    required this.name,
    this.federationEnabled = false,
  });

  final String name;
  final bool federationEnabled;

  factory ServerInfo.fromJson(Map<String, dynamic> json) {
    return ServerInfo(
      name: json['name'] as String? ?? 'Grom Home',
      federationEnabled: json['federation_enabled'] as bool? ?? false,
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
    this.lastSportType,
    this.lastEquipmentBySport = const {},
  });

  final String? lastSportType;
  final Map<String, List<String>> lastEquipmentBySport;

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
    final sport = json['last_sport_type'] as String?;
    return UserProfile(
      lastSportType: (sport == null || sport.isEmpty) ? null : sport,
      lastEquipmentBySport: lastEquipment,
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

class ApiRequest {
  ApiRequest({http.Client? client}) : _client = client ?? http.Client();

  final http.Client _client;

  Uri resolveUri(String path) {
    final base = ServerStorage.cachedBaseUrl;
    if (base == null || base.isEmpty) {
      return Uri.parse(path);
    }
    return Uri.parse(base)
        .resolve(path.startsWith('/') ? path.substring(1) : path);
  }

  Uri _uri(String path) => resolveUri(path);

  Future<ServerInfo> getServerInfo() async {
    final response = await _client.get(_uri('/api/v1/server-info'));
    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return ServerInfo.fromJson(json);
    }
    return ServerInfo(name: 'Grom Home');
  }

  Future<UserInfo> register({
    required String nickname,
    required String name,
    required String email,
    required String password,
  }) async {
    final response = await _client.post(
      _uri('/api/v1/auth/register'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'nickname': nickname,
        'name': name,
        'email': email,
        'password': password,
      }),
    );

    if (response.statusCode == 201) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return UserInfo.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<LoginResult> login({
    required String email,
    required String password,
  }) async {
    final response = await _client.post(
      _uri('/api/v1/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'email': email,
        'password': password,
      }),
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return LoginResult(
        token: json['token'] as String,
        expiresAt: json['expires_at'] as String,
        user: UserInfo.fromJson(json['user'] as Map<String, dynamic>),
      );
    }

    throw _parseError(response);
  }

  Future<UserInfo> getMe(String token) async {
    final response = await _client.get(
      _uri('/api/v1/auth/me'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return UserInfo.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<UserProfile> getProfile(String token) async {
    final response = await _client.get(
      _uri('/api/v1/profile'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return UserProfile.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<UserInfo> updateMe({
    required String token,
    required String name,
  }) async {
    final response = await _client.patch(
      _uri('/api/v1/auth/me'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'name': name}),
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return UserInfo.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<UserInfo> uploadAvatar({
    required String token,
    required List<int> bytes,
  }) async {
    final request = http.MultipartRequest(
      'PUT',
      _uri('/api/v1/auth/me/avatar'),
    );
    request.headers['Authorization'] = 'Bearer $token';
    request.files.add(
      http.MultipartFile.fromBytes(
        'avatar',
        bytes,
        filename: 'avatar.png',
      ),
    );

    final streamed = await _client.send(request);
    final response = await http.Response.fromStream(streamed);

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return UserInfo.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<UserInfo> deleteAvatar({
    required String token,
  }) async {
    final response = await _client.delete(
      _uri('/api/v1/auth/me/avatar'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return UserInfo.fromJson(json);
    }

    throw _parseError(response);
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

  Future<Workout> createWorkout({
    required String token,
    required Map<String, dynamic> body,
  }) async {
    final response = await _client.post(
      _uri('/api/v1/workouts'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body),
    );

    if (response.statusCode == 201) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Workout.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<Workout> updateWorkout({
    required String token,
    required String workoutId,
    required Map<String, dynamic> body,
  }) async {
    final response = await _client.put(
      _uri('/api/v1/workouts/$workoutId'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body),
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Workout.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<Workout> addWorkoutMedia({
    required String token,
    required String workoutId,
    required List<({String filename, List<int> bytes})> photos,
  }) async {
    final request = http.MultipartRequest(
      'POST',
      _uri('/api/v1/workouts/$workoutId/media'),
    );
    request.headers['Authorization'] = 'Bearer $token';
    for (final photo in photos) {
      request.files.add(
        http.MultipartFile.fromBytes(
          'photos',
          photo.bytes,
          filename: photo.filename,
        ),
      );
    }

    final streamed = await _client.send(request);
    final response = await http.Response.fromStream(streamed);

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Workout.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<Workout> deleteWorkoutMedia({
    required String token,
    required String workoutId,
    required String filename,
  }) async {
    final encoded = Uri.encodeComponent(filename);
    final response = await _client.delete(
      _uri('/api/v1/workouts/$workoutId/media/$encoded'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Workout.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<Workout> getWorkout({
    required String token,
    required String workoutId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _client.get(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Workout.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<WorkoutLikesResponse> getWorkoutLikes({
    required String token,
    required String workoutId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/likes');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _client.get(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );
    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return WorkoutLikesResponse.fromJson(json);
    }
    throw _parseError(response);
  }

  Future<WorkoutLikeState> likeWorkout({
    required String token,
    required String workoutId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/likes');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _client.post(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );
    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return WorkoutLikeState.fromJson(json);
    }
    throw _parseError(response);
  }

  Future<WorkoutLikeState> unlikeWorkout({
    required String token,
    required String workoutId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/likes');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _client.delete(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );
    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return WorkoutLikeState.fromJson(json);
    }
    throw _parseError(response);
  }

  Future<Workout> createWorkoutMultipart({
    required String token,
    required Map<String, String> fields,
    List<int>? trackBytes,
    String? trackFilename,
    List<({String filename, List<int> bytes})>? photos,
  }) async {
    final request = http.MultipartRequest(
      'POST',
      _uri('/api/v1/workouts'),
    );
    request.headers['Authorization'] = 'Bearer $token';
    for (final entry in fields.entries) {
      request.fields[entry.key] = entry.value;
    }
    if (trackBytes != null && trackFilename != null) {
      request.files.add(
        http.MultipartFile.fromBytes(
          'track',
          trackBytes,
          filename: trackFilename,
        ),
      );
    }
    if (photos != null) {
      for (final photo in photos) {
        request.files.add(
          http.MultipartFile.fromBytes(
            'photos',
            photo.bytes,
            filename: photo.filename,
          ),
        );
      }
    }

    final streamed = await _client.send(request);
    final response = await http.Response.fromStream(streamed);

    if (response.statusCode == 201) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Workout.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<bool> hasExternalID({
    required String token,
    required String name,
    required String id,
  }) async {
    final response = await _client.get(
      _uri('/api/v1/workouts/external').replace(
        queryParameters: {'name': name, 'id': id},
      ),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return json['exists'] as bool? ?? false;
    }

    throw _parseError(response);
  }

  Future<ParsedTrackMetadata> parseTrack({
    required String token,
    required List<int> trackBytes,
    required String trackFilename,
  }) async {
    final request = http.MultipartRequest(
      'POST',
      _uri('/api/v1/workouts/parse-track'),
    );
    request.headers['Authorization'] = 'Bearer $token';
    request.files.add(
      http.MultipartFile.fromBytes(
        'track',
        trackBytes,
        filename: trackFilename,
      ),
    );

    final streamed = await _client.send(request);
    final response = await http.Response.fromStream(streamed);

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return ParsedTrackMetadata.fromJson(json);
    }

    throw _parseError(response);
  }

  String mapPreviewUrl(String workoutId, {String? owner}) {
    var uri = _uri('/api/v1/workouts/$workoutId/map-preview');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    return uri.toString();
  }

  String mediaPreviewUrl(
    String workoutId,
    String filename, {
    String? owner,
  }) {
    final encoded = Uri.encodeComponent(filename);
    var uri = _uri('/api/v1/workouts/$workoutId/media/$encoded/preview');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    return uri.toString();
  }

  String mediaOriginalUrl(
    String workoutId,
    String filename, {
    String? owner,
  }) {
    final encoded = Uri.encodeComponent(filename);
    var uri = _uri('/api/v1/workouts/$workoutId/media/$encoded');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    return uri.toString();
  }

  String workoutTrackUrl(String workoutId, {String? owner}) {
    var uri = _uri('/api/v1/workouts/$workoutId/track');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    return uri.toString();
  }

  Future<WorkoutSpeedSeries> getWorkoutSpeed({
    required String token,
    required String workoutId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/speed');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _client.get(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return WorkoutSpeedSeries.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<WorkoutHeartRateSeries> getWorkoutHeartRate({
    required String token,
    required String workoutId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/heartrate');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _client.get(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return WorkoutHeartRateSeries.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<DownloadedTrack> downloadWorkoutTrack({
    required String token,
    required String workoutId,
    required String fallbackFilename,
    String? owner,
    String? format,
  }) async {
    final query = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      query['owner'] = owner;
    }
    if (format == 'gpx') {
      query['format'] = 'gpx';
    }
    var uri = _uri('/api/v1/workouts/$workoutId/track');
    if (query.isNotEmpty) {
      uri = uri.replace(queryParameters: query);
    }
    final response = await _client.get(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final filename = _filenameFromContentDisposition(
            response.headers['content-disposition'],
          ) ??
          fallbackFilename;
      return DownloadedTrack(
        bytes: response.bodyBytes,
        filename: filename,
      );
    }

    throw _parseError(response);
  }

  String? _filenameFromContentDisposition(String? header) {
    if (header == null || header.isEmpty) {
      return null;
    }
    final utf8Match = RegExp(
      r"filename\*=UTF-8''([^;]+)",
      caseSensitive: false,
    ).firstMatch(header);
    if (utf8Match != null) {
      return Uri.decodeComponent(utf8Match.group(1)!);
    }
    final match = RegExp(r'filename="?([^";]+)"?').firstMatch(header);
    return match?.group(1);
  }

  Future<WorkoutListPage> listWorkouts(
    String token, {
    String scope = 'feed',
    int limit = 20,
    String? cursor,
  }) async {
    final params = <String, String>{
      'limit': '$limit',
    };
    if (scope != 'feed') {
      params['scope'] = scope;
    }
    if (cursor != null && cursor.isNotEmpty) {
      params['cursor'] = cursor;
    }
    final uri = _uri('/api/v1/workouts').replace(queryParameters: params);
    final response = await _client.get(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return WorkoutListPage.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<List<UserSearchResult>> searchUsers({
    required String token,
    required String query,
  }) async {
    final response = await _client.get(
      _uri('/api/v1/users/search?q=${Uri.encodeQueryComponent(query)}'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as List<dynamic>;
      return json
          .map(
              (item) => UserSearchResult.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response);
  }

  Future<FollowInfo> followUser({
    required String token,
    required String handle,
  }) async {
    final response = await _client.post(
      _uri('/api/v1/social/follow'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'handle': handle}),
    );

    if (response.statusCode == 201) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return FollowInfo.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<void> unfollowUser({
    required String token,
    required String followId,
  }) async {
    final response = await _client.delete(
      _uri('/api/v1/social/follow/$followId'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response);
  }

  Future<List<FollowInfo>> listFollowing(String token) async {
    final response = await _client.get(
      _uri('/api/v1/social/following'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as List<dynamic>;
      return json
          .map((item) => FollowInfo.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response);
  }

  Future<List<FollowerInfo>> listFollowers(String token) async {
    final response = await _client.get(
      _uri('/api/v1/social/followers'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as List<dynamic>;
      return json
          .map((item) => FollowerInfo.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response);
  }

  Future<List<Equipment>> listEquipment(String token) async {
    final response = await _client.get(
      _uri('/api/v1/equipment'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as List<dynamic>;
      return json
          .map((item) => Equipment.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response);
  }

  Future<Equipment> createEquipment({
    required String token,
    required EquipmentDraft body,
  }) async {
    final response = await _client.post(
      _uri('/api/v1/equipment'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body.toJson()),
    );

    if (response.statusCode == 201) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Equipment.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<Equipment> updateEquipment({
    required String token,
    required String id,
    required EquipmentDraft body,
  }) async {
    final response = await _client.put(
      _uri('/api/v1/equipment/$id'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body.toJson()),
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Equipment.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<void> deleteEquipment({
    required String token,
    required String id,
  }) async {
    final response = await _client.delete(
      _uri('/api/v1/equipment/$id'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response);
  }

  Future<void> deleteWorkout({
    required String token,
    required String workoutId,
  }) async {
    final response = await _client.delete(
      _uri('/api/v1/workouts/$workoutId'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response);
  }

  Future<Map<String, dynamic>> uploadStravaArchiveRaw({
    required String token,
    required List<int> bytes,
    void Function(double progress)? onProgress,
  }) async {
    final request = http.StreamedRequest(
      'POST',
      _uri('/api/v1/integrations/strava/import'),
    );
    request.headers['Authorization'] = 'Bearer $token';
    request.headers['Content-Type'] = 'application/zip';
    request.contentLength = bytes.length;
    request.sink.add(bytes);
    await request.sink.close();
    return _sendStravaStreamedRequest(request, onProgress);
  }

  Future<Map<String, dynamic>> uploadStravaArchiveRawStream({
    required String token,
    required Stream<List<int>> stream,
    required int length,
    void Function(double progress)? onProgress,
  }) async {
    final request = http.StreamedRequest(
      'POST',
      _uri('/api/v1/integrations/strava/import'),
    );
    request.headers['Authorization'] = 'Bearer $token';
    request.headers['Content-Type'] = 'application/zip';
    request.contentLength = length;

    var sent = 0;
    await for (final chunk in stream) {
      request.sink.add(chunk);
      sent += chunk.length;
      if (onProgress != null && length > 0) {
        onProgress(sent / length);
      }
    }
    await request.sink.close();
    return _sendStravaStreamedRequest(request, onProgress);
  }

  Future<Map<String, dynamic>> uploadStravaArchive({
    required String token,
    required List<int> bytes,
    required String filename,
    void Function(double progress)? onProgress,
  }) async {
    return uploadStravaArchiveRaw(
      token: token,
      bytes: bytes,
      onProgress: onProgress,
    );
  }

  Future<Map<String, dynamic>> uploadStravaArchiveFromStream({
    required String token,
    required Stream<List<int>> stream,
    required int length,
    required String filename,
    void Function(double progress)? onProgress,
  }) async {
    return uploadStravaArchiveRawStream(
      token: token,
      stream: stream,
      length: length,
      onProgress: onProgress,
    );
  }

  Future<Map<String, dynamic>> _sendStravaStreamedRequest(
    http.StreamedRequest request,
    void Function(double progress)? onProgress,
  ) async {
    if (onProgress != null) {
      onProgress(0);
    }

    final streamed = await _client.send(request);

    if (onProgress != null) {
      onProgress(1);
    }

    final response = await http.Response.fromStream(streamed);

    if (response.statusCode == 202) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }

    throw _parseError(response);
  }

  Future<Map<String, dynamic>> getStravaImportStatus(String token) async {
    final response = await _client.get(
      _uri('/api/v1/integrations/strava/import/status'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }

    throw _parseError(response);
  }

  ApiException _parseError(http.Response response) {
    try {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      final message = json['error'] as String? ?? 'Unknown error';
      return ApiException(message, statusCode: response.statusCode);
    } catch (_) {
      return ApiException('Request failed (${response.statusCode})',
          statusCode: response.statusCode);
    }
  }
}
