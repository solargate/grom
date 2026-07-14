import 'package:latlong2/latlong.dart';

import '../utils/geo_utils.dart';

const kAutoPauseDistanceMeters = 2.0;
const kAutoPauseTimeWindow = Duration(seconds: 5);

class _PositionSample {
  const _PositionSample(this.position, this.time);

  final LatLng position;
  final DateTime time;
}

class MovementDetector {
  LatLng? _anchor;
  DateTime? _lastMovementAt;
  final List<_PositionSample> _recentPositions = [];

  void reset() {
    _anchor = null;
    _lastMovementAt = null;
    _recentPositions.clear();
  }

  void onPosition(LatLng position, DateTime time) {
    if (_anchor == null) {
      _anchor = position;
      _lastMovementAt = time;
      _addRecent(position, time);
      return;
    }

    if (haversineMeters(_anchor!, position) >= kAutoPauseDistanceMeters) {
      _anchor = position;
      _lastMovementAt = time;
    }

    _addRecent(position, time);
  }

  void _addRecent(LatLng position, DateTime time) {
    _recentPositions.add(_PositionSample(position, time));
    _trimRecent(time);
  }

  void _trimRecent(DateTime now) {
    _recentPositions.removeWhere(
      (entry) => now.difference(entry.time) > kAutoPauseTimeWindow,
    );
  }

  bool isStationaryForPause([DateTime? referenceTime]) {
    final lastMovement = _lastMovementAt;
    if (lastMovement == null) {
      return false;
    }
    final now = referenceTime ?? DateTime.now();
    return now.difference(lastMovement) >= kAutoPauseTimeWindow;
  }

  bool hasRecentMovementForResume([DateTime? referenceTime]) {
    final now = referenceTime ?? DateTime.now();
    _trimRecent(now);
    if (_recentPositions.length < 2) {
      return false;
    }
    final oldest = _recentPositions.first;
    final newest = _recentPositions.last;
    return haversineMeters(oldest.position, newest.position) >=
        kAutoPauseDistanceMeters;
  }
}
