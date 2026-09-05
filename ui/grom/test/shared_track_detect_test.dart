import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/platform/shared_track_detect.dart';

void main() {
  group('sharedTrackKindFromFilename', () {
    test('detects gpx and fit extensions case-insensitively', () {
      expect(sharedTrackKindFromFilename('Ride.GPX'), SharedTrackKind.gpx);
      expect(sharedTrackKindFromFilename('a.fit'), SharedTrackKind.fit);
      expect(sharedTrackKindFromFilename('notes.txt'), isNull);
    });
  });

  group('sharedTrackKindFromPreciseMime', () {
    test('maps known track MIME types', () {
      expect(
        sharedTrackKindFromPreciseMime('application/gpx+xml'),
        SharedTrackKind.gpx,
      );
      expect(
        sharedTrackKindFromPreciseMime('application/gpx'),
        SharedTrackKind.gpx,
      );
      expect(
        sharedTrackKindFromPreciseMime('application/x-gpx+xml'),
        SharedTrackKind.gpx,
      );
      expect(
        sharedTrackKindFromPreciseMime('application/vnd.ant.fit'),
        SharedTrackKind.fit,
      );
      expect(
        sharedTrackKindFromPreciseMime('application/fit'),
        SharedTrackKind.fit,
      );
      expect(
        sharedTrackKindFromPreciseMime('application/x-fit'),
        SharedTrackKind.fit,
      );
      expect(
        sharedTrackKindFromPreciseMime('application/octet-stream'),
        isNull,
      );
    });
  });

  group('isBroadSharedTrackMime', () {
    test('treats Drive-like types as broad', () {
      expect(isBroadSharedTrackMime(null), isTrue);
      expect(isBroadSharedTrackMime(''), isTrue);
      expect(isBroadSharedTrackMime('application/octet-stream'), isTrue);
      expect(isBroadSharedTrackMime('text/xml'), isTrue);
      expect(isBroadSharedTrackMime('application/xml'), isTrue);
      expect(isBroadSharedTrackMime('*/*'), isTrue);
      expect(isBroadSharedTrackMime('image/png'), isFalse);
    });
  });

  group('filenameForSharedTrackKind', () {
    test('returns canonical names', () {
      expect(filenameForSharedTrackKind(SharedTrackKind.gpx), 'track.gpx');
      expect(filenameForSharedTrackKind(SharedTrackKind.fit), 'track.fit');
    });
  });

  group('sharedTrackKindFromBytes', () {
    test('detects GPX tag after XML declaration', () {
      final bytes = Uint8List.fromList(
        utf8.encode('<?xml version="1.0"?><gpx version="1.1"></gpx>'),
      );
      expect(sharedTrackKindFromBytes(bytes), SharedTrackKind.gpx);
    });

    test('detects FIT header signature', () {
      final bytes = Uint8List(14);
      bytes[0] = 14;
      bytes[8] = 0x2e;
      bytes[9] = 0x46;
      bytes[10] = 0x49;
      bytes[11] = 0x54;
      expect(sharedTrackKindFromBytes(bytes), SharedTrackKind.fit);
    });

    test('rejects unrelated bytes', () {
      final bytes = Uint8List.fromList(utf8.encode('hello world'));
      expect(sharedTrackKindFromBytes(bytes), isNull);
    });
  });

  group('resolveSharedTrackKind', () {
    test('prefers filename over MIME', () {
      expect(
        resolveSharedTrackKind(
          filename: 'ride.gpx',
          mimeType: 'application/vnd.ant.fit',
        ),
        SharedTrackKind.gpx,
      );
    });

    test('uses precise MIME when name has no extension', () {
      expect(
        resolveSharedTrackKind(
          filename: 'document',
          mimeType: 'application/gpx+xml',
        ),
        SharedTrackKind.gpx,
      );
    });

    test('sniffs bytes only for broad MIME', () {
      final gpx = Uint8List.fromList(utf8.encode('<gpx></gpx>'));
      expect(
        resolveSharedTrackKind(
          filename: 'document',
          mimeType: 'application/octet-stream',
          bytes: gpx,
        ),
        SharedTrackKind.gpx,
      );
      expect(
        resolveSharedTrackKind(
          filename: 'document',
          mimeType: 'image/png',
          bytes: gpx,
        ),
        isNull,
      );
    });

    test('sniffs text/xml as broad MIME', () {
      final gpx = Uint8List.fromList(utf8.encode('<gpx></gpx>'));
      expect(
        resolveSharedTrackKind(
          filename: 'document',
          mimeType: 'text/xml',
          bytes: gpx,
        ),
        SharedTrackKind.gpx,
      );
    });
  });
}
