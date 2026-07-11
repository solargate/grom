import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/equipment_types.dart';
import 'package:grom/models/social.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/models/workout.dart';

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
  });

  final Workout workout;
  final String authToken;
  final WorkoutAuthor? author;
  final bool federationEnabled;
  final int? descriptionMaxLines;

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
    final durationText = formatDuration(l10n, workout.durationSeconds);
    final distanceText = formatDistanceKm(l10n, workout.distance);
    final metadataParts = <String>[
      sportTypeLabel(l10n, workout.sportType),
      dateText,
      if (workout.device.isNotEmpty) workout.device,
    ];

    return Row(
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
        Expanded(
          child: Column(
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
              if (workout.equipment.isNotEmpty) ...[
                const SizedBox(height: 4),
                _EquipmentLine(equipment: workout.equipment),
              ],
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
              const SizedBox(height: 12),
              _StatsRow(
                distanceLabel: l10n.workoutDistance,
                distanceValue: distanceText,
                durationLabel: l10n.workoutDuration,
                durationValue: durationText,
              ),
            ],
          ),
        ),
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

class _StatsRow extends StatelessWidget {
  const _StatsRow({
    required this.distanceLabel,
    required this.distanceValue,
    required this.durationLabel,
    required this.durationValue,
  });

  final String distanceLabel;
  final String distanceValue;
  final String durationLabel;
  final String durationValue;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final labelStyle = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    final valueStyle = theme.textTheme.titleMedium?.copyWith(
      fontWeight: FontWeight.w600,
    );

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _StatBlock(
          label: distanceLabel,
          value: distanceValue,
          labelStyle: labelStyle,
          valueStyle: valueStyle,
        ),
        const Padding(
          padding: EdgeInsets.symmetric(horizontal: 12),
          child: WorkoutFieldSeparator(),
        ),
        _StatBlock(
          label: durationLabel,
          value: durationValue,
          labelStyle: labelStyle,
          valueStyle: valueStyle,
        ),
      ],
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
