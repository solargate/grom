import 'dart:convert';

import 'package:http/http.dart' as http;

import 'models/downloaded_track.dart';
import 'models/parsed_track_metadata.dart';
import 'models/social.dart';
import 'models/workout.dart';
import 'server_storage.dart';

class ApiException implements Exception {
  ApiException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() => message;
}

class UserInfo {
  UserInfo({
    required this.id,
    required this.nickname,
    required this.name,
    required this.email,
  });

  final String id;
  final String nickname;
  final String name;
  final String email;

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['id'] as String,
      nickname: json['nickname'] as String,
      name: json['name'] as String? ?? '',
      email: json['email'] as String,
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
  Uri _uri(String path) {
    final base = ServerStorage.cachedBaseUrl;
    if (base == null || base.isEmpty) {
      return Uri.parse(path);
    }
    return Uri.parse(base).resolve(path.startsWith('/') ? path.substring(1) : path);
  }

  Future<String> getServerInfo() async {
    final response = await http.get(_uri('/api/v1/server_info'));
    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return json['name'] as String;
    }
    return 'Travka Home';
  }

  Future<UserInfo> register({
    required String nickname,
    required String name,
    required String email,
    required String password,
  }) async {
    final response = await http.post(
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
    final response = await http.post(
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
    final response = await http.get(
      _uri('/api/v1/auth/me'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return UserInfo.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<Workout> createWorkout({
    required String token,
    required Map<String, dynamic> body,
  }) async {
    final response = await http.post(
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

  Future<Workout> createWorkoutMultipart({
    required String token,
    required Map<String, String> fields,
    required List<int> trackBytes,
    required String trackFilename,
  }) async {
    final request = http.MultipartRequest(
      'POST',
      _uri('/api/v1/workouts'),
    );
    request.headers['Authorization'] = 'Bearer $token';
    for (final entry in fields.entries) {
      request.fields[entry.key] = entry.value;
    }
    request.files.add(
      http.MultipartFile.fromBytes(
        'track',
        trackBytes,
        filename: trackFilename,
      ),
    );

    final streamed = await request.send();
    final response = await http.Response.fromStream(streamed);

    if (response.statusCode == 201) {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return Workout.fromJson(json);
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

    final streamed = await request.send();
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

  String workoutTrackUrl(String workoutId, {String? owner}) {
    var uri = _uri('/api/v1/workouts/$workoutId/track');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    return uri.toString();
  }

  Future<DownloadedTrack> downloadWorkoutTrack({
    required String token,
    required String workoutId,
    required String fallbackFilename,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/track');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await http.get(
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
    final match = RegExp(r'filename="?([^";]+)"?').firstMatch(header);
    return match?.group(1);
  }

  Future<List<Workout>> listWorkouts(String token) async {
    final response = await http.get(
      _uri('/api/v1/workouts'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as List<dynamic>;
      return json
          .map((item) => Workout.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response);
  }

  Future<List<UserSearchResult>> searchUsers({
    required String token,
    required String query,
  }) async {
    final response = await http.get(
      _uri('/api/v1/users/search?q=${Uri.encodeQueryComponent(query)}'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body) as List<dynamic>;
      return json
          .map((item) => UserSearchResult.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response);
  }

  Future<FollowInfo> followUser({
    required String token,
    required String handle,
  }) async {
    final response = await http.post(
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
    final response = await http.delete(
      _uri('/api/v1/social/follow/$followId'),
      headers: {'Authorization': 'Bearer $token'},
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response);
  }

  Future<List<FollowInfo>> listFollowing(String token) async {
    final response = await http.get(
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
