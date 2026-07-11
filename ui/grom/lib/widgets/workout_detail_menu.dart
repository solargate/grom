import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

enum WorkoutDetailMenuAction { edit, delete }

class WorkoutDetailMenu extends StatelessWidget {
  const WorkoutDetailMenu({
    super.key,
    this.onSelected,
  });

  final ValueChanged<WorkoutDetailMenuAction>? onSelected;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return PopupMenuButton<WorkoutDetailMenuAction>(
      tooltip: l10n.workoutActions,
      onSelected: onSelected,
      itemBuilder: (context) => [
        PopupMenuItem(
          value: WorkoutDetailMenuAction.edit,
          child: Text(l10n.editWorkout),
        ),
        PopupMenuItem(
          value: WorkoutDetailMenuAction.delete,
          child: Text(l10n.deleteWorkout),
        ),
      ],
    );
  }
}
