import 'package:flutter_test/flutter_test.dart';
import 'package:grom/utils/geo_utils.dart';
import 'package:latlong2/latlong.dart';

void main() {
  test('pathDistanceMeters returns zero for paths with fewer than two points',
      () {
    expect(pathDistanceMeters([]), 0);
    expect(pathDistanceMeters([const LatLng(55.75, 37.62)]), 0);
  });

  test('haversineMeters calculates a known equatorial distance', () {
    final distance = haversineMeters(
      const LatLng(0, 1),
      const LatLng(0, 2),
    );

    expect(distance, closeTo(111195, 1));
  });

  test('pathDistanceMeters adds every segment', () {
    final firstSegment = haversineMeters(
      const LatLng(55.75, 37.62),
      const LatLng(55.751, 37.62),
    );
    final secondSegment = haversineMeters(
      const LatLng(55.751, 37.62),
      const LatLng(55.751, 37.621),
    );

    expect(
      pathDistanceMeters(const [
        LatLng(55.75, 37.62),
        LatLng(55.751, 37.62),
        LatLng(55.751, 37.621),
      ]),
      closeTo(firstSegment + secondSegment, 0.0001),
    );
  });

  test('isValidGpsCoordinate rejects unusable GPS fixes', () {
    expect(isValidGpsCoordinate(55.75, 37.62, accuracy: 10), isTrue);
    expect(isValidGpsCoordinate(0, 0), isFalse);
    expect(isValidGpsCoordinate(91, 0), isFalse);
    expect(isValidGpsCoordinate(55.75, 37.62, accuracy: 81), isFalse);
  });
}
