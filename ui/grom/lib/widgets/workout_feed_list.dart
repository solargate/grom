import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/workout.dart';
import 'workout_card.dart';

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
  });

  final String nickname;
  final String scope;
  final ScrollController scrollController;
  final int refreshToken;
  final bool federationEnabled;
  final ValueChanged<Workout> onWorkoutTap;
  final void Function(Workout workout, int photoIndex) onPhotoTap;
  final ValueChanged<String>? onAuthTokenLoaded;

  @override
  State<WorkoutFeedList> createState() => WorkoutFeedListState();
}

class WorkoutFeedListState extends State<WorkoutFeedList> {
  final ApiRequest _api = ApiRequest();

  List<Workout> _workouts = [];
  bool _isLoading = false;
  String? _error;
  String? _authToken;

  @override
  void initState() {
    super.initState();
    _loadWorkouts();
  }

  @override
  void didUpdateWidget(covariant WorkoutFeedList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.refreshToken != oldWidget.refreshToken ||
        widget.nickname != oldWidget.nickname ||
        widget.scope != oldWidget.scope) {
      _loadWorkouts();
    }
  }

  Future<void> reload() => _loadWorkouts();

  Future<void> _loadWorkouts() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }

      final workouts = await _api.listWorkouts(token, scope: widget.scope);
      if (!mounted) return;
      setState(() {
        _workouts = workouts;
        _authToken = token;
        _isLoading = false;
      });
      widget.onAuthTokenLoaded?.call(token);
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _error = l10n.failedToLoadWorkouts;
        _isLoading = false;
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
                onPressed: _loadWorkouts,
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
            l10n.noWorkoutsYet,
            style: Theme.of(context).textTheme.titleMedium,
            textAlign: TextAlign.center,
          ),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadWorkouts,
      child: ListView.builder(
        controller: widget.scrollController,
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: _workouts.length,
        itemBuilder: (context, index) {
          final workout = _workouts[index];
          return WorkoutCard(
            workout: workout,
            authToken: _authToken ?? '',
            federationEnabled: widget.federationEnabled,
            onTap: () => widget.onWorkoutTap(workout),
            onPhotoTap: (photoIndex) =>
                widget.onPhotoTap(workout, photoIndex),
          );
        },
      ),
    );
  }
}
