import 'package:flutter/material.dart';

import '../app_theme.dart';
import '../services/track_recording_service.dart';

enum PauseButtonVariant { none, manual, auto }

class WorkoutRecordControls extends StatelessWidget {
  const WorkoutRecordControls({
    super.key,
    required this.state,
    required this.pauseVariant,
    required this.onPlay,
    required this.onPause,
    required this.onFinish,
    required this.playLabel,
    required this.pauseLabel,
    required this.finishLabel,
  });

  final TrackRecordingState state;
  final PauseButtonVariant pauseVariant;
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
    final isAutoPaused = state == TrackRecordingState.autoPaused;
    final showFinish = state == TrackRecordingState.paused;
    final showPauseButton = isRecording || isAutoPaused;

    final pauseColor = pauseVariant == PauseButtonVariant.auto
        ? kAutoPauseColor
        : colorScheme.primary;

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
          label: showPauseButton ? pauseLabel : playLabel,
          child: Material(
            color: showPauseButton ? pauseColor : colorScheme.primary,
            shape: const CircleBorder(),
            elevation: 2,
            child: InkWell(
              customBorder: const CircleBorder(),
              onTap: showPauseButton ? onPause : onPlay,
              child: SizedBox(
                width: 72,
                height: 72,
                child: Center(
                  child: showPauseButton
                      ? _PauseIcon(
                          color: pauseVariant == PauseButtonVariant.auto
                              ? Colors.white
                              : colorScheme.onPrimary,
                        )
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
  const _PauseIcon({required this.color});

  final Color color;

  @override
  Widget build(BuildContext context) {
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
