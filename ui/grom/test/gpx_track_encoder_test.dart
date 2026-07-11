import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/recorded_track_point.dart';
import 'package:grom/services/gpx_track_encoder.dart';

void main() {
  test('GpxTrackEncoder produces valid GPX XML', () {
    final encoder = GpxTrackEncoder();
    final start = DateTime.utc(2024, 6, 1, 10, 0, 0);
    final points = [
      RecordedTrackPoint(
        latitude: 55.7558,
        longitude: 37.6173,
        timestamp: start,
        altitude: 150.5,
        speedMps: 2.5,
      ),
      RecordedTrackPoint(
        latitude: 55.7560,
        longitude: 37.6180,
        timestamp: start.add(const Duration(seconds: 10)),
        speedMps: 2.8,
      ),
    ];

    final bytes = encoder.encode(points: points);
    final xml = utf8.decode(bytes);

    expect(bytes, isNotEmpty);
    expect(xml, contains('<?xml version="1.0"'));
    expect(xml, contains('<gpx version="1.1"'));
    expect(xml, contains('<trkpt lat="55.755800" lon="37.617300">'));
    expect(xml, contains('<time>2024-06-01T10:00:00.000Z</time>'));
    expect(xml, contains('<ele>150.500000</ele>'));
    expect(xml, contains('<trkpt lat="55.756000" lon="37.618000">'));
    expect(xml, contains('<time>2024-06-01T10:00:10.000Z</time>'));
  });

  test('GpxTrackEncoder rejects empty points', () {
    final encoder = GpxTrackEncoder();
    expect(
      () => encoder.encode(points: []),
      throwsArgumentError,
    );
  });
}
