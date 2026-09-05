import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_gpx.dart';

void main() {
  test('buildGpxFromStravaStreams writes trkpt with lat lon time', () {
    final start = DateTime.utc(2026, 9, 5, 10, 0, 0);
    final bytes = buildGpxFromStravaStreams(
      name: 'Morning <Run>',
      startDate: start,
      points: const [
        StravaStreamPoint(lat: 55.75, lon: 37.61, timeSeconds: 0, elevation: 140),
        StravaStreamPoint(lat: 55.76, lon: 37.62, timeSeconds: 60, elevation: 142),
      ],
    );
    final xml = utf8.decode(bytes);
    expect(xml, contains('<name>Morning &lt;Run&gt;</name>'));
    expect(xml, contains('lat="55.75"'));
    expect(xml, contains('lon="37.61"'));
    expect(xml, contains('<ele>140.0</ele>'));
    expect(xml, contains('<time>2026-09-05T10:00:00.000Z</time>'));
    expect(xml, contains('<time>2026-09-05T10:01:00.000Z</time>'));
  });

  test('buildGpxFromStravaStreams rejects short tracks', () {
    expect(
      () => buildGpxFromStravaStreams(
        name: 'x',
        startDate: DateTime.utc(2026, 1, 1),
        points: const [StravaStreamPoint(lat: 1, lon: 2)],
      ),
      throwsArgumentError,
    );
  });
}
