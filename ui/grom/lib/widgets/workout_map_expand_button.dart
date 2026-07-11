import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

class WorkoutMapExpandButton extends StatelessWidget {
  const WorkoutMapExpandButton({
    super.key,
    required this.isExpanded,
    required this.onToggle,
    required this.child,
  });

  final bool isExpanded;
  final VoidCallback onToggle;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    return Stack(
      fit: StackFit.expand,
      children: [
        child,
        Positioned(
          top: 8,
          right: 8,
          child: Material(
            color: colorScheme.surface.withValues(alpha: 0.9),
            elevation: 2,
            shape: const CircleBorder(),
            child: IconButton(
              tooltip: isExpanded ? l10n.collapseMap : l10n.expandMap,
              icon: Icon(
                isExpanded ? Icons.fullscreen_exit : Icons.open_in_full,
              ),
              onPressed: onToggle,
            ),
          ),
        ),
      ],
    );
  }
}
