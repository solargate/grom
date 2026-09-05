import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/platform/shared_track_receive.dart';

void main() {
  Future<Uint8List> readMap(Map<String, Uint8List> files, String path) async {
    final bytes = files[path];
    if (bytes == null) {
      throw StateError('missing $path');
    }
    return bytes;
  }

  group('sharedTrackImportAllowed', () {
    test('allows when shell is logged in', () {
      expect(
        sharedTrackImportAllowed(isLoggedIn: true, token: null),
        isTrue,
      );
    });

    test('allows JWT without nickname session (Drive cold start)', () {
      expect(
        sharedTrackImportAllowed(isLoggedIn: false, token: 'jwt'),
        isTrue,
      );
    });

    test('rejects when neither session nor token', () {
      expect(
        sharedTrackImportAllowed(isLoggedIn: false, token: null),
        isFalse,
      );
      expect(
        sharedTrackImportAllowed(isLoggedIn: false, token: ''),
        isFalse,
      );
    });
  });

  group('filenameFromSharedPath', () {
    test('takes last segment from path and file URI', () {
      expect(filenameFromSharedPath('/cache/ride.gpx'), 'ride.gpx');
      expect(
        filenameFromSharedPath('file:///data/user/0/app/cache/a.fit'),
        'a.fit',
      );
    });
  });

  group('receiveSharedTrackFiles', () {
    test('returns payload for gpx by extension', () async {
      final bytes = Uint8List.fromList(utf8.encode('<gpx></gpx>'));
      final result = await receiveSharedTrackFiles(
        [(path: '/tmp/ride.gpx', mimeType: 'application/octet-stream')],
        readBytes: (path) => readMap({'/tmp/ride.gpx': bytes}, path),
      );
      expect(result.payload?.filename, 'ride.gpx');
      expect(result.payload?.bytes, bytes);
      expect(result.readFailed, isFalse);
      expect(result.unsupportedFormat, isFalse);
    });

    test('sniffs Drive-style name without extension (octet-stream)', () async {
      final bytes = Uint8List.fromList(utf8.encode('<gpx version="1.1"></gpx>'));
      final result = await receiveSharedTrackFiles(
        [(path: '/cache/12345_document', mimeType: 'application/octet-stream')],
        readBytes: (path) => readMap({'/cache/12345_document': bytes}, path),
      );
      expect(result.payload?.filename, 'track.gpx');
      expect(result.unsupportedFormat, isFalse);
    });

    test('sniffs FIT header for broad MIME', () async {
      final bytes = Uint8List(14);
      bytes[0] = 14;
      bytes[8] = 0x2e;
      bytes[9] = 0x46;
      bytes[10] = 0x49;
      bytes[11] = 0x54;
      final result = await receiveSharedTrackFiles(
        [(path: '/cache/doc', mimeType: '*/*')],
        readBytes: (path) => readMap({'/cache/doc': bytes}, path),
      );
      expect(result.payload?.filename, 'track.fit');
    });

    test('marks unsupportedFormat for broad MIME non-track bytes', () async {
      final bytes = Uint8List.fromList(utf8.encode('not a track'));
      final result = await receiveSharedTrackFiles(
        [(path: '/cache/doc', mimeType: 'application/octet-stream')],
        readBytes: (path) => readMap({'/cache/doc': bytes}, path),
      );
      expect(result.payload, isNull);
      expect(result.readFailed, isFalse);
      expect(result.unsupportedFormat, isTrue);
    });

    test('marks readFailed when bytes cannot be read', () async {
      final result = await receiveSharedTrackFiles(
        [(path: '/missing.gpx', mimeType: null)],
        readBytes: (path) async => throw Exception('io'),
      );
      expect(result.payload, isNull);
      expect(result.readFailed, isTrue);
      expect(result.unsupportedFormat, isFalse);
    });

    test('skips non-candidate MIME without reading', () async {
      var reads = 0;
      final result = await receiveSharedTrackFiles(
        [(path: '/cache/photo', mimeType: 'image/png')],
        readBytes: (path) async {
          reads++;
          return Uint8List(0);
        },
      );
      expect(reads, 0);
      expect(result.payload, isNull);
      expect(result.unsupportedFormat, isFalse);
      expect(result.readFailed, isFalse);
    });

    test('uses precise MIME when path has no extension', () async {
      final bytes = Uint8List.fromList([1, 2, 3]);
      final result = await receiveSharedTrackFiles(
        [(path: '/cache/blob', mimeType: 'application/vnd.ant.fit')],
        readBytes: (path) => readMap({'/cache/blob': bytes}, path),
      );
      expect(result.payload?.filename, 'track.fit');
      expect(result.payload?.bytes, bytes);
    });

    test('skips junk then accepts later track file', () async {
      final gpx = Uint8List.fromList(utf8.encode('<gpx></gpx>'));
      final result = await receiveSharedTrackFiles(
        [
          (path: '/a.bin', mimeType: 'application/octet-stream'),
          (path: '/b.gpx', mimeType: 'application/gpx+xml'),
        ],
        readBytes: (path) async {
          if (path == '/a.bin') {
            return Uint8List.fromList(utf8.encode('nope'));
          }
          return gpx;
        },
      );
      expect(result.payload?.filename, 'b.gpx');
    });
  });
}
