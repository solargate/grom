import 'dart:convert';

import 'package:http/http.dart' as http;

import 'models/workout.dart';

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
  Future<String> getServerInfo() async {
    final response = await http.get(Uri.parse('/api/v1/server_info'));
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
      Uri.parse('/api/v1/auth/register'),
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
      Uri.parse('/api/v1/auth/login'),
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
      Uri.parse('/api/v1/auth/me'),
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
      Uri.parse('/api/v1/workouts'),
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

  Future<List<Workout>> listWorkouts(String token) async {
    final response = await http.get(
      Uri.parse('/api/v1/workouts'),
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
