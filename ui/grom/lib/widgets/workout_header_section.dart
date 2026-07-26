import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/equipment_types.dart';
import 'package:grom/models/social.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/models/workout.dart';
import 'package:grom/models/workout_stats.dart';
import 'package:grom/platform/is_mobile_client.dart';

import 'user_avatar.dart';
import 'workout_field_separator.dart';

const _avatarRadius = 30.0;
const _sportAvatarRadius = _avatarRadius / 1.5;

class WorkoutHeaderSection extends StatelessWidget {
  const WorkoutHeaderSection({
    super.key,
    required this.workout,
    required this.authToken,
    this.author,
    this.federationEnabled = false,
    this.descriptionMaxLines,
    this.statsMaxRows,
    this.showEquipment = false,
  });

  final Workout workout;
  final String authToken;
  final WorkoutAuthor? author;
  final bool federationEnabled;
  final int? descriptionMaxLines;

  /// When set, only the first N rows of stats (3 per row) are shown.
  final int? statsMaxRows;

  /// When true, equipment is shown below the stats table (detail view).
  final bool showEquipment;

  String _authorLine(WorkoutAuthor author) {
    final displayName =
        author.name.isNotEmpty ? author.name : author.nickname;
    final nicknameInParens =
        federationEnabled ? author.handle : author.nickname;
    return '$displayName ($nicknameInParens)';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final sportInfo = sportTypeById(workout.sportType);
    final sportIcon = sportInfo?.icon ?? Icons.sports;
    final sportColor = sportTypeColor(workout.sportType);
    final dateText = DateFormat.yMMMd(l10n.localeName).add_Hm().format(
          workout.startDate.toLocal(),
        );
    final metadataParts = <String>[
      sportTypeLabel(l10n, workout.sportType),
      dateText,
      if (workout.device.isNotEmpty) workout.device,
    ];
    final stats = buildWorkoutStats(l10n, workout);
    final statsRows = chunkWorkoutStats(stats, maxRows: statsMaxRows);
    final flushStatsToAvatar = isMobileClient && author != null;
    final showEquipmentLine =
        showEquipment && workout.equipment.isNotEmpty;
    final equipmentLine = showEquipmentLine
        ? _EquipmentLine(equipment: workout.equipment)
        : null;

    final metaColumn = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (author != null) ...[
          Text(
            _authorLine(author!),
            style: theme.textTheme.bodyMedium,
          ),
          const SizedBox(height: 4),
        ],
        Wrap(
          crossAxisAlignment: WrapCrossAlignment.center,
          children: joinWorkoutTextFields(context, metadataParts),
        ),
        const SizedBox(height: 8),
        Text(
          workout.name,
          style: theme.textTheme.titleLarge?.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        if (workout.description.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(
            workout.description,
            maxLines: descriptionMaxLines,
            overflow: descriptionMaxLines != null
                ? TextOverflow.ellipsis
                : null,
            style: theme.textTheme.bodyMedium,
          ),
        ],
        if (!flushStatsToAvatar && statsRows.isNotEmpty) ...[
          const SizedBox(height: 12),
          _StatsGrid(rows: statsRows),
        ],
        if (!flushStatsToAvatar && equipmentLine != null) ...[
          SizedBox(height: statsRows.isNotEmpty ? 8 : 12),
          equipmentLine,
        ],
      ],
    );

    final headerRow = Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (author != null) ...[
          SizedBox(
            width: _avatarRadius * 2,
            child: Column(
              children: [
                UserAvatar(
                  nickname: author!.nickname,
                  hasAvatar: author!.hasAvatar,
                  avatarUrl: author!.avatarUrl,
                  authToken: authToken,
                  radius: _avatarRadius,
                ),
                const SizedBox(height: 6),
                CircleAvatar(
                  radius: _sportAvatarRadius,
                  backgroundColor: sportColor.withValues(alpha: 0.15),
                  child: Icon(
                    sportIcon,
                    color: sportColor,
                    size: _sportAvatarRadius,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
        ],
        Expanded(child: metaColumn),
      ],
    );

    final flushBelow = flushStatsToAvatar &&
        (statsRows.isNotEmpty || equipmentLine != null);
    if (!flushBelow) {
      return headerRow;
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        headerRow,
        if (statsRows.isNotEmpty) ...[
          const SizedBox(height: 12),
          _StatsGrid(rows: statsRows),
        ],
        if (equipmentLine != null) ...[
          SizedBox(height: statsRows.isNotEmpty ? 8 : 12),
          equipmentLine,
        ],
      ],
    );
  }
}

class _EquipmentLine extends StatelessWidget {
  const _EquipmentLine({required this.equipment});

  final List<WorkoutEquipmentItem> equipment;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final iconColor = theme.colorScheme.onSurfaceVariant;
    final textStyle = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );

    return Wrap(
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        for (var i = 0; i < equipment.length; i++) ...[
          if (i > 0) WorkoutFieldSeparator(style: textStyle),
          Icon(
            _equipmentIcon(equipment[i]),
            size: 14,
            color: iconColor,
          ),
          const SizedBox(width: 4),
          Text(equipment[i].name, style: textStyle),
        ],
      ],
    );
  }

  IconData _equipmentIcon(WorkoutEquipmentItem item) {
    final type = equipmentTypeFromId(item.type);
    if (type == null) {
      return Icons.category;
    }
    return equipmentTypeIcon(type);
  }
}

class _StatsGrid extends StatelessWidget {
  const _StatsGrid({required this.rows});

  final List<List<WorkoutStatItem>> rows;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final labelStyle = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    final valueStyle = theme.textTheme.titleMedium?.copyWith(
      fontWeight: FontWeight.w600,
    );
    final columnCount = rows.fold<int>(
      0,
      (max, row) => row.length > max ? row.length : max,
    );
    if (columnCount == 0) {
      return const SizedBox.shrink();
    }

    return Align(
      alignment: Alignment.centerLeft,
      child: Table(
        defaultColumnWidth: const IntrinsicColumnWidth(),
        defaultVerticalAlignment: TableCellVerticalAlignment.top,
        children: [
          for (var rowIndex = 0; rowIndex < rows.length; rowIndex++)
            TableRow(
              children: [
                for (var col = 0; col < columnCount; col++)
                  col < rows[rowIndex].length
                      ? Padding(
                          padding: EdgeInsets.only(
                            top: rowIndex == 0 ? 0 : 10,
                            left: col == 0 ? 0 : 24,
                          ),
                          child: _StatBlock(
                            label: rows[rowIndex][col].label,
                            value: rows[rowIndex][col].value,
                            labelStyle: labelStyle,
                            valueStyle: valueStyle,
                          ),
                        )
                      : const SizedBox.shrink(),
              ],
            ),
        ],
      ),
    );
  }
}

class _StatBlock extends StatelessWidget {
  const _StatBlock({
    required this.label,
    required this.value,
    required this.labelStyle,
    required this.valueStyle,
  });

  final String label;
  final String value;
  final TextStyle? labelStyle;
  final TextStyle? valueStyle;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: labelStyle),
        const SizedBox(height: 2),
        Text(value, style: valueStyle),
      ],
    );
  }
}
