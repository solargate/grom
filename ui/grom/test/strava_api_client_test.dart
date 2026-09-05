import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_client.dart';
import 'package:grom/services/strava_api_constants.dart';
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

  test('getActivityTrackPoints reads key_by_type streams', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response(
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
            }),
            200,
          )),
    );

    final points = await client.getActivityTrackPoints(
      accessToken: 'tok',
      activityId: 9,
    );
    expect(points, hasLength(2));
    expect(points.first.lat, 55.75);
    expect(points.first.timeSeconds, 0);
    expect(points.first.elevation, 100);
    expect(points.last.lon, 37.62);
  });

  test('getActivityTrackPoints returns empty on 404', () async {
    final client = StravaApiClient(
      httpClient: MockClient((_) async => http.Response('missing', 404)),
    );
    expect(
      await client.getActivityTrackPoints(accessToken: 'tok', activityId: 1),
      isEmpty,
    );
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
}
