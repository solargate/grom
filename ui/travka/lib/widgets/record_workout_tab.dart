import 'package:flutter/material.dart';
import 'package:latlong2/latlong.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:travka/l10n/app_localizations.dart';
import 'package:travka/l10n/sport_type_localizations.dart';

import '../models/recorded_track.dart';
import '../services/track_recording_service.dart';
import 'workout_record_controls.dart';
import 'workout_record_map.dart';

class RecordWorkoutTab extends StatefulWidget {
  const RecordWorkoutTab({
    super.key,
    required this.onFinished,
  });

  final ValueChanged<RecordedTrack> onFinished;

  @override
  State<RecordWorkoutTab> createState() => _RecordWorkoutTabState();
}

class _RecordWorkoutTabState extends State<RecordWorkoutTab> {
  final TrackRecordingService _recorder = TrackRecordingService.instance;
  LatLng? _initialCenter;
  bool _loadingLocation = true;

  @override
  void initState() {
    super.initState();
    _recorder.addListener(_onRecorderChanged);
    _initLocation();
  }

  @override
  void dispose() {
    _recorder.removeListener(_onRecorderChanged);
    super.dispose();
  }

  Future<void> _initLocation() async {
    final ready = await _recorder.ensureLocationReady();
    if (!mounted) {
      return;
    }
    if (!ready) {
      setState(() => _loadingLocation = false);
      _showLocationError();
      return;
    }

    final position = await _recorder.getCurrentPosition();
    if (!mounted) {
      return;
    }
    setState(() {
      _loadingLocation = false;
      if (position != null) {
        _initialCenter = LatLng(position.latitude, position.longitude);
      }
    });
  }

  void _onRecorderChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  TrackRecordingNotificationStrings _notificationStrings(
    AppLocalizations l10n,
  ) {
    return TrackRecordingNotificationStrings(
      title: l10n.recordingNotificationTitle,
      text: l10n.recordingNotificationText,
    );
  }

  Future<void> _onPlay() async {
    final l10n = AppLocalizations.of(context)!;
    final state = _recorder.state;

    if (state == TrackRecordingState.idle) {
      final proceed = await _confirmBackgroundAccess(l10n);
      if (!proceed || !mounted) {
        return;
      }
    }

    try {
      if (state == TrackRecordingState.paused) {
        await _recorder.resumeRecording(
          notificationStrings: _notificationStrings(l10n),
        );
      } else {
        await _recorder.startRecording(
          notificationStrings: _notificationStrings(l10n),
        );
      }
    } catch (_) {
      if (!mounted) return;
      _showLocationError();
    }
  }

  Future<void> _onPause() async {
    await _recorder.pauseRecording();
  }

  Future<void> _onFinish() async {
    final track = await _recorder.finishRecording();
    if (track != null && mounted) {
      widget.onFinished(track);
    }
  }

  Future<bool> _confirmBackgroundAccess(AppLocalizations l10n) async {
    final granted = await _recorder.ensureBackgroundPermission();
    if (granted) {
      return true;
    }
    if (!mounted) {
      return false;
    }

    final openSettings = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.backgroundLocationRationale),
        content: Text(l10n.backgroundLocationRationale),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.openSettings),
          ),
        ],
      ),
    );

    if (openSettings == true) {
      await openAppSettings();
    }
    return _recorder.ensureBackgroundPermission();
  }

  void _showLocationError() {
    final l10n = AppLocalizations.of(context)!;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.locationPermissionDenied)),
    );
  }

  String _formatSpeed(AppLocalizations l10n) {
    if (_recorder.state != TrackRecordingState.recording ||
        _recorder.currentSpeedKmh <= 0) {
      return l10n.speedUnavailable;
    }
    return l10n.speedKmh(_recorder.currentSpeedKmh.toStringAsFixed(1));
  }

  String _formatDistance(AppLocalizations l10n) {
    final distanceKm = _recorder.distanceMeters / 1000;
    if (distanceKm == 0) {
      return l10n.distanceZero;
    }
    return l10n.distanceKilometers(
      distanceKm >= 10
          ? distanceKm.toStringAsFixed(1)
          : distanceKm.toStringAsFixed(2),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final mapHeight = MediaQuery.sizeOf(context).height * 0.42;
    final trackPoints = _recorder.polylinePoints;
    final followUser = _recorder.state == TrackRecordingState.recording;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        SizedBox(
          height: mapHeight.clamp(220, 420),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: _loadingLocation
                ? const Center(child: CircularProgressIndicator())
                : WorkoutRecordMap(
                    trackPoints: trackPoints,
                    followUser: followUser,
                    initialCenter: _initialCenter,
                  ),
          ),
        ),
        const SizedBox(height: 12),
        Text(
          l10n.doNotDismissNotification,
          style: Theme.of(context).textTheme.bodySmall,
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 16),
        WorkoutRecordControls(
          state: _recorder.state,
          onPlay: _onPlay,
          onPause: _onPause,
          onFinish: _onFinish,
          playLabel: l10n.recordStart,
          pauseLabel: l10n.recordPause,
          finishLabel: l10n.recordFinish,
        ),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            Column(
              children: [
                Text(
                  l10n.recordingDuration,
                  style: Theme.of(context).textTheme.labelMedium,
                ),
                Text(
                  formatDuration(l10n, _recorder.durationSeconds),
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ],
            ),
            Column(
              children: [
                Text(
                  l10n.currentSpeed,
                  style: Theme.of(context).textTheme.labelMedium,
                ),
                Text(
                  _formatSpeed(l10n),
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ],
            ),
            Column(
              children: [
                Text(
                  l10n.workoutDistance,
                  style: Theme.of(context).textTheme.labelMedium,
                ),
                Text(
                  _formatDistance(l10n),
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }
}
