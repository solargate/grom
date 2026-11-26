part of 'api_request.dart';

mixin _AccountApi on _ApiTransport {
  Future<void> forgotPassword({
    required String email,
    String? altcha,
  }) async {
    final body = <String, dynamic>{'email': email};
    if (altcha != null && altcha.isNotEmpty) {
      body['altcha'] = altcha;
    }
    final response = await _request(
      'POST',
      _uri('/api/v1/auth/password/forgot'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(body),
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response);
  }

  Future<void> resetPassword({
    required String token,
    required String password,
  }) async {
    final response = await _request(
      'POST',
      _uri('/api/v1/auth/password/reset'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'token': token,
        'password': password,
      }),
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response);
  }

  Future<UserInfo> register({
    required String nickname,
    required String name,
    required String email,
    required String password,
    String? altcha,
  }) async {
    final body = <String, dynamic>{
      'nickname': nickname,
      'name': name,
      'email': email,
      'password': password,
    };
    if (altcha != null && altcha.isNotEmpty) {
      body['altcha'] = altcha;
    }
    final response = await _request(
      'POST',
      _uri('/api/v1/auth/register'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(body),
    );

    if (response.statusCode == 201) {
      final json = _decode<Map<String, dynamic>>(response);
      return UserInfo.fromJson(json);
    }

    throw _parseError(response);
  }

  Future<LoginResult> login({
    required String email,
    required String password,
    String? altcha,
  }) async {
    final body = <String, dynamic>{
      'email': email,
      'password': password,
    };
    if (altcha != null && altcha.isNotEmpty) {
      body['altcha'] = altcha;
    }
    final response = await _request(
      'POST',
      _uri('/api/v1/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(body),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return LoginResult(
        token: json['token'] as String,
        expiresAt: json['expires_at'] as String,
        user: UserInfo.fromJson(json['user'] as Map<String, dynamic>),
      );
    }

    throw _parseError(response);
  }

  Future<UserInfo> getMe(
    String token, {
    bool notifyOnUnauthorized = true,
  }) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/auth/me'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return UserInfo.fromJson(json);
    }

    throw _parseError(
      response,
      authenticated: notifyOnUnauthorized,
    );
  }

  Future<UserProfile> getProfile(String token) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/profile'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return UserProfile.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<UserInfo> updateMe({
    required String token,
    required String name,
  }) async {
    final response = await _request(
      'PATCH',
      _uri('/api/v1/auth/me'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'name': name}),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return UserInfo.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<UserInfo> uploadAvatar({
    required String token,
    required List<int> bytes,
  }) async {
    final (:request, :abort) = _multipart(
      'PUT',
      _uri('/api/v1/auth/me/avatar'),
    );
    request.headers.addAll(_bearerHeaders(token));
    request.files.add(
      http.MultipartFile.fromBytes(
        'avatar',
        bytes,
        filename: 'avatar.png',
      ),
    );

    final response =
        await _send(request, timeout: _uploadTimeout, onTimeout: abort);

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return UserInfo.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<UserInfo> deleteAvatar({
    required String token,
  }) async {
    final response = await _request(
      'DELETE',
      _uri('/api/v1/auth/me/avatar'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return UserInfo.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<void> deleteAccount({
    required String token,
    required String password,
  }) async {
    final response = await _request(
      'DELETE',
      _uri('/api/v1/auth/me'),
      headers: _bearerHeaders(token, contentType: 'application/json'),
      body: jsonEncode({'password': password}),
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response, authenticated: true);
  }

  Future<List<PersonalAccessToken>> listPersonalAccessTokens(
      String token) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/auth/pat'),
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map((item) =>
              PersonalAccessToken.fromJson(item as Map<String, dynamic>))
          .toList();
    }
    throw _parseError(response, authenticated: true);
  }

  Future<CreatePersonalAccessTokenResult> createPersonalAccessToken({
    required String token,
    required String name,
    required List<String> scopes,
    int? expiresInDays,
    bool noExpiration = false,
  }) async {
    final body = <String, dynamic>{
      'name': name,
      'scopes': scopes,
      'no_expiration': noExpiration,
    };
    if (!noExpiration && expiresInDays != null) {
      body['expires_in_days'] = expiresInDays;
    }
    final response = await _request(
      'POST',
      _uri('/api/v1/auth/pat'),
      headers: _bearerHeaders(token, contentType: 'application/json'),
      body: jsonEncode(body),
    );
    if (response.statusCode == 201) {
      return CreatePersonalAccessTokenResult.fromJson(
        _decode<Map<String, dynamic>>(response),
      );
    }
    throw _parseError(response, authenticated: true);
  }

  Future<void> revokePersonalAccessToken({
    required String token,
    required String id,
  }) async {
    final response = await _request(
      'DELETE',
      _uri('/api/v1/auth/pat/$id'),
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 204) {
      return;
    }
    throw _parseError(response, authenticated: true);
  }
}
