import 'package:flutter/material.dart';

import '../services/track_recording_service.dart';

class WorkoutRecordControls extends StatelessWidget {
  const WorkoutRecordControls({
    super.key,
    required this.state,
    required this.onPlay,
    required this.onPause,
    required this.onFinish,
    required this.playLabel,
    required this.pauseLabel,
    required this.finishLabel,
  });

  final TrackRecordingState state;
  final VoidCallback onPlay;
  final VoidCallback onPause;
  final VoidCallback onFinish;
  final String playLabel;
  final String pauseLabel;
  final String finishLabel;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final isRecording = state == TrackRecordingState.recording;
    final showFinish = state == TrackRecordingState.paused;

    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        if (showFinish) ...[
          OutlinedButton(
            onPressed: onFinish,
            child: Text(finishLabel),
          ),
          const SizedBox(width: 16),
        ],
        Semantics(
          button: true,
          label: isRecording ? pauseLabel : playLabel,
          child: Material(
            color: colorScheme.primary,
            shape: const CircleBorder(),
            elevation: 2,
            child: InkWell(
              customBorder: const CircleBorder(),
              onTap: isRecording ? onPause : onPlay,
              child: SizedBox(
                width: 72,
                height: 72,
                child: Center(
                  child: isRecording
                      ? const _PauseIcon()
                      : const _PlayIcon(),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _PlayIcon extends StatelessWidget {
  const _PlayIcon();

  @override
  Widget build(BuildContext context) {
    return Icon(
      Icons.play_arrow,
      color: Theme.of(context).colorScheme.onPrimary,
      size: 40,
    );
  }
}

class _PauseIcon extends StatelessWidget {
  const _PauseIcon();

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.onPrimary;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 6, height: 28, color: color),
        const SizedBox(width: 8),
        Container(width: 6, height: 28, color: color),
      ],
    );
  }
}
