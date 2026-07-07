import 'dart:math' as math;

import 'package:flutter/material.dart';

const double kWorkoutMapPreviewAspectRatio = 640 / 360;
const double kWorkoutMapPreviewMaxWidth = 640;

class WorkoutMapPreview extends StatelessWidget {
  const WorkoutMapPreview({
    super.key,
    required this.child,
  });

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final displayWidth = math.min(
          kWorkoutMapPreviewMaxWidth,
          constraints.maxWidth,
        );
        return Align(
          alignment: Alignment.centerLeft,
          child: SizedBox(
            width: displayWidth,
            height: displayWidth / kWorkoutMapPreviewAspectRatio,
            child: child,
          ),
        );
      },
    );
  }
}
