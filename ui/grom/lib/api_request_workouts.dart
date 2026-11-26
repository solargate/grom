part of 'api_request.dart';

mixin _WorkoutsApi on _ApiTransport {
  Future<Workout> createWorkout({
    required String token,
    required Map<String, dynamic> body,
  }) async {
    final response = await _request(
      'POST',
      _uri('/api/v1/workouts'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body),
    );

    if (response.statusCode == 201) {
      final json = _decode<Map<String, dynamic>>(response);
      return Workout.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<Workout> updateWorkout({
    required String token,
    required String workoutId,
    required Map<String, dynamic> body,
  }) async {
    final response = await _request(
      'PUT',
      _uri('/api/v1/workouts/$workoutId'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return Workout.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<Workout> addWorkoutMedia({
    required String token,
    required String workoutId,
    required List<({String filename, List<int> bytes})> photos,
  }) async {
    final (:request, :abort) = _multipart(
      'POST',
      _uri('/api/v1/workouts/$workoutId/media'),
    );
    request.headers.addAll(_bearerHeaders(token));
    for (final photo in photos) {
      request.files.add(
        http.MultipartFile.fromBytes(
          'photos',
          photo.bytes,
          filename: photo.filename,
        ),
      );
    }

    final response =
        await _send(request, timeout: _uploadTimeout, onTimeout: abort);

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return Workout.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<Workout> deleteWorkoutMedia({
    required String token,
    required String workoutId,
    required String filename,
  }) async {
    final encoded = Uri.encodeComponent(filename);
    final response = await _request(
      'DELETE',
      _uri('/api/v1/workouts/$workoutId/media/$encoded'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return Workout.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
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
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return Workout.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
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
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutLikesResponse.fromJson(json);
    }
    throw _parseError(response, authenticated: true);
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
    final response = await _request(
      'POST',
      uri,
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutLikeState.fromJson(json);
    }
    throw _parseError(response, authenticated: true);
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
    final response = await _request(
      'DELETE',
      uri,
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutLikeState.fromJson(json);
    }
    throw _parseError(response, authenticated: true);
  }

  Future<WorkoutCommentsResponse> getWorkoutComments({
    required String token,
    required String workoutId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/comments');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutCommentsResponse.fromJson(json);
    }
    throw _parseError(response, authenticated: true);
  }

  Future<WorkoutCommentCreateResponse> createWorkoutComment({
    required String token,
    required String workoutId,
    required String text,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/comments');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _request(
      'POST',
      uri,
      headers: _bearerHeaders(token, contentType: 'application/json'),
      body: jsonEncode({'text': text}),
    );
    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutCommentCreateResponse.fromJson(json);
    }
    throw _parseError(response, authenticated: true);
  }

  Future<int> deleteWorkoutComment({
    required String token,
    required String workoutId,
    required String commentId,
    String? owner,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/comments/$commentId');
    if (owner != null && owner.isNotEmpty) {
      uri = uri.replace(queryParameters: {'owner': owner});
    }
    final response = await _request(
      'DELETE',
      uri,
      headers: _bearerHeaders(token),
    );
    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return json['count'] as int? ?? 0;
    }
    throw _parseError(response, authenticated: true);
  }

  Future<Workout> createWorkoutMultipart({
    required String token,
    required Map<String, String> fields,
    List<int>? trackBytes,
    String? trackFilename,
    List<({String filename, List<int> bytes})>? photos,
  }) async {
    final (:request, :abort) = _multipart(
      'POST',
      _uri('/api/v1/workouts'),
    );
    request.headers.addAll(_bearerHeaders(token));
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

    final response =
        await _send(request, timeout: _uploadTimeout, onTimeout: abort);

    if (response.statusCode == 201) {
      final json = _decode<Map<String, dynamic>>(response);
      return Workout.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<bool> hasExternalID({
    required String token,
    required String name,
    required String id,
  }) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/workouts/external').replace(
        queryParameters: {'name': name, 'id': id},
      ),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return json['exists'] as bool? ?? false;
    }

    throw _parseError(response, authenticated: true);
  }

  Future<ParsedTrackMetadata> parseTrack({
    required String token,
    required List<int> trackBytes,
    required String trackFilename,
  }) async {
    final (:request, :abort) = _multipart(
      'POST',
      _uri('/api/v1/workouts/parse-track'),
    );
    request.headers.addAll(_bearerHeaders(token));
    request.files.add(
      http.MultipartFile.fromBytes(
        'track',
        trackBytes,
        filename: trackFilename,
      ),
    );

    final response =
        await _send(request, timeout: _uploadTimeout, onTimeout: abort);

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return ParsedTrackMetadata.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  String mapPreviewUrl(String workoutId, {String? owner, String? objectId}) {
    var uri = _uri('/api/v1/workouts/$workoutId/map-preview');
    final params = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      params['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      params['object_id'] = objectId;
    }
    if (params.isNotEmpty) {
      uri = uri.replace(queryParameters: params);
    }
    return uri.toString();
  }

  String mediaPreviewUrl(
    String workoutId,
    String filename, {
    String? owner,
    String? objectId,
  }) {
    final encoded = Uri.encodeComponent(filename);
    var uri = _uri('/api/v1/workouts/$workoutId/media/$encoded/preview');
    final params = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      params['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      params['object_id'] = objectId;
    }
    if (params.isNotEmpty) {
      uri = uri.replace(queryParameters: params);
    }
    return uri.toString();
  }

  String mediaOriginalUrl(
    String workoutId,
    String filename, {
    String? owner,
    String? objectId,
  }) {
    final encoded = Uri.encodeComponent(filename);
    var uri = _uri('/api/v1/workouts/$workoutId/media/$encoded');
    final params = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      params['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      params['object_id'] = objectId;
    }
    if (params.isNotEmpty) {
      uri = uri.replace(queryParameters: params);
    }
    return uri.toString();
  }

  String workoutTrackUrl(String workoutId, {String? owner, String? objectId}) {
    var uri = _uri('/api/v1/workouts/$workoutId/track');
    final params = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      params['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      params['object_id'] = objectId;
    }
    if (params.isNotEmpty) {
      uri = uri.replace(queryParameters: params);
    }
    return uri.toString();
  }

  Future<WorkoutSpeedSeries> getWorkoutSpeed({
    required String token,
    required String workoutId,
    String? owner,
    String? objectId,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/speed');
    final params = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      params['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      params['object_id'] = objectId;
    }
    if (params.isNotEmpty) {
      uri = uri.replace(queryParameters: params);
    }
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutSpeedSeries.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<WorkoutHeartRateSeries> getWorkoutHeartRate({
    required String token,
    required String workoutId,
    String? owner,
    String? objectId,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/heartrate');
    final params = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      params['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      params['object_id'] = objectId;
    }
    if (params.isNotEmpty) {
      uri = uri.replace(queryParameters: params);
    }
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutHeartRateSeries.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<WorkoutCadenceSeries> getWorkoutCadence({
    required String token,
    required String workoutId,
    String? owner,
    String? objectId,
  }) async {
    var uri = _uri('/api/v1/workouts/$workoutId/cadence');
    final params = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      params['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      params['object_id'] = objectId;
    }
    if (params.isNotEmpty) {
      uri = uri.replace(queryParameters: params);
    }
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      final json = _decode<Map<String, dynamic>>(response);
      return WorkoutCadenceSeries.fromJson(json);
    }

    throw _parseError(response, authenticated: true);
  }

  Future<DownloadedTrack> downloadWorkoutTrack({
    required String token,
    required String workoutId,
    required String fallbackFilename,
    String? owner,
    String? objectId,
    String? format,
  }) async {
    final query = <String, String>{};
    if (owner != null && owner.isNotEmpty) {
      query['owner'] = owner;
    }
    if (objectId != null && objectId.isNotEmpty) {
      query['object_id'] = objectId;
    }
    if (format == 'gpx') {
      query['format'] = 'gpx';
    }
    var uri = _uri('/api/v1/workouts/$workoutId/track');
    if (query.isNotEmpty) {
      uri = uri.replace(queryParameters: query);
    }
    final response = await _request(
      'GET',
      uri,
      headers: _bearerHeaders(token),
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

    throw _parseError(response, authenticated: true);
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
    List<String>? sportTypes,
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
    if (sportTypes != null) {
      params['sport_types'] = sportTypes.join(',');
    }
    final uri = _uri('/api/v1/workouts').replace(queryParameters: params);
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

  Future<void> deleteWorkout({
    required String token,
    required String workoutId,
  }) async {
    final response = await _request(
      'DELETE',
      _uri('/api/v1/workouts/$workoutId'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 204) {
      return;
    }

    throw _parseError(response, authenticated: true);
  }
}
