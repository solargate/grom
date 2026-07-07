import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/workout.dart';
import '../pages/workout_detail_page.dart';
import '../widgets/workout_card.dart';

class HomePage extends StatefulWidget {
  const HomePage({
    super.key,
    this.nickname,
    this.refreshToken = 0,
    this.viewingWorkout,
    this.onViewingWorkoutChanged,
  });

  final String? nickname;
  final int refreshToken;
  final Workout? viewingWorkout;
  final ValueChanged<Workout?>? onViewingWorkoutChanged;

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
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
  void didUpdateWidget(covariant HomePage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.nickname != oldWidget.nickname ||
        widget.refreshToken != oldWidget.refreshToken) {
      _loadWorkouts();
    }
  }

  Future<void> _loadWorkouts() async {
    if (widget.nickname == null) {
      setState(() {
        _workouts = [];
        _error = null;
        _isLoading = false;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }

      final workouts = await _api.listWorkouts(token);
      if (!mounted) return;
      setState(() {
        _workouts = workouts;
        _authToken = token;
        _isLoading = false;
      });
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

  void _openWorkout(Workout workout) {
    widget.onViewingWorkoutChanged?.call(workout);
  }

  @override
  Widget build(BuildContext context) {
    if (widget.nickname == null) {
      return const SizedBox.shrink();
    }

    final viewingWorkout = widget.viewingWorkout;
    if (viewingWorkout != null && _authToken != null) {
      return WorkoutDetailView(
        workout: viewingWorkout,
        authToken: _authToken!,
      );
    }

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
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: _workouts.length,
        itemBuilder: (context, index) {
          return WorkoutCard(
            workout: _workouts[index],
            authToken: _authToken ?? '',
            onTap: () => _openWorkout(_workouts[index]),
          );
        },
      ),
    );
  }
}
