import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../models/workout.dart';
import 'user_avatar.dart';

class WorkoutLikeBar extends StatefulWidget {
  const WorkoutLikeBar({
    super.key,
    required this.workout,
    required this.authToken,
  });

  final Workout workout;
  final String authToken;

  @override
  State<WorkoutLikeBar> createState() => _WorkoutLikeBarState();
}

class _WorkoutLikeBarState extends State<WorkoutLikeBar> {
  final ApiRequest _api = ApiRequest();

  late int _likesCount = widget.workout.likesCount;
  late bool _likedByMe = widget.workout.likedByMe;
  bool _isSaving = false;

  String? get _owner => widget.workout.ownerNickname.isNotEmpty
      ? widget.workout.ownerNickname
      : null;

  @override
  void didUpdateWidget(covariant WorkoutLikeBar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.workout.id != widget.workout.id ||
        oldWidget.workout.likesCount != widget.workout.likesCount ||
        oldWidget.workout.likedByMe != widget.workout.likedByMe) {
      _likesCount = widget.workout.likesCount;
      _likedByMe = widget.workout.likedByMe;
    }
  }

  Future<void> _toggleLike() async {
    if (!widget.workout.canLike || _isSaving) {
      return;
    }
    setState(() => _isSaving = true);
    try {
      final state = _likedByMe
          ? await _api.unlikeWorkout(
              token: widget.authToken,
              workoutId: widget.workout.id,
              owner: _owner,
            )
          : await _api.likeWorkout(
              token: widget.authToken,
              workoutId: widget.workout.id,
              owner: _owner,
            );
      if (!mounted) {
        return;
      }
      setState(() {
        _likesCount = state.count;
        _likedByMe = state.likedByMe;
        _isSaving = false;
      });
    } on ApiException catch (e) {
      if (!mounted) {
        return;
      }
      setState(() => _isSaving = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      final l10n = AppLocalizations.of(context)!;
      setState(() => _isSaving = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToUpdateWorkoutLike)),
      );
    }
  }

  Future<void> _showLikes() async {
    final l10n = AppLocalizations.of(context)!;
    try {
      final likes = await _api.getWorkoutLikes(
        token: widget.authToken,
        workoutId: widget.workout.id,
        owner: _owner,
      );
      if (!mounted) {
        return;
      }
      final users = likes.users;
      await showModalBottomSheet<void>(
        context: context,
        isScrollControlled: true,
        showDragHandle: true,
        builder: (context) {
          return SafeArea(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    l10n.workoutLikesTitle(likes.count.toString()),
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  const SizedBox(height: 16),
                  if (users.isEmpty)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Text(l10n.workoutNoLikesYet),
                    )
                  else
                    Flexible(
                      child: ListView.separated(
                        shrinkWrap: true,
                        itemCount: users.length,
                        separatorBuilder: (_, __) => const Divider(height: 1),
                        itemBuilder: (context, index) {
                          final user = users[index];
                          return ListTile(
                            contentPadding: EdgeInsets.zero,
                            leading: UserAvatar(
                              nickname: user.nickname,
                              hasAvatar: user.hasAvatar,
                              avatarUrl: user.avatarUrl,
                              authToken: widget.authToken,
                            ),
                            title: Text(
                              user.name.isNotEmpty ? user.name : user.nickname,
                            ),
                            subtitle: Text(user.handle),
                          );
                        },
                      ),
                    ),
                ],
              ),
            ),
          );
        },
      );
    } on ApiException catch (e) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToLoadWorkoutLikes)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final iconColor = _likedByMe
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurfaceVariant;

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
      child: Row(
        children: [
          IconButton(
            onPressed: widget.workout.canLike ? _toggleLike : null,
            icon: _isSaving
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Icon(
                    _likedByMe ? Icons.thumb_up : Icons.thumb_up_outlined,
                    color: iconColor,
                  ),
            tooltip: AppLocalizations.of(context)!.workoutLikeAction,
          ),
          InkWell(
            onTap: _showLikes,
            borderRadius: BorderRadius.circular(8),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              child: Text(
                '$_likesCount',
                style: theme.textTheme.titleSmall,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
