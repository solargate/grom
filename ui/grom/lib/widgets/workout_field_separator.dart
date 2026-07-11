import 'package:flutter/material.dart';

const workoutFieldSeparatorText = ' · ';

class WorkoutFieldSeparator extends StatelessWidget {
  const WorkoutFieldSeparator({super.key, this.style});

  final TextStyle? style;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Text(
      workoutFieldSeparatorText,
      style: style ??
          theme.textTheme.bodySmall?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
    );
  }
}

List<Widget> joinWorkoutFields(List<Widget> fields) {
  if (fields.isEmpty) {
    return const [];
  }

  final result = <Widget>[fields.first];
  for (var i = 1; i < fields.length; i++) {
    result.add(const WorkoutFieldSeparator());
    result.add(fields[i]);
  }
  return result;
}

List<Widget> joinWorkoutTextFields(
  BuildContext context,
  List<String> parts, {
  TextStyle? style,
}) {
  if (parts.isEmpty) {
    return const [];
  }

  final theme = Theme.of(context);
  final textStyle = style ??
      theme.textTheme.bodySmall?.copyWith(
        color: theme.colorScheme.onSurfaceVariant,
      );

  return joinWorkoutFields(
    parts.map((part) => Text(part, style: textStyle)).toList(),
  );
}
