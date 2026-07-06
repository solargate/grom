import 'dart:convert';
import 'dart:typed_data';

import '../models/recorded_track_point.dart';

class GpxTrackEncoder {
  Uint8List encode({
    required List<RecordedTrackPoint> points,
    String trackName = 'Recorded workout',
  }) {
    if (points.isEmpty) {
      throw ArgumentError('Cannot encode GPX without GPS points');
    }

    final buffer = StringBuffer()
      ..writeln('<?xml version="1.0" encoding="UTF-8"?>')
      ..writeln(
        '<gpx version="1.1" creator="Travka" '
        'xmlns="http://www.topografix.com/2009/GPX">',
      )
      ..writeln('  <trk>')
      ..writeln('    <name>${_escapeXml(trackName)}</name>')
      ..writeln('    <trkseg>');

    for (final point in points) {
      buffer.writeln(
        '      <trkpt lat="${_formatCoord(point.latitude)}" '
        'lon="${_formatCoord(point.longitude)}">',
      );
      buffer.writeln('        <time>${_formatTime(point.timestamp)}</time>');
      final altitude = point.altitude;
      if (altitude != null && !altitude.isNaN) {
        buffer.writeln('        <ele>${_formatCoord(altitude)}</ele>');
      }
      buffer.writeln('      </trkpt>');
    }

    buffer
      ..writeln('    </trkseg>')
      ..writeln('  </trk>')
      ..writeln('</gpx>');

    return Uint8List.fromList(utf8.encode(buffer.toString()));
  }

  String _formatCoord(double value) => value.toStringAsFixed(6);

  String _formatTime(DateTime time) {
    final utc = time.toUtc();
    final iso = utc.toIso8601String();
    return iso.endsWith('Z') ? iso : '${iso}Z';
  }

  String _escapeXml(String value) {
    return value
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&apos;');
  }
}
