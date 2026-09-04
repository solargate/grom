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

  test('ParsedTrackMetadata.fromJson reads optional sport_type', () {
    final meta = ParsedTrackMetadata.fromJson({
      'sport_type': '  Ride  ',
      'has_gps': true,
    });
    expect(meta.sportType, 'Ride');
  });

  test('ParsedTrackMetadata.fromJson treats blank sport_type as absent', () {
    final meta = ParsedTrackMetadata.fromJson({
      'sport_type': '   ',
      'has_gps': false,
    });
    expect(meta.sportType, isNull);
  });

  test('ParsedTrackMetadata.fromJson omits sport_type when missing', () {
    final meta = ParsedTrackMetadata.fromJson({
      'has_gps': true,
    });
    expect(meta.sportType, isNull);
  });
}
