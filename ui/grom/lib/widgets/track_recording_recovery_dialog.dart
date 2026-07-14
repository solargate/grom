import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../services/track_recording_service.dart';

Future<void> showTrackRecordingRecoveryDialog(
  BuildContext context,
  TrackRecordingService recorder,
) async {
  final l10n = AppLocalizations.of(context)!;
  final restore = await showDialog<bool>(
    context: context,
    barrierDismissible: false,
    builder: (context) => AlertDialog(
      title: Text(l10n.restoreRecordingTitle),
      content: Text(l10n.restoreRecordingMessage),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: Text(l10n.restoreRecordingDiscard),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, true),
          child: Text(l10n.restoreRecordingConfirm),
        ),
      ],
    ),
  );

  if (restore == true) {
    try {
      await recorder.restoreInterruptedRecording(
        notificationStrings: TrackRecordingNotificationStrings(
          title: l10n.recordingNotificationTitle,
          text: l10n.recordingNotificationText,
          channelName: l10n.recordingNotificationChannelName,
          pausedText: l10n.recordingPausedNotificationText,
          autoPausedText: l10n.recordingAutoPausedNotificationText,
        ),
      );
    } catch (_) {
      if (!context.mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.locationPermissionDenied)),
      );
    }
    return;
  }

  await recorder.discardRecording();
}
