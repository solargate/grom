import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/device_track_import_id.dart';
import 'package:grom/services/device_track_import_service.dart';

void main() {
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

  test('TrackPickResult typedef shape', () {
    final pick = (
      filename: 'a.gpx',
      bytes: Uint8List.fromList([1, 2, 3]),
    );
    expect(pick.filename, 'a.gpx');
    expect(pick.bytes.length, 3);
  });
}
