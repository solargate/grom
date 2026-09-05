import 'dart:convert';

/// One aligned sample from Strava activity streams.
class StravaStreamPoint {
  const StravaStreamPoint({
    required this.lat,
    required this.lon,
    this.timeSeconds,
    this.elevation,
  });

  final double lat;
  final double lon;
  final int? timeSeconds;
  final double? elevation;
}

/// Builds a minimal GPX 1.1 document from Strava latlng (+ optional time/ele).
List<int> buildGpxFromStravaStreams({
  required String name,
  required DateTime startDate,
  required List<StravaStreamPoint> points,
}) {
  if (points.length < 2) {
    throw ArgumentError('Need at least 2 GPS points for a track');
  }

  final buffer = StringBuffer();
  buffer.writeln('<?xml version="1.0" encoding="UTF-8"?>');
  buffer.writeln(
    '<gpx version="1.1" creator="grom-strava-api" '
    'xmlns="http://www.topografix.com/GPX/1/1">',
  );
  buffer.writeln('<trk>');
  buffer.writeln('<name>${_escapeXml(name)}</name>');
  buffer.writeln('<trkseg>');
  for (final point in points) {
    buffer.write('<trkpt lat="${point.lat}" lon="${point.lon}">');
    if (point.elevation != null) {
      buffer.write('<ele>${point.elevation}</ele>');
    }
    final when = point.timeSeconds != null
        ? startDate.toUtc().add(Duration(seconds: point.timeSeconds!))
        : startDate.toUtc();
    buffer.write('<time>${when.toIso8601String()}</time>');
    buffer.writeln('</trkpt>');
  }
  buffer.writeln('</trkseg>');
  buffer.writeln('</trk>');
  buffer.writeln('</gpx>');
  return utf8.encode(buffer.toString());
}

String _escapeXml(String value) {
  return value
      .replaceAll('&', '&amp;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
      .replaceAll('"', '&quot;')
      .replaceAll("'", '&apos;');
}
