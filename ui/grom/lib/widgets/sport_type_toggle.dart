import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../models/sport_types.dart';

/// Toggle button for a single sport type (selected = included in a filter).
class SportTypeToggleButton extends StatelessWidget {
  const SportTypeToggleButton({
    super.key,
    required this.sportTypeId,
    required this.selected,
    required this.onSelected,
  });

  final String sportTypeId;
  final bool selected;
  final ValueChanged<bool> onSelected;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final info = sportTypeById(sportTypeId);
    final color = sportTypeColor(sportTypeId);
    final label = sportTypeLabel(l10n, sportTypeId);
    final icon = info?.icon ?? Icons.sports;

    final mutedOutline = theme.colorScheme.outline.withValues(alpha: 0.45);
    final borderColor = selected ? color : mutedOutline;
    final fill = selected ? color.withValues(alpha: 0.18) : Colors.transparent;
    final foreground = selected
        ? color
        : theme.colorScheme.onSurface.withValues(alpha: 0.7);

    return Material(
      color: fill,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
        side: BorderSide(color: borderColor, width: selected ? 1.5 : 1),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(20),
        onTap: () => onSelected(!selected),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, size: 18, color: foreground),
              const SizedBox(width: 6),
              Text(
                label,
                style: theme.textTheme.labelLarge?.copyWith(color: foreground),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Wrapping row of [SportTypeToggleButton]s in [sportTypeIds] order.
class SportTypeToggleWrap extends StatelessWidget {
  const SportTypeToggleWrap({
    super.key,
    required this.sportTypeIds,
    required this.selectedIds,
    required this.onToggle,
    this.padding = const EdgeInsets.fromLTRB(12, 8, 12, 8),
  });

  final List<String> sportTypeIds;
  final Set<String> selectedIds;
  final ValueChanged<String> onToggle;
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context) {
    if (sportTypeIds.isEmpty) {
      return const SizedBox.shrink();
    }

    return Padding(
      padding: padding,
      child: Wrap(
        spacing: 8,
        runSpacing: 8,
        children: [
          for (final id in sportTypeIds)
            SportTypeToggleButton(
              sportTypeId: id,
              selected: selectedIds.contains(id),
              onSelected: (_) => onToggle(id),
            ),
        ],
      ),
    );
  }
}
