import 'dart:math' as math;

import 'package:latlong2/latlong.dart';

const _earthRadiusMeters = 6371000.0;

double haversineMeters(LatLng a, LatLng b) {
  final lat1 = a.latitude * math.pi / 180;
  final lat2 = b.latitude * math.pi / 180;
  final dLat = (b.latitude - a.latitude) * math.pi / 180;
  final dLng = (b.longitude - a.longitude) * math.pi / 180;

  final sinDLat = math.sin(dLat / 2);
  final sinDLng = math.sin(dLng / 2);
  final h = sinDLat * sinDLat +
      math.cos(lat1) * math.cos(lat2) * sinDLng * sinDLng;
  return 2 * _earthRadiusMeters * math.asin(math.sqrt(h));
}

double pathDistanceMeters(List<LatLng> points) {
  if (points.length < 2) {
    return 0;
  }
  var total = 0.0;
  for (var i = 1; i < points.length; i++) {
    total += haversineMeters(points[i - 1], points[i]);
  }
  return total;
}

bool isValidGpsCoordinate(double lat, double lng, {double? accuracy}) {
  if (lat.isNaN || lng.isNaN) {
    return false;
  }
  if (lat < -90 || lat > 90 || lng < -180 || lng > 180) {
    return false;
  }
  if (lat == 0 && lng == 0) {
    return false;
  }
  if (accuracy != null && accuracy > 80) {
    return false;
  }
  return true;
}
