import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_client.dart';
import 'package:grom/services/strava_api_constants.dart';
import 'package:grom/services/strava_api_streams.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  test('listAthleteActivities parses page and sends per_page', () async {
    http.Request? seen;
    final client = StravaApiClient(
      httpClient: MockClient((request) async {
        seen = request;
        return http.Response(
          jsonEncode([
            {
              'id': 11,
              'name': 'Run',
              'sport_type': 'Run',
              'type': 'Run',
              'start_date': '2026-09-05T10:00:00Z',
              'moving_time': 100,
              'elapsed_time': 110,
              'distance': 1000,
              'map': {'summary_polyline': 'abc'},
              'device_name': 'Garmin Edge 1030',
            },
          ]),
          200,
          headers: {'content-type': 'application/json'},
        );
      }),
    );

    final list = await client.listAthleteActivities(
      accessToken: 'tok',
      perPage: 25,
      page: 2,
    );
    expect(list, hasLength(1));
    expect(list.first.id, 11);
    expect(list.first.hasMapPolyline, isTrue);
    expect(list.first.deviceName, 'Garmin Edge 1030');
    expect(seen!.url.queryParameters['per_page'], '25');
    expect(seen!.url.queryParameters['page'], '2');
    expect(seen!.headers['authorization'], 'Bearer tok');
    expect(seen!.headers['user-agent'], kStravaHttpUserAgent);
  });

  test('listAthleteActivities throws StravaApiException on error', () async {
    final client = StravaApiClient(
      httpClient: MockClient(
        (_) async => http.Response(
          jsonEncode({'message': 'Rate Limit Exceeded'}),
          429,
        ),
      ),
    );
    expect(
      () => client.listAthleteActivities(accessToken: 'tok'),
      throwsA(
        isA<StravaApiException>()
            .having((e) => e.statusCode, 'statusCode', 429)
            .having((e) => e.message, 'message', 'Rate Limit Exceeded'),
      ),
    );
  });

  test('getActivityStreamSamples keeps HR when latlng has gaps', () async {
    http.Request? seen;
    final client = StravaApiClient(
      httpClient: MockClient((request) async {
        seen = request;
        return http.Response(
          jsonEncode({
            'latlng': {
              'data': [
                [55.75, 37.61],
                [0, 0],
                [55.76, 37.62],
              ],
            },
            'time': {
              'data': [0, 30, 60],
            },
            'altitude': {
              'data': [100, 101, 102],
            },
            'heartrate': {
              'data': [120, 130, 145],
            },
          }),
          200,
        );
      }),
    );

    final samples = await client.getActivityStreamSamples(
      accessToken: 'tok',
      activityId: 9,
    );
    expect(seen!.url.queryParameters['keys'], kStravaStreamKeys);
    expect(samples, hasLength(3));
    expect(samples[0].lat, 55.75);
    expect(samples[0].heartRateBpm, 120);
    expect(samples[1].hasGps, isFalse);
    expect(samples[1].heartRateBpm, 130);
    expect(samples[2].lon, 37.62);
    expect(samples[2].heartRateBpm, 145);
  });

  test('getActivityStreamSamples works without latlng (HR only)', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response(
            jsonEncode({
              'time': {
                'data': [0, 30, 60],
              },
              'heartrate': {
                'data': [110, 120, 130],
              },
            }),
            200,
          )),
    );

    final samples = await client.getActivityStreamSamples(
      accessToken: 'tok',
      activityId: 9,
    );
    expect(samples, hasLength(3));
    expect(samples.every((s) => !s.hasGps), isTrue);
    expect(samples.map((s) => s.heartRateBpm).toList(), [110, 120, 130]);
  });

  test('getActivityStreamSamples omits HR when stream missing', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response(
            jsonEncode({
              'latlng': {
                'data': [
                  [55.75, 37.61],
                  [55.76, 37.62],
                ],
              },
              'time': {
                'data': [0, 60],
              },
            }),
            200,
          )),
    );

    final samples = await client.getActivityStreamSamples(
      accessToken: 'tok',
      activityId: 9,
    );
    expect(samples, hasLength(2));
    expect(samples.every((p) => p.heartRateBpm == null), isTrue);
  });

  test('getActivityStreamSamples handles short HR stream and rejects negative',
      () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response(
            jsonEncode({
              'latlng': {
                'data': [
                  [55.75, 37.61],
                  [55.76, 37.62],
                  [55.77, 37.63],
                ],
              },
              'time': {
                'data': [0, 30, 60],
              },
              'heartrate': {
                'data': [120, -1],
              },
            }),
            200,
          )),
    );

    final samples = await client.getActivityStreamSamples(
      accessToken: 'tok',
      activityId: 9,
    );
    expect(samples, hasLength(3));
    expect(samples[0].heartRateBpm, 120);
    expect(samples[1].heartRateBpm, isNull);
    expect(samples[2].heartRateBpm, isNull);
  });

  test('getActivityStreamSamples reads HR from array-shaped streams', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response(
            jsonEncode([
              {
                'type': 'latlng',
                'data': [
                  [55.75, 37.61],
                  [55.76, 37.62],
                ],
              },
              {
                'type': 'time',
                'data': [0, 60],
              },
              {
                'type': 'heartrate',
                'data': [110, 130],
              },
            ]),
            200,
          )),
    );

    final samples = await client.getActivityStreamSamples(
      accessToken: 'tok',
      activityId: 9,
    );
    expect(samples, hasLength(2));
    expect(samples.first.heartRateBpm, 110);
    expect(samples.last.heartRateBpm, 130);
  });

  test('getActivityStreamSamples returns empty on 404', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response('missing', 404)),
    );
    expect(
      await client.getActivityStreamSamples(accessToken: 'tok', activityId: 1),
      isEmpty,
    );
  });

  test('getActivity parses elevation and sensor summary fields', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response(
            jsonEncode({
              'id': 99,
              'name': 'Climb',
              'sport_type': 'Ride',
              'type': 'Ride',
              'start_date': '2026-09-05T10:00:00Z',
              'moving_time': 1000,
              'elapsed_time': 1100,
              'distance': 15000,
              'total_elevation_gain': 516,
              'elev_low': 10.5,
              'elev_high': 200.25,
              'average_heartrate': 140.2,
              'max_heartrate': 178,
              'average_cadence': 82.5,
              'average_watts': 190,
              'max_watts': 400,
              'calories': 512.4,
            }),
            200,
            headers: {'content-type': 'application/json'},
          )),
    );
    final activity = await client.getActivity(
      accessToken: 'tok',
      activityId: 99,
    );
    expect(activity.totalElevationGain, 516);
    expect(activity.elevLow, 10.5);
    expect(activity.elevHigh, 200.25);
    expect(activity.averageHeartrate, 140.2);
    expect(activity.maxHeartrate, 178);
    expect(activity.averageCadence, 82.5);
    expect(activity.averageWatts, 190);
    expect(activity.maxWatts, 400);
    expect(activity.calories, 512.4);
  });

  test('listActivityPhotos picks largest url size', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response(
            jsonEncode([
              {
                'unique_id': 'p1',
                'urls': {
                  '100': 'https://example.com/small.jpg',
                  '2048': 'https://example.com/large.jpg',
                },
              },
            ]),
            200,
          )),
    );
    final photos = await client.listActivityPhotos(
      accessToken: 'tok',
      activityId: 3,
    );
    expect(photos, hasLength(1));
    expect(photos.first.uniqueId, 'p1');
    expect(photos.first.url, 'https://example.com/large.jpg');
  });

  test('listActivityPhotos returns empty on failure', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response('nope', 500)),
    );
    expect(
      await client.listActivityPhotos(accessToken: 'tok', activityId: 3),
      isEmpty,
    );
  });

  test('parseStravaStreamsByType reads velocity cadence watts', () {
    final samples = parseStravaStreamsByType({
      'time': {
        'data': [0, 1],
      },
      'velocity_smooth': {
        'data': [2.5, 3.0],
      },
      'cadence': {
        'data': [80, 90],
      },
      'watts': {
        'data': [150, 200],
      },
      'distance': {
        'data': [0.0, 3.0],
      },
    });
    expect(samples, hasLength(2));
    expect(samples.first.speedMps, 2.5);
    expect(samples.last.cadenceRpm, 90);
    expect(samples.last.watts, 200);
    expect(samples.last.distanceMeters, 3.0);
  });
}
