import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/services/strava_api_auth.dart';
import 'package:grom/services/strava_api_client.dart';
import 'package:grom/services/strava_api_constants.dart';
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

Map<String, dynamic> _activityJson(int id, {String name = 'Activity'}) => {
      'id': id,
      'name': name,
      'sport_type': 'Run',
      'type': 'Run',
      'start_date': '2026-09-05T12:00:00Z',
      'moving_time': 600,
      'elapsed_time': 620,
      'distance': 2000,
      'total_photo_count': 0,
    };

Future<void> _seedConnectedPrefs({int? syncLimit}) async {
  SharedPreferences.setMockInitialValues({
    stravaApiEnabledStorageKey: true,
    stravaApiClientIdStorageKey: '1',
    stravaApiClientSecretStorageKey: 'secret',
    stravaApiConnectedStorageKey: true,
    stravaApiAccessTokenStorageKey: 'access',
    stravaApiRefreshTokenStorageKey: 'refresh',
    stravaApiExpiresAtStorageKey:
        DateTime.now().millisecondsSinceEpoch ~/ 1000 + 3600,
    if (syncLimit != null) stravaApiSyncLimitStorageKey: syncLimit,
  });
  await ServerStorage.clear();
  await ServerStorage.saveBaseUrl('https://grom.test');
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  tearDown(() async {
    await ServerStorage.clear();
  });

  test('sync stops at first already-imported activity', () async {
    await _seedConnectedPrefs();
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
            _activityJson(103, name: 'Newest'),
            _activityJson(102, name: 'Already imported'),
            _activityJson(101, name: 'Older should not run'),
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
          jsonEncode(_activityJson(id)),
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

  test('sync returns noNewWorkouts when newest already exists', () async {
    await _seedConnectedPrefs();
    final gromClient = MockClient((request) async {
      if (request.url.path.endsWith('/workouts/external')) {
        return http.Response(
          jsonEncode({'exists': true}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected grom request ${request.url}');
    });
    final stravaClient = MockClient((request) async {
      if (request.url.path.endsWith('/athlete/activities')) {
        return http.Response(
          jsonEncode([_activityJson(1)]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected strava request ${request.url}');
    });

    final service = StravaApiSyncService.forTesting(
      api: ApiRequest(client: gromClient),
      auth: StravaApiAuth(
        httpClient: MockClient((_) async => fail('no refresh')),
      ),
      client: StravaApiClient(httpClient: stravaClient),
      tokenProvider: () async => 'grom-jwt',
    );
    await service.loadFromStorage();
    final result = await service.syncWorkouts();
    expect(result.kind, StravaApiSyncResultKind.noNewWorkouts);
  });

  test('sync uses stored sync limit as per_page', () async {
    await _seedConnectedPrefs(syncLimit: 7);
    http.Request? listRequest;
    final gromClient = MockClient((request) async {
      if (request.url.path.endsWith('/workouts/external')) {
        return http.Response(
          jsonEncode({'exists': true}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected ${request.url}');
    });
    final stravaClient = MockClient((request) async {
      if (request.url.path.endsWith('/athlete/activities')) {
        listRequest = request;
        return http.Response(
          jsonEncode([_activityJson(1)]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected ${request.url}');
    });

    final service = StravaApiSyncService.forTesting(
      api: ApiRequest(client: gromClient),
      auth: StravaApiAuth(
        httpClient: MockClient((_) async => fail('no refresh')),
      ),
      client: StravaApiClient(httpClient: stravaClient),
      tokenProvider: () async => 'grom-jwt',
    );
    await service.loadFromStorage();
    expect(service.syncLimit, 7);
    await service.syncWorkouts();
    expect(listRequest!.url.queryParameters['per_page'], '7');
  });

  test('sync returns notEnabled and notConnected', () async {
    SharedPreferences.setMockInitialValues({
      stravaApiEnabledStorageKey: false,
      stravaApiConnectedStorageKey: false,
    });
    final service = StravaApiSyncService.forTesting(
      api: ApiRequest(
        client: MockClient((_) async => fail('no network')),
      ),
      auth: StravaApiAuth(
        httpClient: MockClient((_) async => fail('no network')),
      ),
      client: StravaApiClient(
        httpClient: MockClient((_) async => fail('no network')),
      ),
      tokenProvider: () async => 'grom-jwt',
    );
    await service.loadFromStorage();
    expect(
      (await service.syncWorkouts()).kind,
      StravaApiSyncResultKind.notEnabled,
    );

    await service.setEnabled(true);
    expect(
      (await service.syncWorkouts()).kind,
      StravaApiSyncResultKind.notConnected,
    );
  });

  test('sync maps Strava 401 to authFailed', () async {
    await _seedConnectedPrefs();
    final service = StravaApiSyncService.forTesting(
      api: ApiRequest(
        client: MockClient((_) async => fail('grom unused')),
      ),
      auth: StravaApiAuth(
        httpClient: MockClient((_) async => fail('no refresh')),
      ),
      client: StravaApiClient(
        httpClient: MockClient(
          (_) async => http.Response(
            jsonEncode({'message': 'Authorization Error'}),
            401,
          ),
        ),
      ),
      tokenProvider: () async => 'grom-jwt',
    );
    await service.loadFromStorage();
    final result = await service.syncWorkouts();
    expect(result.kind, StravaApiSyncResultKind.authFailed);
    expect(result.message, 'Authorization Error');
  });

  test('sync imports without track when streams unavailable', () async {
    await _seedConnectedPrefs();
    Map<String, String>? createFields;
    final gromClient = MockClient((request) async {
      if (request.url.path.endsWith('/workouts/external')) {
        return http.Response(
          jsonEncode({'exists': false}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.method == 'POST' && request.url.path.endsWith('/workouts')) {
        // Multipart: inspect body for external id field markers.
        final body = request.body;
        createFields = {
          'has_track': body.contains('filename="track"') ||
                  body.contains('name="track"')
              ? 'yes'
              : 'no',
          'external': body.contains('strava') ? 'yes' : 'no',
        };
        return http.Response(
          jsonEncode(_workoutJson('w1')),
          201,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected ${request.url}');
    });
    final stravaClient = MockClient((request) async {
      if (request.url.path.endsWith('/athlete/activities')) {
        return http.Response(
          jsonEncode([_activityJson(55, name: 'Indoor')]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.url.path.contains('/streams')) {
        return http.Response('[]', 404);
      }
      if (RegExp(r'/activities/\d+$').hasMatch(request.url.path)) {
        return http.Response(
          jsonEncode(_activityJson(55, name: 'Indoor')),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      fail('unexpected ${request.url}');
    });

    final service = StravaApiSyncService.forTesting(
      api: ApiRequest(client: gromClient),
      auth: StravaApiAuth(
        httpClient: MockClient((_) async => fail('no refresh')),
      ),
      client: StravaApiClient(httpClient: stravaClient),
      tokenProvider: () async => 'grom-jwt',
    );
    await service.loadFromStorage();
    final result = await service.syncWorkouts();
    expect(result.kind, StravaApiSyncResultKind.imported);
    expect(result.importedCount, 1);
    expect(createFields!['has_track'], 'no');
    expect(createFields!['external'], 'yes');
  });

  test('stravaApiSyncResultSnackBarMessage maps kinds', () {
    expect(
      stravaApiSyncResultSnackBarMessage(
        const StravaApiSyncResult(
          kind: StravaApiSyncResultKind.imported,
          importedCount: 3,
        ),
        imported: (c) => 'imported $c',
        noNewWorkouts: 'none',
        notConnected: 'nc',
        notEnabled: 'ne',
        authFailed: 'af',
        cancelled: 'ca',
        syncError: (m) => 'err $m',
      ),
      'imported 3',
    );
    expect(
      stravaApiSyncResultSnackBarMessage(
        const StravaApiSyncResult(
          kind: StravaApiSyncResultKind.error,
          message: 'boom',
        ),
        imported: (_) => '',
        noNewWorkouts: 'none',
        notConnected: 'nc',
        notEnabled: 'ne',
        authFailed: 'af',
        cancelled: 'ca',
        syncError: (m) => 'err $m',
      ),
      'err boom',
    );
    expect(
      stravaApiSyncResultSnackBarMessage(
        const StravaApiSyncResult(kind: StravaApiSyncResultKind.notEnabled),
        imported: (_) => '',
        noNewWorkouts: 'none',
        notConnected: 'nc',
        notEnabled: 'ne',
        authFailed: 'af',
        cancelled: 'ca',
        syncError: (m) => 'err $m',
      ),
      'ne',
    );
  });

  test('clampStravaApiSyncLimit bounds and default', () {
    expect(clampStravaApiSyncLimit(null), kStravaApiSyncLimitDefault);
    expect(clampStravaApiSyncLimit(0), kStravaApiSyncLimitDefault);
    expect(clampStravaApiSyncLimit(1), 1);
    expect(clampStravaApiSyncLimit(10), 10);
    expect(clampStravaApiSyncLimit(200), 200);
    expect(clampStravaApiSyncLimit(201), kStravaApiSyncLimitMax);
  });
}
