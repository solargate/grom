import 'dart:math' as math;

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';

import '../api_request.dart';
import '../models/workout.dart';
import 'workout_map_preview.dart';

const kWorkoutMediaVisibleCount = 7;
const kWorkoutMediaGap = 4.0;

class WorkoutMediaStrip extends StatefulWidget {
  const WorkoutMediaStrip({
    super.key,
    required this.workout,
    required this.authToken,
    this.onPhotoTap,
  });

  final Workout workout;
  final String authToken;
  final ValueChanged<int>? onPhotoTap;

  @override
  State<WorkoutMediaStrip> createState() => _WorkoutMediaStripState();
}

class _WorkoutMediaStripState extends State<WorkoutMediaStrip> {
  final _scrollController = ScrollController();
  final _api = ApiRequest();

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _handlePointerSignal(PointerSignalEvent event) {
    if (event is! PointerScrollEvent || !_scrollController.hasClients) {
      return;
    }
    final delta = event.scrollDelta.dy + event.scrollDelta.dx;
    if (delta == 0) {
      return;
    }
    final target = (_scrollController.offset + delta).clamp(
      _scrollController.position.minScrollExtent,
      _scrollController.position.maxScrollExtent,
    );
    _scrollController.jumpTo(target);
  }

  @override
  Widget build(BuildContext context) {
    if (!widget.workout.hasMedia || widget.workout.mediaFiles.isEmpty) {
      return const SizedBox.shrink();
    }

    final owner = widget.workout.ownerNickname;
    final headers = {'Authorization': 'Bearer ${widget.authToken}'};

    return LayoutBuilder(
      builder: (context, constraints) {
        final displayWidth = math.min(
          kWorkoutMapPreviewMaxWidth,
          constraints.maxWidth,
        );
        final thumbSize = displayWidth / kWorkoutMediaVisibleCount;

        return Align(
          alignment: Alignment.centerLeft,
          child: SizedBox(
            width: displayWidth,
            height: thumbSize,
            child: Listener(
              onPointerSignal: _handlePointerSignal,
              child: ListView.separated(
                controller: _scrollController,
                scrollDirection: Axis.horizontal,
                itemCount: widget.workout.mediaFiles.length,
                separatorBuilder: (_, __) =>
                    const SizedBox(width: kWorkoutMediaGap),
                itemBuilder: (context, index) {
                  final filename = widget.workout.mediaFiles[index];
                  return _WorkoutMediaThumb(
                    previewUrl: _api.mediaPreviewUrl(
                      widget.workout.id,
                      filename,
                      owner: owner.isNotEmpty ? owner : null,
                    ),
                    headers: headers,
                    size: thumbSize,
                    onTap: widget.onPhotoTap == null
                        ? null
                        : () => widget.onPhotoTap!(index),
                  );
                },
              ),
            ),
          ),
        );
      },
    );
  }
}

class _WorkoutMediaThumb extends StatelessWidget {
  const _WorkoutMediaThumb({
    required this.previewUrl,
    required this.headers,
    required this.size,
    this.onTap,
  });

  final String previewUrl;
  final Map<String, String> headers;
  final double size;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Theme.of(context).colorScheme.surfaceContainerHighest,
      clipBehavior: Clip.antiAlias,
      child: GestureDetector(
        onTap: onTap,
        child: SizedBox(
          width: size,
          height: size,
          child: Image.network(
            previewUrl,
            headers: headers,
            fit: BoxFit.cover,
            width: size,
            height: size,
            filterQuality: FilterQuality.medium,
            errorBuilder: (context, error, stackTrace) {
              return const Center(
                child: Icon(Icons.broken_image_outlined),
              );
            },
          ),
        ),
      ),
    );
  }
}
