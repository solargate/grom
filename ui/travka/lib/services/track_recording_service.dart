import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';
import 'package:latlong2/latlong.dart';

import '../models/recorded_track.dart';
import '../models/recorded_track_point.dart';
import 'package:permission_handler/permission_handler.dart';

import '../utils/geo_utils.dart';
import 'gpx_track_encoder.dart';

enum TrackRecordingState { idle, recording, paused }

class TrackRecordingNotificationStrings {
  const TrackRecordingNotificationStrings({
    required this.title,
    required this.text,
  });

  final String title;
  final String text;
}

class TrackRecordingService extends ChangeNotifier {
  TrackRecordingService._();

  static final TrackRecordingService instance = TrackRecordingService._();

  final GpxTrackEncoder _gpxEncoder = GpxTrackEncoder();

  TrackRecordingState _state = TrackRecordingState.idle;
  final List<RecordedTrackPoint> _points = [];
  DateTime? _startTime;
  DateTime? _segmentStartedAt;
  Duration _accumulatedDuration = Duration.zero;
  double _currentSpeedKmh = 0;
  StreamSubscription<Position>? _positionSub;
  Timer? _tickTimer;
  TrackRecordingNotificationStrings? _notificationStrings;

  TrackRecordingState get state => _state;
  List<RecordedTrackPoint> get points => List.unmodifiable(_points);
  List<LatLng> get polylinePoints =>
      _points.map((p) => LatLng(p.latitude, p.longitude)).toList(growable: false);
  DateTime? get startTime => _startTime;
  int get durationSeconds {
    var total = _accumulatedDuration;
    if (_state == TrackRecordingState.recording && _segmentStartedAt != null) {
      total += DateTime.now().difference(_segmentStartedAt!);
    }
    return total.inSeconds;
  }
  double get distanceMeters => pathDistanceMeters(polylinePoints);
  double get currentSpeedKmh => _currentSpeedKmh;
  bool get isActive =>
      _state == TrackRecordingState.recording ||
      _state == TrackRecordingState.paused;

  Future<bool> ensureLocationReady() async {
    final serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      return false;
    }

    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
  if (permission == LocationPermission.denied ||
        permission == LocationPermission.deniedForever) {
      return false;
    }
    return true;
  }

  Future<bool> ensureBackgroundPermission() async {
    if (kIsWeb) {
      return true;
    }
    if (defaultTargetPlatform != TargetPlatform.android) {
      return true;
    }

    final status = await Permission.locationAlways.status;
    if (status.isGranted) {
      return true;
    }
    final result = await Permission.locationAlways.request();
    return result.isGranted;
  }

  Future<Position?> getCurrentPosition() async {
    if (!await ensureLocationReady()) {
      return null;
    }
    try {
      return await Geolocator.getCurrentPosition(
        locationSettings: _buildLocationSettings(forStream: false),
      );
    } catch (_) {
      return null;
    }
  }

  Future<void> startRecording({
    required TrackRecordingNotificationStrings notificationStrings,
  }) async {
    if (_state == TrackRecordingState.recording) {
      return;
    }

    if (!await ensureLocationReady()) {
      throw StateError('location_unavailable');
    }

    _notificationStrings = notificationStrings;
    final now = DateTime.now();

    if (_state == TrackRecordingState.idle) {
      _points.clear();
      _startTime = now;
      _accumulatedDuration = Duration.zero;
    }

    _state = TrackRecordingState.recording;
    _segmentStartedAt = now;
    _startTickTimer();
    await _startPositionStream();
    notifyListeners();
  }

  Future<void> pauseRecording() async {
    if (_state != TrackRecordingState.recording) {
      return;
    }
    _finalizeOpenSegment();
    _state = TrackRecordingState.paused;
    _stopTickTimer();
    await _stopPositionStream();
    notifyListeners();
  }

  Future<void> resumeRecording({
    required TrackRecordingNotificationStrings notificationStrings,
  }) async {
    if (_state != TrackRecordingState.paused) {
      return;
    }
    _notificationStrings = notificationStrings;
    _state = TrackRecordingState.recording;
    _segmentStartedAt = DateTime.now();
    _startTickTimer();
    await _startPositionStream();
    notifyListeners();
  }

  Future<RecordedTrack?> finishRecording() async {
    if (!isActive) {
      return null;
    }

    if (_state == TrackRecordingState.recording) {
      _finalizeOpenSegment();
    }

    await _stopPositionStream();
    _stopTickTimer();
    _state = TrackRecordingState.idle;

    if (_points.isEmpty || _startTime == null) {
      _resetSession();
      notifyListeners();
      return null;
    }

    final track = RecordedTrack(
      points: List.unmodifiable(_points),
      startTime: _startTime!,
      durationSeconds: durationSeconds,
      distanceMeters: distanceMeters,
      gpxBytes: _gpxEncoder.encode(points: _points),
    );

    _resetSession();
    notifyListeners();
    return track;
  }

  Future<void> discardRecording() async {
    await _stopPositionStream();
    _stopTickTimer();
    _resetSession();
    _state = TrackRecordingState.idle;
    notifyListeners();
  }

  void _resetSession() {
    _points.clear();
    _startTime = null;
    _segmentStartedAt = null;
    _accumulatedDuration = Duration.zero;
    _currentSpeedKmh = 0;
    _notificationStrings = null;
  }

  void _finalizeOpenSegment() {
    final segmentStart = _segmentStartedAt;
    if (segmentStart == null) {
      return;
    }
    _accumulatedDuration += DateTime.now().difference(segmentStart);
    _segmentStartedAt = null;
  }

  Future<void> _startPositionStream() async {
    await _stopPositionStream();
    _positionSub = Geolocator.getPositionStream(
      locationSettings: _buildLocationSettings(forStream: true),
    ).listen(_onPosition, onError: (_) {});
  }

  Future<void> _stopPositionStream() async {
    await _positionSub?.cancel();
    _positionSub = null;
  }

  void _startTickTimer() {
    _tickTimer?.cancel();
    _tickTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (_state == TrackRecordingState.recording) {
        notifyListeners();
      }
    });
  }

  void _stopTickTimer() {
    _tickTimer?.cancel();
    _tickTimer = null;
  }

  void _onPosition(Position position) {
    if (_state != TrackRecordingState.recording) {
      return;
    }

    if (!isValidGpsCoordinate(
      position.latitude,
      position.longitude,
      accuracy: position.accuracy,
    )) {
      return;
    }

    final point = RecordedTrackPoint(
      latitude: position.latitude,
      longitude: position.longitude,
      timestamp: position.timestamp,
      altitude: position.altitude,
      speedMps: position.speed >= 0 ? position.speed : null,
      heading: position.heading >= 0 ? position.heading : null,
      accuracy: position.accuracy,
    );

    _points.add(point);
    if (position.speed >= 0) {
      _currentSpeedKmh = position.speed * 3.6;
    }
    notifyListeners();
  }

  LocationSettings _buildLocationSettings({required bool forStream}) {
    if (defaultTargetPlatform == TargetPlatform.android) {
      final notification = _notificationStrings;
      return AndroidSettings(
        accuracy: LocationAccuracy.high,
        distanceFilter: 5,
        foregroundNotificationConfig: notification == null
            ? null
            : ForegroundNotificationConfig(
                notificationTitle: notification.title,
                notificationText: notification.text,
                enableWakeLock: true,
              ),
      );
    }

    if (defaultTargetPlatform == TargetPlatform.iOS ||
        defaultTargetPlatform == TargetPlatform.macOS) {
      return AppleSettings(
        accuracy: LocationAccuracy.high,
        activityType: ActivityType.fitness,
        distanceFilter: 5,
        pauseLocationUpdatesAutomatically: false,
        showBackgroundLocationIndicator: true,
      );
    }

    return const LocationSettings(
      accuracy: LocationAccuracy.high,
      distanceFilter: 5,
    );
  }
}
