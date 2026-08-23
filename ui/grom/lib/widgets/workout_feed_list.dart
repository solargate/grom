import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/my_workouts_layout.dart';
import '../models/workout.dart';
import '../platform/is_mobile_client.dart';
import 'workout_card.dart';
import 'workout_list_row.dart';

class WorkoutFeedList extends StatefulWidget {
  const WorkoutFeedList({
    super.key,
    required this.nickname,
    required this.scope,
    required this.scrollController,
    required this.refreshToken,
    required this.federationEnabled,
    required this.onWorkoutTap,
    required this.onPhotoTap,
    this.onAuthTokenLoaded,
    this.api,
    this.sportTypes,
    this.emptyMessage,
    this.layout = MyWorkoutsLayout.cards,
  });

  final String nickname;
  final String scope;
  final ScrollController scrollController;
  final int refreshToken;
  final bool federationEnabled;
  final ValueChanged<Workout> onWorkoutTap;
  final void Function(Workout workout, int photoIndex) onPhotoTap;
  final ValueChanged<String>? onAuthTokenLoaded;
  final ApiRequest? api;

  /// When non-null, sent as `sport_types` (empty list → empty query value).
  final List<String>? sportTypes;

  /// Overrides [AppLocalizations.noWorkoutsYet] when the list is empty.
  final String? emptyMessage;

  /// Card vs compact list row. List mode is intended for `scope=own` only.
  final MyWorkoutsLayout layout;

  @override
  State<WorkoutFeedList> createState() => WorkoutFeedListState();
}

class WorkoutFeedListState extends State<WorkoutFeedList> {
  static const _pageLimit = 20;

  late final ApiRequest _api = widget.api ?? ApiRequest();

  List<Workout> _workouts = [];
  bool _isLoading = false;
  bool _isLoadingMore = false;
  bool _hasMore = false;
  String? _nextCursor;
  String? _error;
  String? _authToken;

  @override
  void initState() {
    super.initState();
    widget.scrollController.addListener(_onScroll);
    _loadWorkouts(reset: true);
  }

  @override
  void dispose() {
    widget.scrollController.removeListener(_onScroll);
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant WorkoutFeedList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.scrollController != widget.scrollController) {
      oldWidget.scrollController.removeListener(_onScroll);
      widget.scrollController.addListener(_onScroll);
    }
    if (widget.refreshToken != oldWidget.refreshToken ||
        widget.nickname != oldWidget.nickname ||
        widget.scope != oldWidget.scope ||
        !_sameSportTypes(oldWidget.sportTypes, widget.sportTypes)) {
      _loadWorkouts(reset: true);
    }
  }

  static bool _sameSportTypes(List<String>? a, List<String>? b) {
    if (identical(a, b)) {
      return true;
    }
    if (a == null || b == null) {
      return a == b;
    }
    if (a.length != b.length) {
      return false;
    }
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) {
        return false;
      }
    }
    return true;
  }

  Future<void> reload() => _loadWorkouts(reset: true);

  void _onScroll() {
    if (!_hasMore || _isLoadingMore || _isLoading) {
      return;
    }
    final position = widget.scrollController.position;
    if (position.pixels >= position.maxScrollExtent - 400) {
      _loadWorkouts(reset: false);
    }
  }

  Future<void> _loadWorkouts({required bool reset}) async {
    if (reset) {
      setState(() {
        _isLoading = true;
        _isLoadingMore = false;
        _error = null;
        _nextCursor = null;
        _hasMore = false;
      });
    } else {
      if (!_hasMore || _isLoadingMore || _nextCursor == null) {
        return;
      }
      setState(() {
        _isLoadingMore = true;
        _error = null;
      });
    }

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }

      final page = await _api.listWorkouts(
        token,
        scope: widget.scope,
        limit: _pageLimit,
        cursor: reset ? null : _nextCursor,
        sportTypes: widget.sportTypes,
      );
      if (!mounted) return;
      setState(() {
        if (reset) {
          _workouts = page.items;
        } else {
          _workouts = [..._workouts, ...page.items];
        }
        _nextCursor = page.nextCursor;
        _hasMore = page.hasMore && page.nextCursor != null;
        _authToken = token;
        _isLoading = false;
        _isLoadingMore = false;
      });
      widget.onAuthTokenLoaded?.call(token);
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _isLoading = false;
        _isLoadingMore = false;
      });
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _error = l10n.failedToLoadWorkouts;
        _isLoading = false;
        _isLoadingMore = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    if (_isLoading && _workouts.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null && _workouts.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_error!, textAlign: TextAlign.center),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: () => _loadWorkouts(reset: true),
                child: Text(l10n.retry),
              ),
            ],
          ),
        ),
      );
    }

    if (_workouts.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Text(
            widget.emptyMessage ?? l10n.noWorkoutsYet,
            style: Theme.of(context).textTheme.titleMedium,
            textAlign: TextAlign.center,
          ),
        ),
      );
    }

    final compact = isMobileClient;
    final itemCount = _workouts.length + (_isLoadingMore ? 1 : 0);
    final listView = ListView.builder(
      controller: widget.scrollController,
      padding: compact ? EdgeInsets.zero : const EdgeInsets.symmetric(vertical: 8),
      itemCount: itemCount,
      itemBuilder: (context, index) {
        if (index >= _workouts.length) {
          return const Padding(
            padding: EdgeInsets.symmetric(vertical: 16),
            child: Center(child: CircularProgressIndicator()),
          );
        }
        final workout = _workouts[index];
        final useList = widget.layout == MyWorkoutsLayout.list;
        return Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            AnimatedSwitcher(
              duration: const Duration(milliseconds: 200),
              switchInCurve: Curves.easeInOut,
              switchOutCurve: Curves.easeInOut,
              child: useList
                  ? WorkoutListRow(
                      key: ValueKey('list-${workout.id}'),
                      workout: workout,
                      onTap: () => widget.onWorkoutTap(workout),
                    )
                  : WorkoutCard(
                      key: ValueKey('card-${workout.id}'),
                      workout: workout,
                      authToken: _authToken ?? '',
                      federationEnabled: widget.federationEnabled,
                      compact: compact,
                      onTap: () => widget.onWorkoutTap(workout),
                      onPhotoTap: (photoIndex) =>
                          widget.onPhotoTap(workout, photoIndex),
                    ),
            ),
            if (compact && index < _workouts.length - 1)
              Container(
                height: 8,
                color: Theme.of(context).colorScheme.surfaceContainerHighest,
              ),
          ],
        );
      },
    );

    return RefreshIndicator(
      onRefresh: () => _loadWorkouts(reset: true),
      child: compact
          ? ScrollConfiguration(
              behavior: ScrollConfiguration.of(context)
                  .copyWith(scrollbars: false),
              child: listView,
            )
          : listView,
    );
  }
}
