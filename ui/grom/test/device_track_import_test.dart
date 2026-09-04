import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/platform/track_file_picker.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/services/device_track_import_id.dart';
import 'package:grom/services/device_track_import_service.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ServerStorage.clear();
    await ServerStorage.saveBaseUrl('https://grom.example');
  });

  tearDown(() async {
    await ServerStorage.clear();
  });

  group('deviceImportExternalID', () {
    test('is stable for same filename and bytes', () {
      final bytes = utf8.encode('track-bytes');
      final a = deviceImportExternalID(filename: 'Ride.FIT', bytes: bytes);
      final b = deviceImportExternalID(filename: 'ride.fit', bytes: bytes);
      expect(a, b);
      expect(a, startsWith('ride.fit:'));
      expect(a.split(':').last.length, 16);
    });

    test('differs when content changes', () {
      final a = deviceImportExternalID(
        filename: 'a.gpx',
        bytes: utf8.encode('one'),
      );
      final b = deviceImportExternalID(
        filename: 'a.gpx',
        bytes: utf8.encode('two'),
      );
      expect(a, isNot(b));
    });
  });

  group('isTrackFilename', () {
    test('accepts gpx and fit case-insensitively', () {
      expect(isTrackFilename('x.GPX'), isTrue);
      expect(isTrackFilename('x.fit'), isTrue);
      expect(isTrackFilename('x.csv'), isFalse);
    });
  });

  group('DeviceTrackImportState', () {
    test('progress reflects current over total', () {
      const state = DeviceTrackImportState(
        active: true,
        importCurrent: 2,
        importTotal: 4,
      );
      expect(state.importProgress, 0.5);
      expect(state.showImportProgress, isTrue);
    });
  });

  group('DeviceTrackImportService.importPickedFiles', () {
    TrackPickResult pick(String name, String content) {
      return (
        filename: name,
        bytes: Uint8List.fromList(utf8.encode(content)),
      );
    }

    http.StreamedResponse jsonResponse(Object body, {int status = 200}) {
      return http.StreamedResponse(
        Stream.value(utf8.encode(jsonEncode(body))),
        status,
        headers: {'content-type': 'application/json'},
      );
    }

    test('aggregates created skipped invalid failed and continues mid-batch',
        () async {
      final createdFile = pick('ok.gpx', 'created-bytes');
      final skippedFile = pick('dup.fit', 'skipped-bytes');
      final invalidFile = pick('notes.csv', 'not-a-track');
      final failedParseFile = pick('bad.gpx', 'parse-fail');
      final missingStartFile = pick('nostart.gpx', 'no-start');
      final afterFailFile = pick('after.gpx', 'after-fail');

      final skippedId = deviceImportExternalID(
        filename: skippedFile.filename,
        bytes: skippedFile.bytes,
      );

      String? createdSport;
      final progressSnapshots = <int>[];

      final client = MockClient.streaming((request, bodyStream) async {
        await bodyStream.drain();
        final path = request.url.path;

        if (path == '/api/v1/profile') {
          return jsonResponse({'last_sport_type': 'Walk'});
        }
        if (path == '/api/v1/workouts/external') {
          final id = request.url.queryParameters['id'];
          return jsonResponse({'exists': id == skippedId});
        }
        if (path == '/api/v1/workouts/parse-track') {
          final multipart = request as http.MultipartRequest;
          final filename = multipart.files.first.filename;
          if (filename == 'bad.gpx') {
            return jsonResponse({'error': 'invalid track'}, status: 400);
          }
          if (filename == 'nostart.gpx') {
            return jsonResponse({
              'has_gps': false,
              'duration_seconds': 60,
              'sport_type': 'Ride',
            });
          }
          if (filename == 'ok.gpx') {
            return jsonResponse({
              'name': 'From track',
              'sport_type': 'Ride',
              'start_date': '2026-07-06T08:40:00Z',
              'duration_seconds': 120,
              'has_gps': true,
            });
          }
          // after.gpx — no sport_type → profile last_sport_type Walk
          return jsonResponse({
            'start_date': '2026-07-07T08:40:00Z',
            'duration_seconds': 90,
            'has_gps': false,
          });
        }
        if (path == '/api/v1/workouts' && request.method == 'POST') {
          final multipart = request as http.MultipartRequest;
          createdSport ??= multipart.fields['sport_type'];
          if (multipart.files.any((f) => f.filename == 'ok.gpx')) {
            expect(multipart.fields['sport_type'], 'Ride');
            expect(multipart.fields['name'], 'From track');
          }
          if (multipart.files.any((f) => f.filename == 'after.gpx')) {
            expect(multipart.fields['sport_type'], 'Walk');
          }
          return jsonResponse({
            'id': 'wid-1',
            'name': multipart.fields['name'] ?? 'w',
            'description': '',
            'sport_type': multipart.fields['sport_type'] ?? 'Run',
            'start_date':
                multipart.fields['start_date'] ?? '2026-07-06T08:40:00Z',
            'duration_seconds': 0,
            'distance': 0,
          }, status: 201);
        }

        return jsonResponse({'error': 'unexpected ${request.method} $path'},
            status: 500);
      });

      final service = DeviceTrackImportService.forTesting(
        api: ApiRequest(client: client),
        tokenProvider: () async => 'tok',
      );
      service.addListener(() {
        if (service.state.active || service.state.completed) {
          progressSnapshots.add(service.state.importCurrent);
        }
      });

      final result = await service.importPickedFiles([
        createdFile,
        skippedFile,
        invalidFile,
        failedParseFile,
        missingStartFile,
        afterFailFile,
      ]);

      expect(result.completed, isTrue);
      expect(result.active, isFalse);
      expect(result.importTotal, 6);
      expect(result.importCurrent, 6);
      expect(result.created, 2);
      expect(result.skipped, 1);
      expect(result.invalid, 1);
      expect(result.failed, 2); // bad.gpx parse + nostart.gpx
      expect(progressSnapshots, contains(6));
      expect(createdSport, isNotNull);
    });

    test('uses default Run when profile and track sport are absent', () async {
      String? sport;
      final client = MockClient.streaming((request, bodyStream) async {
        await bodyStream.drain();
        final path = request.url.path;
        if (path == '/api/v1/profile') {
          return jsonResponse({});
        }
        if (path == '/api/v1/workouts/external') {
          return jsonResponse({'exists': false});
        }
        if (path == '/api/v1/workouts/parse-track') {
          return jsonResponse({
            'start_date': '2026-07-06T08:40:00Z',
            'duration_seconds': 10,
            'has_gps': false,
          });
        }
        if (path == '/api/v1/workouts') {
          final multipart = request as http.MultipartRequest;
          sport = multipart.fields['sport_type'];
          return jsonResponse({
            'id': 'wid-2',
            'name': 'a',
            'description': '',
            'sport_type': sport,
            'start_date': '2026-07-06T08:40:00Z',
            'duration_seconds': 10,
            'distance': 0,
          }, status: 201);
        }
        return jsonResponse({'error': 'unexpected'}, status: 500);
      });

      final service = DeviceTrackImportService.forTesting(
        api: ApiRequest(client: client),
        tokenProvider: () async => 'tok',
      );
      final result = await service.importPickedFiles([
        pick('a.gpx', 'bytes'),
      ]);
      expect(result.created, 1);
      expect(sport, 'Run');
    });

    test('returns unchanged when token is missing', () async {
      final service = DeviceTrackImportService.forTesting(
        api: ApiRequest(client: MockClient((_) async => http.Response('', 500))),
        tokenProvider: () async => null,
      );
      final result = await service.importPickedFiles([
        pick('a.gpx', 'bytes'),
      ]);
      expect(result.completed, isFalse);
      expect(result.created, 0);
    });
  });
}
