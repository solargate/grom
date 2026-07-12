import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/workout.dart';
import '../pages/workout_detail_page.dart';
import '../widgets/workout_card.dart';
import '../widgets/workout_photo_viewer.dart';

class HomePage extends StatefulWidget {
  const HomePage({
    super.key,
    this.nickname,
    this.federationEnabled = false,
    this.refreshToken = 0,
    this.viewingWorkout,
    this.isMapExpanded = false,
    this.onViewingWorkoutChanged,
    this.onMapExpandedChanged,
    this.photoViewerIndex,
    this.onPhotoViewerIndexChanged,
    this.feedPhotoViewerWorkout,
    this.onFeedPhotoViewerWorkoutChanged,
  });

  final String? nickname;
  final bool federationEnabled;
  final int refreshToken;
  final Workout? viewingWorkout;
  final bool isMapExpanded;
  final ValueChanged<Workout?>? onViewingWorkoutChanged;
  final ValueChanged<bool>? onMapExpandedChanged;
  final int? photoViewerIndex;
  final ValueChanged<int?>? onPhotoViewerIndexChanged;
  final Workout? feedPhotoViewerWorkout;
  final ValueChanged<Workout?>? onFeedPhotoViewerWorkoutChanged;

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
    widget.onFeedPhotoViewerWorkoutChanged?.call(null);
    widget.onPhotoViewerIndexChanged?.call(null);
    widget.onViewingWorkoutChanged?.call(workout);
  }

  void _openWorkoutPhoto(Workout workout, int index) {
    widget.onFeedPhotoViewerWorkoutChanged?.call(workout);
    widget.onPhotoViewerIndexChanged?.call(index);
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
        federationEnabled: widget.federationEnabled,
        isMapExpanded: widget.isMapExpanded,
        onMapExpandedChanged: widget.onMapExpandedChanged,
        photoViewerIndex: widget.photoViewerIndex,
        onPhotoViewerIndexChanged: widget.onPhotoViewerIndexChanged,
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

    final feedPhotoWorkout = widget.feedPhotoViewerWorkout;
    final photoViewerIndex = widget.photoViewerIndex;
    final showFeedPhotoViewer = feedPhotoWorkout != null &&
        photoViewerIndex != null &&
        feedPhotoWorkout.mediaFiles.isNotEmpty &&
        _authToken != null;

    return Stack(
      children: [
        RefreshIndicator(
          onRefresh: _loadWorkouts,
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: _workouts.length,
            itemBuilder: (context, index) {
              return WorkoutCard(
                workout: _workouts[index],
                authToken: _authToken ?? '',
                federationEnabled: widget.federationEnabled,
                onTap: () => _openWorkout(_workouts[index]),
                onPhotoTap: (photoIndex) =>
                    _openWorkoutPhoto(_workouts[index], photoIndex),
              );
            },
          ),
        ),
        if (showFeedPhotoViewer)
          Positioned.fill(
            child: WorkoutPhotoViewer(
              workout: feedPhotoWorkout,
              authToken: _authToken!,
              initialIndex: photoViewerIndex.clamp(
                0,
                feedPhotoWorkout.mediaFiles.length - 1,
              ),
            ),
          ),
      ],
    );
  }
}
