import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/movement_detector.dart';
import 'package:latlong2/latlong.dart';

void main() {
  test('MovementDetector triggers pause after 3 seconds without movement', () {
    final detector = MovementDetector();
    final start = DateTime(2026, 3, 6, 12, 0, 0);

    detector.onPosition(const LatLng(55.75, 37.62), start);
    detector.onPosition(
      const LatLng(55.75001, 37.62001),
      start.add(const Duration(seconds: 1)),
    );

    expect(
      detector.isStationaryForPause(start.add(const Duration(seconds: 1))),
      isFalse,
    );
    expect(
      detector.isStationaryForPause(start.add(const Duration(seconds: 4))),
      isTrue,
    );
  });

  test('MovementDetector triggers resume after 2 meters within 3 seconds', () {
    final detector = MovementDetector();
    final start = DateTime(2026, 3, 6, 12, 0, 0);
    final sampleTime = start.add(const Duration(seconds: 2));

    detector.onPosition(const LatLng(55.75, 37.62), start);
    detector.onPosition(const LatLng(55.752, 37.62), sampleTime);

    expect(detector.hasRecentMovementForResume(sampleTime), isTrue);
  });

  test('MovementDetector ignores small GPS jitter below threshold', () {
    final detector = MovementDetector();
    final start = DateTime(2026, 3, 6, 12, 0, 0);
    final sampleTime = start.add(const Duration(seconds: 2));

    detector.onPosition(const LatLng(55.75, 37.62), start);
    detector.onPosition(
      const LatLng(55.750005, 37.620005),
      start.add(const Duration(seconds: 1)),
    );
    detector.onPosition(
      const LatLng(55.750008, 37.620008),
      sampleTime,
    );

    expect(detector.hasRecentMovementForResume(sampleTime), isFalse);
  });
}
