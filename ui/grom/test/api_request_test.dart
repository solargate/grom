import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/server_storage.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ServerStorage.clear();
  });

  tearDown(() async {
    await ServerStorage.clear();
  });

  test('ServerInfo.fromJson reads name and federation flag', () {
    final info = ServerInfo.fromJson({
      'name': 'Home Lab',
      'federation_enabled': true,
    });
    expect(info.name, 'Home Lab');
    expect(info.federationEnabled, isTrue);

    final defaults = ServerInfo.fromJson({});
    expect(defaults.name, 'Grom Home');
    expect(defaults.federationEnabled, isFalse);
  });

  test('UserInfo.fromJson reads optional avatar fields', () {
    final user = UserInfo.fromJson({
      'id': '1',
      'nickname': 'alice',
      'name': 'Alice',
      'email': 'alice@example.com',
      'has_avatar': true,
      'avatar_url': '/api/v1/users/alice/avatar',
    });
    expect(user.hasAvatar, isTrue);
    expect(user.avatarUrl, '/api/v1/users/alice/avatar');
  });

  test('UserProfile.fromJson reads last sport and equipment', () {
    final profile = UserProfile.fromJson({
      'last_sport_type': 'Ride',
      'last_equipment_by_sport': {
        'Run': ['eq-1', 'eq-2'],
        'Ride': ['bike-1'],
      },
    });
    expect(profile.lastSportType, 'Ride');
    expect(profile.lastEquipmentBySport['Run'], ['eq-1', 'eq-2']);
    expect(profile.lastEquipmentBySport['Ride'], ['bike-1']);

    final empty = UserProfile.fromJson({});
    expect(empty.lastSportType, isNull);
    expect(empty.lastEquipmentBySport, isEmpty);
  });

  test('ApiException keeps status code', () {
    final err = ApiException('boom', statusCode: 409);
    expect(err.toString(), 'boom');
    expect(err.statusCode, 409);
  });

  test('resolveUri and media helpers use cached server base URL', () async {
    await ServerStorage.saveBaseUrl('https://grom.example:8443/');
    final api = ApiRequest(client: http.Client());

    expect(
      api.resolveUri('/api/v1/status').toString(),
      'https://grom.example:8443/api/v1/status',
    );
    expect(
      api.mapPreviewUrl('wid', owner: 'alice'),
      'https://grom.example:8443/api/v1/workouts/wid/map-preview?owner=alice',
    );
    expect(
      api.mediaPreviewUrl('wid', 'shot.png'),
      'https://grom.example:8443/api/v1/workouts/wid/media/shot.png/preview',
    );
    expect(
      api.mediaOriginalUrl('wid', 'shot.png', owner: 'bob'),
      'https://grom.example:8443/api/v1/workouts/wid/media/shot.png?owner=bob',
    );
    expect(
      api.workoutTrackUrl('wid'),
      'https://grom.example:8443/api/v1/workouts/wid/track',
    );
  });

  test('resolveAvatarUrl prefers absolute and relative avatar URLs', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    expect(
      ApiRequest.resolveAvatarUrl(
        nickname: 'alice',
        avatarUrl: 'https://cdn.example/a.webp',
      ),
      'https://cdn.example/a.webp',
    );
    expect(
      ApiRequest.resolveAvatarUrl(
        nickname: 'alice',
        avatarUrl: '/api/v1/users/alice/avatar',
      ),
      'https://grom.example/api/v1/users/alice/avatar',
    );
    expect(
      ApiRequest.resolveAvatarUrl(nickname: 'alice', hasAvatar: false),
      '',
    );
    expect(
      ApiRequest.resolveAvatarUrl(nickname: 'alice', hasAvatar: true),
      'https://grom.example/api/v1/users/alice/avatar',
    );
  });

	test('getWorkout requests owner query and parses body', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient((request) async {
      expect(request.method, 'GET');
      expect(request.url.path, '/api/v1/workouts/wid');
      expect(request.url.queryParameters['owner'], 'alice');
      expect(request.headers['Authorization'], 'Bearer tok');
      return http.Response(
        jsonEncode({
          'id': 'wid',
          'owner': 'alice',
          'name': 'Morning run',
          'sport_type': 'Run',
          'start_date': '2026-07-08T10:00:00Z',
          'duration_seconds': 1800,
          'distance': 5000,
          'has_map_preview': false,
          'has_media': false,
          'author': {
            'nickname': 'alice',
            'name': 'Alice',
            'handle': 'alice@grom.example',
            'is_local': true,
            'has_avatar': false,
          },
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final workout = await ApiRequest(client: client).getWorkout(
      token: 'tok',
      workoutId: 'wid',
      owner: 'alice',
    );
    expect(workout.id, 'wid');
    expect(workout.owner, 'alice');
    expect(workout.name, 'Morning run');
  });

  test('listWorkouts parses cursor page envelope', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient((request) async {
      expect(request.method, 'GET');
      expect(request.url.path, '/api/v1/workouts');
      expect(request.url.queryParameters['limit'], '20');
      expect(request.url.queryParameters['scope'], 'own');
      expect(request.url.queryParameters['cursor'], 'abc');
      return http.Response(
        jsonEncode({
          'items': [
            {
              'id': 'wid',
              'name': 'Morning run',
              'sport_type': 'Run',
              'start_date': '2026-07-08T10:00:00Z',
              'duration_seconds': 1800,
              'distance': 5000,
            },
          ],
          'next_cursor': 'next',
          'has_more': true,
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final page = await ApiRequest(client: client).listWorkouts(
      'tok',
      scope: 'own',
      cursor: 'abc',
    );
    expect(page.items, hasLength(1));
    expect(page.items.first.id, 'wid');
    expect(page.nextCursor, 'next');
    expect(page.hasMore, isTrue);
  });

  test('getServerInfo parses JSON and falls back on errors', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final okClient = MockClient((request) async {
      expect(request.url.path, '/api/v1/server-info');
      return http.Response(
        jsonEncode({'name': 'Lab', 'federation_enabled': true}),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final ok = await ApiRequest(client: okClient).getServerInfo();
    expect(ok.name, 'Lab');
    expect(ok.federationEnabled, isTrue);

    final failClient = MockClient((request) async {
      return http.Response('nope', 500);
    });
    final fallback = await ApiRequest(client: failClient).getServerInfo();
    expect(fallback.name, 'Grom Home');
  });

  test('login parses success and maps API errors', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final okClient = MockClient((request) async {
      expect(request.method, 'POST');
      return http.Response(
        jsonEncode({
          'token': 'tok',
          'expires_at': '2026-07-25T00:00:00Z',
          'user': {
            'id': '1',
            'nickname': 'alice',
            'name': 'Alice',
            'email': 'alice@example.com',
          },
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final login = await ApiRequest(client: okClient).login(
      email: 'alice@example.com',
      password: 'secret',
    );
    expect(login.token, 'tok');
    expect(login.user.nickname, 'alice');

    final errClient = MockClient((request) async {
      return http.Response(
        jsonEncode({'error': 'invalid credentials'}),
        401,
        headers: {'content-type': 'application/json'},
      );
    });
    try {
      await ApiRequest(client: errClient).login(
        email: 'alice@example.com',
        password: 'bad',
      );
      fail('expected ApiException');
    } on ApiException catch (e) {
      expect(e.message, 'invalid credentials');
      expect(e.statusCode, 401);
    }
  });

  test('getStravaImportStatus returns decoded map', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient((request) async {
      expect(request.headers['Authorization'], 'Bearer tok');
      return http.Response(
        jsonEncode({
          'active': true,
          'phase': 'importing',
          'import_progress': 0.5,
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final status = await ApiRequest(client: client).getStravaImportStatus('tok');
    expect(status['phase'], 'importing');
    expect(status['import_progress'], 0.5);
  });

  test('getWorkoutSpeed requests owner and parses samples', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient((request) async {
      expect(request.method, 'GET');
      expect(request.url.path, '/api/v1/workouts/wid/speed');
      expect(request.url.queryParameters['owner'], 'alice');
      expect(request.headers['Authorization'], 'Bearer tok');
      return http.Response(
        jsonEncode({
          'samples': [
            {
              't': '2026-07-08T10:00:00Z',
              'speed_kmh': 12.5,
              'distance_m': 10,
            },
            {
              't': '2026-07-08T10:00:10Z',
              'speed_kmh': 13.0,
              'distance_m': 40,
            },
          ],
          'speed_avg_kmh': 12.7,
          'speed_max_kmh': 13.0,
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final series = await ApiRequest(client: client).getWorkoutSpeed(
      token: 'tok',
      workoutId: 'wid',
      owner: 'alice',
    );
    expect(series.samples, hasLength(2));
    expect(series.samples.first.speedKmh, 12.5);
    expect(series.speedMaxKmh, 13.0);
  });

  test('getWorkoutHeartRate parses has_gps false and empty distance', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient((request) async {
      expect(request.url.path, '/api/v1/workouts/wid/heartrate');
      return http.Response(
        jsonEncode({
          'samples': [
            {'t': '2026-07-08T10:00:00Z', 'heart_rate_bpm': 110},
            {'t': '2026-07-08T10:01:00Z', 'heart_rate_bpm': 140},
          ],
          'has_gps': false,
          'heart_rate_avg': 125,
          'heart_rate_max': 140,
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final series = await ApiRequest(client: client).getWorkoutHeartRate(
      token: 'tok',
      workoutId: 'wid',
    );
    expect(series.hasGps, isFalse);
    expect(series.samples.first.distanceM, isNull);
    expect(series.heartRateMax, 140);
  });

  test('hasExternalID returns exists flag from API', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient((request) async {
      expect(request.url.path, '/api/v1/workouts/external');
      expect(request.url.queryParameters['name'], 'health-sync/strava');
      expect(request.url.queryParameters['id'], 'CYCLING 2026.07.30 16.26 Strava.csv');
      return http.Response(
        jsonEncode({'exists': true}),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    final exists = await ApiRequest(client: client).hasExternalID(
      token: 'tok',
      name: 'health-sync/strava',
      id: 'CYCLING 2026.07.30 16.26 Strava.csv',
    );
    expect(exists, isTrue);
  });

  test('createWorkoutMultipart sends external_id fields', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient.streaming((request, bodyStream) async {
      expect(request.url.path, '/api/v1/workouts');
      expect(request.method, 'POST');
      expect(request, isA<http.MultipartRequest>());
      final multipart = request as http.MultipartRequest;
      expect(multipart.fields['external_id_name'], 'health-sync/strava');
      expect(multipart.fields['external_id_id'], 'CYCLING 2026.07.30 16.26 Strava.csv');
      expect(multipart.fields['sport_type'], 'Ride');
      await bodyStream.drain();
      final body = jsonEncode({
        'id': 'wid-1',
        'name': 'Synced',
        'description': '',
        'sport_type': 'Ride',
        'start_date': '2026-07-30T16:26:00Z',
        'duration_seconds': 0,
        'distance': 0,
      });
      return http.StreamedResponse(
        Stream.value(utf8.encode(body)),
        201,
        headers: {'content-type': 'application/json'},
      );
    });

    final workout = await ApiRequest(client: client).createWorkoutMultipart(
      token: 'tok',
      fields: {
        'name': 'Synced',
        'sport_type': 'Ride',
        'start_date': '2026-07-30T16:26:00Z',
        'external_id_name': 'health-sync/strava',
        'external_id_id': 'CYCLING 2026.07.30 16.26 Strava.csv',
      },
    );
    expect(workout.id, 'wid-1');
    expect(workout.sportType, 'Ride');
  });

  test('getWorkoutSpeed maps API errors', () async {
    await ServerStorage.saveBaseUrl('https://grom.example');
    final client = MockClient((request) async {
      return http.Response(
        jsonEncode({'error': 'workout not found'}),
        404,
        headers: {'content-type': 'application/json'},
      );
    });
    expect(
      () => ApiRequest(client: client).getWorkoutSpeed(
        token: 'tok',
        workoutId: 'missing',
      ),
      throwsA(
        isA<ApiException>().having((e) => e.statusCode, 'statusCode', 404),
      ),
    );
  });
}
