import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/services/strava_api_auth.dart';
import 'package:grom/services/strava_api_client.dart';
import 'package:grom/services/strava_api_storage.dart';
import 'package:grom/services/strava_api_sync_service.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

Map<String, dynamic> _workoutJson(String id) => {
      'id': id,
      'name': 'Workout',
      'sport_type': 'Run',
      'start_date': '2026-09-05T10:00:00Z',
      'duration_seconds': 100,
      'distance': 1000.0,
      'has_media': false,
      'media_files': <String>[],
    };

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({
      stravaApiEnabledStorageKey: true,
      stravaApiClientIdStorageKey: '1',
      stravaApiClientSecretStorageKey: 'secret',
      stravaApiConnectedStorageKey: true,
      stravaApiAccessTokenStorageKey: 'access',
      stravaApiRefreshTokenStorageKey: 'refresh',
      stravaApiExpiresAtStorageKey:
          DateTime.now().millisecondsSinceEpoch ~/ 1000 + 3600,
    });
    await ServerStorage.clear();
    await ServerStorage.saveBaseUrl('https://grom.test');
  });

  tearDown(() async {
    await ServerStorage.clear();
  });

  test('sync stops at first already-imported activity', () async {
    final existingIds = <String>{'102'};
    var createCount = 0;

    final gromClient = MockClient((request) async {
      if (request.url.path.endsWith('/workouts/external')) {
        final id = request.url.queryParameters['id'] ?? '';
        return http.Response(
          jsonEncode({'exists': existingIds.contains(id)}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.method == 'POST' && request.url.path.endsWith('/workouts')) {
        createCount++;
        return http.Response(
          jsonEncode(_workoutJson('w$createCount')),
          201,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected grom request ${request.method} ${request.url}');
    });

    final stravaClient = MockClient((request) async {
      if (request.url.path.endsWith('/athlete/activities')) {
        return http.Response(
          jsonEncode([
            {
              'id': 103,
              'name': 'Newest',
              'sport_type': 'Run',
              'type': 'Run',
              'start_date': '2026-09-05T12:00:00Z',
              'moving_time': 600,
              'elapsed_time': 620,
              'distance': 2000,
              'total_photo_count': 0,
            },
            {
              'id': 102,
              'name': 'Already imported',
              'sport_type': 'Run',
              'type': 'Run',
              'start_date': '2026-09-05T11:00:00Z',
              'moving_time': 600,
              'elapsed_time': 620,
              'distance': 2000,
              'total_photo_count': 0,
            },
            {
              'id': 101,
              'name': 'Older should not run',
              'sport_type': 'Run',
              'type': 'Run',
              'start_date': '2026-09-05T10:00:00Z',
              'moving_time': 600,
              'elapsed_time': 620,
              'distance': 2000,
              'total_photo_count': 0,
            },
          ]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.url.path.contains('/streams')) {
        return http.Response('[]', 404);
      }
      if (RegExp(r'/activities/\d+$').hasMatch(request.url.path)) {
        final id = int.parse(request.url.pathSegments.last);
        return http.Response(
          jsonEncode({
            'id': id,
            'name': 'Activity $id',
            'sport_type': 'Run',
            'type': 'Run',
            'start_date': '2026-09-05T12:00:00Z',
            'moving_time': 600,
            'elapsed_time': 620,
            'distance': 2000,
            'total_photo_count': 0,
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected strava request ${request.url}');
    });

    final service = StravaApiSyncService.forTesting(
      api: ApiRequest(client: gromClient),
      auth: StravaApiAuth(
        httpClient: MockClient((_) async {
          fail('token refresh should not be needed');
        }),
      ),
      client: StravaApiClient(httpClient: stravaClient),
      tokenProvider: () async => 'grom-jwt',
    );
    await service.loadFromStorage();

    final result = await service.syncWorkouts();
    expect(result.kind, StravaApiSyncResultKind.imported);
    expect(result.importedCount, 1);
    expect(createCount, 1);
  });
}
