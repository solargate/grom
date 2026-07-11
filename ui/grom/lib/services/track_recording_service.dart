import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';
import 'package:latlong2/latlong.dart';
import 'package:permission_handler/permission_handler.dart';

import '../models/recorded_track.dart';
import '../models/recorded_track_point.dart';
import '../utils/geo_utils.dart';
import 'gpx_track_encoder.dart';
import 'track_recording_foreground.dart';
import 'track_recording_session.dart';
import 'track_recording_state.dart';
import 'track_recording_store.dart';

export 'track_recording_state.dart';

class TrackRecordingNotificationStrings {
  const TrackRecordingNotificationStrings({
    required this.title,
    required this.text,
    required this.channelName,
    this.pausedText,
  });

  final String title;
  final String text;
  final String channelName;
  final String? pausedText;
}

class TrackRecordingService extends ChangeNotifier {
  TrackRecordingService._();

  static final TrackRecordingService instance = TrackRecordingService._();

  final GpxTrackEncoder _gpxEncoder = GpxTrackEncoder();
  final TrackRecordingStore _store = TrackRecordingStore.instance;

  TrackRecordingState _state = TrackRecordingState.idle;
  final List<RecordedTrackPoint> _points = [];
  DateTime? _startTime;
  DateTime? _segmentStartedAt;
  Duration _accumulatedDuration = Duration.zero;
  double _currentSpeedKmh = 0;
  StreamSubscription<Position>? _positionSub;
  Timer? _tickTimer;
  TrackRecordingNotificationStrings? _notificationStrings;
  String? _streamError;
  bool _initialized = false;

  bool get _useForegroundTask => TrackRecordingForeground.isSupported;

  TrackRecordingState get state => _state;
  List<RecordedTrackPoint> get points => List.unmodifiable(_points);
  List<LatLng> get polylinePoints =>
      _points.map((p) => LatLng(p.latitude, p.longitude)).toList(growable: false);
  DateTime? get startTime => _startTime;
  String? get streamError => _streamError;
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

  Future<void> initialize() async {
    if (_initialized) {
      return;
    }
    _initialized = true;

    await _store.init();
    await TrackRecordingForeground.initialize();
    TrackRecordingForeground.registerTaskDataCallback(_onTaskData);

    final session = await _store.load();
    if (session == null || !session.isActive) {
      return;
    }

    if (await TrackRecordingForeground.isRunning()) {
      _applySession(session);
      if (_state == TrackRecordingState.recording) {
        _startTickTimer();
      }
      notifyListeners();
    }
  }

  Future<bool> needsRecovery() async {
    await initialize();
    final session = await _store.load();
    if (session == null || !session.isActive) {
      return false;
    }
    if (await TrackRecordingForeground.isRunning()) {
      return false;
    }
    _applySession(session);
    notifyListeners();
    return true;
  }

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

    if (await Geolocator.checkPermission() == LocationPermission.always) {
      return true;
    }

    final status = await Permission.locationAlways.status;
    if (status.isGranted) {
      return true;
    }
    final result = await Permission.locationAlways.request();
    return result.isGranted &&
        await Geolocator.checkPermission() == LocationPermission.always;
  }

  Future<bool> ensureNotificationPermission() async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) {
      return true;
    }

    if (await TrackRecordingForeground.ensureNotificationPermission()) {
      return true;
    }

    final status = await Permission.notification.status;
    if (status.isGranted) {
      return true;
    }
    final result = await Permission.notification.request();
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
    if (!await ensureBackgroundPermission()) {
      throw StateError('background_permission_denied');
    }
    if (!await ensureNotificationPermission()) {
      throw StateError('notification_permission_denied');
    }

    _notificationStrings = notificationStrings;
    _streamError = null;
    final now = DateTime.now();

    if (_state == TrackRecordingState.idle) {
      _points.clear();
      _startTime = now;
      _accumulatedDuration = Duration.zero;
    }

    _state = TrackRecordingState.recording;
    _segmentStartedAt = now;
    await _persistSession();
    _startTickTimer();

    if (_useForegroundTask) {
      await _startForegroundRecording(notificationStrings);
    } else {
      await _startPositionStream();
    }

    notifyListeners();
  }

  Future<void> pauseRecording() async {
    if (_state != TrackRecordingState.recording) {
      return;
    }
    _finalizeOpenSegment();
    _state = TrackRecordingState.paused;
    _currentSpeedKmh = 0;
    await _persistSession();
    _stopTickTimer();

    if (_useForegroundTask) {
      await TrackRecordingForeground.sendCommand('pause');
      final pausedText = _notificationStrings?.pausedText;
      if (pausedText != null && _notificationStrings != null) {
        await TrackRecordingForeground.updateNotification(
          title: _notificationStrings!.title,
          text: pausedText,
        );
      }
    } else {
      await _stopPositionStream();
    }

    notifyListeners();
  }

  Future<void> resumeRecording({
    required TrackRecordingNotificationStrings notificationStrings,
  }) async {
    if (_state != TrackRecordingState.paused) {
      return;
    }
    if (!await ensureBackgroundPermission()) {
      throw StateError('background_permission_denied');
    }
    if (!await ensureNotificationPermission()) {
      throw StateError('notification_permission_denied');
    }

    _notificationStrings = notificationStrings;
    _streamError = null;
    _state = TrackRecordingState.recording;
    _segmentStartedAt = DateTime.now();
    await _persistSession();
    _startTickTimer();

    if (_useForegroundTask) {
      if (!await TrackRecordingForeground.isRunning()) {
        await _startForegroundRecording(notificationStrings);
      } else {
        await TrackRecordingForeground.updateNotification(
          title: notificationStrings.title,
          text: notificationStrings.text,
        );
        await TrackRecordingForeground.sendCommand('resume');
      }
    } else {
      await _startPositionStream();
    }

    notifyListeners();
  }

  Future<void> restoreInterruptedRecording({
    required TrackRecordingNotificationStrings notificationStrings,
  }) async {
    if (!isActive) {
      return;
    }
    _notificationStrings = notificationStrings;
    _streamError = null;

    if (_state == TrackRecordingState.recording) {
      _startTickTimer();
      await _startForegroundRecording(notificationStrings);
    } else {
      await TrackRecordingForeground.start(
        title: notificationStrings.title,
        text: notificationStrings.pausedText ?? notificationStrings.text,
        channelName: notificationStrings.channelName,
        pausedText: notificationStrings.pausedText,
      );
    }

    notifyListeners();
  }

  Future<RecordedTrack?> finishRecording() async {
    if (!isActive) {
      return null;
    }

    if (_state == TrackRecordingState.recording) {
      _finalizeOpenSegment();
    }

    await _stopRecordingInfrastructure();
    _stopTickTimer();
    _state = TrackRecordingState.idle;

    if (_points.isEmpty || _startTime == null) {
      await _resetSession();
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

    await _resetSession();
    notifyListeners();
    return track;
  }

  Future<void> discardRecording() async {
    await _stopRecordingInfrastructure();
    _stopTickTimer();
    await _resetSession();
    _state = TrackRecordingState.idle;
    notifyListeners();
  }

  Future<void> _startForegroundRecording(
    TrackRecordingNotificationStrings notificationStrings,
  ) async {
    final result = await TrackRecordingForeground.start(
      title: notificationStrings.title,
      text: notificationStrings.text,
      channelName: notificationStrings.channelName,
      pausedText: notificationStrings.pausedText,
    );
    if (result is ForegroundServiceFailure) {
      _streamError = result.error.toString();
      throw StateError('foreground_service_start_failed');
    }
    if (_state == TrackRecordingState.recording) {
      await TrackRecordingForeground.sendCommand('resume');
    }
  }

  Future<void> _stopRecordingInfrastructure() async {
    if (_useForegroundTask) {
      await TrackRecordingForeground.sendCommand('stop');
      await TrackRecordingForeground.stop();
    } else {
      await _stopPositionStream();
    }
  }

  Future<void> _resetSession() async {
    _points.clear();
    _startTime = null;
    _segmentStartedAt = null;
    _accumulatedDuration = Duration.zero;
    _currentSpeedKmh = 0;
    _notificationStrings = null;
    _streamError = null;
    await _store.clear();
  }

  void _finalizeOpenSegment() {
    final segmentStart = _segmentStartedAt;
    if (segmentStart == null) {
      return;
    }
    _accumulatedDuration += DateTime.now().difference(segmentStart);
    _segmentStartedAt = null;
  }

  Future<void> _persistSession() async {
    final startTime = _startTime;
    if (startTime == null || !isActive) {
      return;
    }
    await _store.save(
      TrackRecordingSession(
        state: _state,
        startTime: startTime,
        accumulatedDurationMs: _accumulatedDuration.inMilliseconds,
        segmentStartedAt: _segmentStartedAt,
        points: List<RecordedTrackPoint>.from(_points),
      ),
    );
  }

  void _applySession(TrackRecordingSession session) {
    _state = session.state;
    _startTime = session.startTime;
    _accumulatedDuration = Duration(milliseconds: session.accumulatedDurationMs);
    _segmentStartedAt = session.segmentStartedAt;
    _points
      ..clear()
      ..addAll(session.points);
    if (_points.isNotEmpty) {
      final lastPoint = _points.last;
      if (lastPoint.speedMps != null && lastPoint.speedMps! >= 0) {
        _currentSpeedKmh = lastPoint.speedMps! * 3.6;
      }
    }
  }

  void _onTaskData(Object data) {
    if (data is! Map) {
      return;
    }

    final type = data['type'];
    if (type == 'error') {
      _streamError = data['message']?.toString();
      notifyListeners();
      return;
    }

    if (type == 'point') {
      final pointJson = data['point'];
      if (pointJson is! Map) {
        return;
      }
      final point = RecordedTrackPoint.fromJson(
        Map<String, dynamic>.from(pointJson),
      );
      if (_points.isEmpty ||
          _points.last.timestamp != point.timestamp ||
          _points.last.latitude != point.latitude ||
          _points.last.longitude != point.longitude) {
        _points.add(point);
      }
      final speedKmh = data['speedKmh'];
      if (speedKmh is num) {
        _currentSpeedKmh = speedKmh.toDouble();
      }
      notifyListeners();
    }
  }

  Future<void> _startPositionStream() async {
    await _stopPositionStream();
    _positionSub = Geolocator.getPositionStream(
      locationSettings: _buildLocationSettings(forStream: true),
    ).listen(
      _onPosition,
      onError: (Object error) {
        _streamError = error.toString();
        notifyListeners();
      },
    );
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

  Future<void> _onPosition(Position position) async {
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
    await _persistSession();
    notifyListeners();
  }

  LocationSettings _buildLocationSettings({required bool forStream}) {
    if (defaultTargetPlatform == TargetPlatform.android) {
      final notification = _notificationStrings;
      return AndroidSettings(
        accuracy: LocationAccuracy.high,
        distanceFilter: 5,
        intervalDuration: const Duration(seconds: 5),
        foregroundNotificationConfig: notification == null
            ? null
            : ForegroundNotificationConfig(
                notificationTitle: notification.title,
                notificationText: notification.text,
                notificationChannelName: notification.channelName,
                enableWakeLock: true,
                setOngoing: true,
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
