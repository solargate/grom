import 'dart:convert';
import 'dart:typed_data';

import 'package:fit_sdk/fit_sdk.dart';
import 'package:gpx/gpx.dart';
import 'package:latlong2/latlong.dart';

const _maxRenderPoints = 500;
const _semicirclesToDegrees = 11930464.7111;

class TrackParseException implements Exception {
  TrackParseException(this.message);

  final String message;

  @override
  String toString() => message;
}

List<LatLng> parseTrackPoints(List<int> bytes, String filename) {
  final lower = filename.toLowerCase();
  List<LatLng> points;
  if (lower.endsWith('.gpx')) {
    points = _parseGpx(bytes);
  } else if (lower.endsWith('.fit')) {
    points = _parseFit(bytes);
  } else {
    throw TrackParseException('Unsupported track format');
  }

  if (points.length < 2) {
    throw TrackParseException('Track has no GPS points');
  }
  return simplifyForRender(points);
}

List<LatLng> _parseGpx(List<int> bytes) {
  final gpx = GpxReader().fromString(utf8.decode(bytes));
  final points = <LatLng>[];
  for (final track in gpx.trks) {
    for (final segment in track.trksegs) {
      for (final pt in segment.trkpts) {
        final lat = pt.lat;
        final lon = pt.lon;
        if (lat != null && lon != null && _validCoord(lat, lon)) {
          points.add(LatLng(lat, lon));
        }
      }
    }
  }
  return points;
}

List<LatLng> _parseFit(List<int> bytes) {
  final points = <LatLng>[];
  final decoder = Decode();
  decoder.onMesg = (Mesg mesg) {
    if (mesg.num != MesgNum.record) {
      return;
    }

    double? lat;
    double? lon;
    for (final field in mesg.fields) {
      if (field.isInvalid || field.value == null) {
        continue;
      }
      if (field.num == 0 && field.value is num) {
        lat = (field.value as num).toDouble() / _semicirclesToDegrees;
      } else if (field.num == 1 && field.value is num) {
        lon = (field.value as num).toDouble() / _semicirclesToDegrees;
      }
    }

    lat ??= _semicircleFieldValue(mesg.getFieldValue(0));
    lon ??= _semicircleFieldValue(mesg.getFieldValue(1));

    if (lat != null && lon != null && _validCoord(lat, lon)) {
      points.add(LatLng(lat, lon));
    }
  };
  decoder.read(Uint8List.fromList(bytes));
  return points;
}

bool _validCoord(double lat, double lon) {
  if (lat.isNaN || lon.isNaN) {
    return false;
  }
  if (lat < -90 || lat > 90 || lon < -180 || lon > 180) {
    return false;
  }
  if (lat == 0 && lon == 0) {
    return false;
  }
  return true;
}

double? _semicircleFieldValue(Object? value) {
  if (value is! num) {
    return null;
  }
  return value.toDouble() / _semicirclesToDegrees;
}

List<LatLng> simplifyForRender(List<LatLng> points) {
  if (points.length <= _maxRenderPoints) {
    return points;
  }
  final step = (points.length - 1) / (_maxRenderPoints - 1);
  final simplified = <LatLng>[];
  for (var i = 0; i < _maxRenderPoints; i++) {
    var idx = (i * step).round();
    if (idx >= points.length) {
      idx = points.length - 1;
    }
    simplified.add(points[idx]);
  }
  return simplified;
}
