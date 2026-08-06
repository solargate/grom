import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:intl/intl.dart';

import '../api_request.dart';
import '../models/workout.dart';
import 'user_avatar.dart';
import 'workout_map_preview.dart';

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
  late int _commentsCount = widget.workout.commentsCount;
  bool _isSaving = false;

  String? get _owner => widget.workout.ownerNickname.isNotEmpty
      ? widget.workout.ownerNickname
      : null;

  @override
  void didUpdateWidget(covariant WorkoutLikeBar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.workout.id != widget.workout.id ||
        oldWidget.workout.likesCount != widget.workout.likesCount ||
        oldWidget.workout.likedByMe != widget.workout.likedByMe ||
        oldWidget.workout.commentsCount != widget.workout.commentsCount) {
      _likesCount = widget.workout.likesCount;
      _likedByMe = widget.workout.likedByMe;
      _commentsCount = widget.workout.commentsCount;
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

  Future<void> _showComments() async {
    final l10n = AppLocalizations.of(context)!;
    try {
      final initial = await _api.getWorkoutComments(
        token: widget.authToken,
        workoutId: widget.workout.id,
        owner: _owner,
      );
      if (!mounted) {
        return;
      }
      await showModalBottomSheet<void>(
        context: context,
        isScrollControlled: true,
        showDragHandle: true,
        builder: (sheetContext) {
          return _CommentsSheet(
            authToken: widget.authToken,
            workoutId: widget.workout.id,
            owner: _owner,
            initial: initial,
            onCountChanged: (count) {
              if (mounted) {
                setState(() => _commentsCount = count);
              }
            },
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
        SnackBar(content: Text(l10n.failedToLoadWorkoutComments)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final iconColor = _likedByMe
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurfaceVariant;

    // Match map preview / media strip: same max width, left-aligned, so
    // comments sit on the right edge of that media block (not the screen).
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final displayWidth = math.min(
            kWorkoutMapPreviewMaxWidth,
            constraints.maxWidth,
          );
          return Align(
            alignment: Alignment.centerLeft,
            child: SizedBox(
              width: displayWidth,
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
                            _likedByMe
                                ? Icons.thumb_up
                                : Icons.thumb_up_outlined,
                            color: iconColor,
                          ),
                    tooltip: AppLocalizations.of(context)!.workoutLikeAction,
                  ),
                  InkWell(
                    onTap: _showLikes,
                    borderRadius: BorderRadius.circular(8),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      child: Text(
                        '$_likesCount',
                        style: theme.textTheme.titleSmall,
                      ),
                    ),
                  ),
                  const Spacer(),
                  InkWell(
                    onTap: _showComments,
                    borderRadius: BorderRadius.circular(8),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      child: Text(
                        '$_commentsCount',
                        style: theme.textTheme.titleSmall,
                      ),
                    ),
                  ),
                  IconButton(
                    onPressed: _showComments,
                    icon: Icon(
                      Icons.comment_outlined,
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                    tooltip:
                        AppLocalizations.of(context)!.workoutCommentAction,
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}

class _CommentsSheet extends StatefulWidget {
  const _CommentsSheet({
    required this.authToken,
    required this.workoutId,
    required this.owner,
    required this.initial,
    required this.onCountChanged,
  });

  final String authToken;
  final String workoutId;
  final String? owner;
  final WorkoutCommentsResponse initial;
  final ValueChanged<int> onCountChanged;

  @override
  State<_CommentsSheet> createState() => _CommentsSheetState();
}

class _CommentsSheetState extends State<_CommentsSheet> {
  final ApiRequest _api = ApiRequest();
  final TextEditingController _controller = TextEditingController();
  late List<WorkoutComment> _comments = List.of(widget.initial.comments);
  late int _count = widget.initial.count;
  bool _submitting = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final text = _controller.text.trim();
    if (text.isEmpty || _submitting) {
      return;
    }
    setState(() => _submitting = true);
    final l10n = AppLocalizations.of(context)!;
    try {
      final created = await _api.createWorkoutComment(
        token: widget.authToken,
        workoutId: widget.workoutId,
        text: text,
        owner: widget.owner,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _comments = [..._comments, created.comment];
        _count = created.count;
        _submitting = false;
        _controller.clear();
      });
      widget.onCountChanged(_count);
    } on ApiException catch (e) {
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToAddWorkoutComment)),
      );
    }
  }

  Future<void> _delete(WorkoutComment comment) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text(l10n.deleteWorkoutCommentTitle),
          content: Text(l10n.deleteWorkoutCommentConfirm),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: Text(l10n.cancel),
            ),
            TextButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: Text(l10n.deleteWorkoutCommentAction),
            ),
          ],
        );
      },
    );
    if (confirmed != true || !mounted) {
      return;
    }
    try {
      final count = await _api.deleteWorkoutComment(
        token: widget.authToken,
        workoutId: widget.workoutId,
        commentId: comment.id,
        owner: widget.owner,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _comments = _comments.where((c) => c.id != comment.id).toList();
        _count = count;
      });
      widget.onCountChanged(_count);
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
        SnackBar(content: Text(l10n.failedToDeleteWorkoutComment)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final bottom = MediaQuery.viewInsetsOf(context).bottom;
    final dateFormat = DateFormat.yMMMd().add_Hm();

    return Padding(
      padding: EdgeInsets.only(bottom: bottom),
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                l10n.workoutCommentsTitle(_count.toString()),
                style: theme.textTheme.titleLarge,
              ),
              const SizedBox(height: 16),
              if (_comments.isEmpty)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Text(l10n.workoutNoCommentsYet),
                )
              else
                Flexible(
                  child: ListView.separated(
                    shrinkWrap: true,
                    itemCount: _comments.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final comment = _comments[index];
                      final user = comment.user;
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
                        subtitle: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(user.handle),
                            const SizedBox(height: 4),
                            Text(
                              dateFormat.format(comment.datetime.toLocal()),
                              style: theme.textTheme.bodySmall,
                            ),
                            const SizedBox(height: 2),
                            Text(comment.text),
                          ],
                        ),
                        isThreeLine: true,
                        trailing: comment.canDelete
                            ? IconButton(
                                icon: const Icon(Icons.delete, size: 20),
                                tooltip: l10n.deleteWorkoutCommentAction,
                                onPressed: () => _delete(comment),
                              )
                            : null,
                      );
                    },
                  ),
                ),
              const SizedBox(height: 12),
              Row(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Expanded(
                    child: TextField(
                      controller: _controller,
                      maxLength: 1000,
                      minLines: 1,
                      maxLines: 4,
                      decoration: InputDecoration(
                        hintText: l10n.workoutCommentHint,
                        border: const OutlineInputBorder(),
                      ),
                      onSubmitted: (_) => _submit(),
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton.filled(
                    onPressed: _submitting ? null : _submit,
                    icon: _submitting
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.add_comment),
                    tooltip: l10n.addWorkoutCommentAction,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
