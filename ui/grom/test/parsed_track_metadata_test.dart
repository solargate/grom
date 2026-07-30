import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/parsed_track_metadata.dart';

void main() {
  test('ParsedTrackMetadata.fromJson reads optional name', () {
    final meta = ParsedTrackMetadata.fromJson({
      'name': '  Test track  ',
      'has_gps': true,
      'distance': 1200,
    });
    expect(meta.name, 'Test track');
    expect(meta.hasGps, isTrue);
    expect(meta.distanceMeters, 1200);
  });

  test('ParsedTrackMetadata.fromJson treats blank name as absent', () {
    final meta = ParsedTrackMetadata.fromJson({
      'name': '   ',
      'has_gps': false,
    });
    expect(meta.name, isNull);
  });
}
