part of 'api_request.dart';

mixin _SocialApi on _ApiTransport {
  Future<List<UserSearchResult>> listLocalUsers(String token) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/users'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map(
              (item) => UserSearchResult.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response, authenticated: true);
  }

  Future<List<UserSearchResult>> searchUsers({
    required String token,
    required String query,
  }) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/users/search?q=${Uri.encodeQueryComponent(query)}'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map(
              (item) => UserSearchResult.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response, authenticated: true);
  }

  Future<FollowInfo> followUser({
    required String token,
    required String handle,
  }) async {
    final response = await _request(
      'POST',
      _uri('/api/v1/social/follow'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'handle': handle}),
    );

    if (response.statusCode == 201) {
      final json = _decode<Map<String, dynamic>>(response);
      return FollowInfo.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<void> unfollowUser({
    required String token,
    required String followId,
  }) async {
    final response = await _request(
      'DELETE',
      _uri('/api/v1/social/follow/$followId'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response, authenticated: true);
  }

  Future<List<FollowInfo>> listFollowing(String token) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/social/following'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map((item) => FollowInfo.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response, authenticated: true);
  }

  Future<List<FollowerInfo>> listFollowers(String token) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/social/followers'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map((item) => FollowerInfo.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response, authenticated: true);
  }

  Uri _userHandleUri(String handle, [String subPath = '']) {
    final encoded = Uri.encodeComponent(handle);
    final path = subPath.isEmpty
        ? '/api/v1/users/$encoded'
        : '/api/v1/users/$encoded/$subPath';
    return _uri(path);
  }

  Future<UserPublicProfile> getUserProfile({
    required String token,
    required String handle,
  }) async {
    final response = await _request(
      'GET',
      _userHandleUri(handle),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return UserPublicProfile.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<List<FollowInfo>> listUserFollowing({
    required String token,
    required String handle,
  }) async {
    final response = await _request(
      'GET',
      _userHandleUri(handle, 'following'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map((item) => FollowInfo.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response, authenticated: true);
  }

  Future<List<FollowerInfo>> listUserFollowers({
    required String token,
    required String handle,
  }) async {
    final response = await _request(
      'GET',
      _userHandleUri(handle, 'followers'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<List<dynamic>>(response);
      return json
          .map((item) => FollowerInfo.fromJson(item as Map<String, dynamic>))
          .toList();
    }

    throw _parseError(response, authenticated: true);
  }

  Future<WorkoutListPage> listUserWorkouts(
    String token, {
    required String handle,
    int limit = 20,
    String? cursor,
  }) async {
    final params = <String, String>{'limit': '$limit'};
    if (cursor != null && cursor.isNotEmpty) {
      params['cursor'] = cursor;
    }
    final uri =
        _userHandleUri(handle, 'workouts').replace(queryParameters: params);
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutListPage.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }
}
