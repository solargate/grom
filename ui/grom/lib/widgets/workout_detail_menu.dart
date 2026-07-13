import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

enum WorkoutDetailMenuAction {
  edit,
  delete,
  downloadGpx,
  downloadOriginal,
}

class WorkoutDetailMenu extends StatelessWidget {
  const WorkoutDetailMenu({
    super.key,
    required this.hasTrack,
    required this.canDownloadOriginal,
    this.canEdit = false,
    this.canDelete = false,
    this.onSelected,
  });

  final bool hasTrack;
  final bool canDownloadOriginal;
  final bool canEdit;
  final bool canDelete;
  final ValueChanged<WorkoutDetailMenuAction>? onSelected;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return PopupMenuButton<WorkoutDetailMenuAction>(
      tooltip: l10n.workoutActions,
      onSelected: onSelected,
      itemBuilder: (context) => [
        if (canEdit)
          PopupMenuItem(
            value: WorkoutDetailMenuAction.edit,
            child: Text(l10n.editWorkout),
          ),
        if (canDelete)
          PopupMenuItem(
            value: WorkoutDetailMenuAction.delete,
            child: Text(l10n.deleteWorkout),
          ),
        if (hasTrack)
          PopupMenuItem(
            value: WorkoutDetailMenuAction.downloadGpx,
            child: Text(l10n.downloadTrackAsGpx),
          ),
        if (hasTrack && canDownloadOriginal)
          PopupMenuItem(
            value: WorkoutDetailMenuAction.downloadOriginal,
            child: Text(l10n.downloadTrackOriginal),
          ),
      ],
    );
  }
}
