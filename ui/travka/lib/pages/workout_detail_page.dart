import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../api_request.dart';
import '../models/workout.dart';
import '../services/track_parser.dart';
import '../widgets/workout_info_section.dart';
import '../widgets/workout_map_expand_button.dart';
import '../widgets/workout_map_preview.dart';
import '../widgets/workout_record_map.dart';

class WorkoutDetailView extends StatefulWidget {
  const WorkoutDetailView({
    super.key,
    required this.workout,
    required this.authToken,
  });

  final Workout workout;
  final String authToken;

  @override
  State<WorkoutDetailView> createState() => _WorkoutDetailViewState();
}

class _WorkoutDetailViewState extends State<WorkoutDetailView> {
  final ApiRequest _api = ApiRequest();
  final MapController _mapController = MapController();

  List<LatLng>? _trackPoints;
  bool _isLoadingTrack = false;
  String? _trackError;
  bool _isMapExpanded = false;

  bool get _hasTrack => widget.workout.track.isNotEmpty;

  bool get _hasInteractiveMap {
    final points = _trackPoints;
    return points != null &&
        points.length >= 2 &&
        !_isLoadingTrack &&
        _trackError == null;
  }

  @override
  void initState() {
    super.initState();
    if (_hasTrack) {
      _loadTrack();
    }
  }

  Future<void> _loadTrack() async {
    setState(() {
      _isLoadingTrack = true;
      _trackError = null;
    });

    try {
      final downloaded = await _api.downloadWorkoutTrack(
        token: widget.authToken,
        workoutId: widget.workout.id,
        fallbackFilename: widget.workout.track,
      );
      final points = parseTrackPoints(downloaded.bytes, downloaded.filename);
      if (!mounted) {
        return;
      }
      setState(() {
        _trackPoints = points;
        _isLoadingTrack = false;
      });
    } on ApiException catch (e) {
      if (!mounted) {
        return;
      }
      setState(() {
        _trackError = e.message;
        _isLoadingTrack = false;
      });
    } on TrackParseException catch (e) {
      if (!mounted) {
        return;
      }
      setState(() {
        _trackError = e.message;
        _isLoadingTrack = false;
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _trackError = l10n.failedToLoadWorkoutTrack;
        _isLoadingTrack = false;
      });
    }
  }

  void _collapseMap() {
    if (_isMapExpanded) {
      setState(() => _isMapExpanded = false);
    }
  }

  void _expandMap() {
    if (!_isMapExpanded) {
      setState(() => _isMapExpanded = true);
    }
  }

  Widget _buildWorkoutRecordMap(List<LatLng> points) {
    return WorkoutRecordMap(
      controller: _mapController,
      trackPoints: points,
      followUser: false,
      showCurrentLocation: false,
      fitToTrack: true,
      initialCenter: points.first,
    );
  }

  Widget _buildTrackSection(AppLocalizations l10n) {
    if (!_hasTrack) {
      return const SizedBox.shrink();
    }

    if (_isLoadingTrack) {
      return const WorkoutMapPreview(
        child: Center(child: CircularProgressIndicator()),
      );
    }

    if (_trackError != null) {
      return WorkoutMapPreview(
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _trackError!,
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: _loadTrack,
                  child: Text(l10n.retry),
                ),
              ],
            ),
          ),
        ),
      );
    }

    final points = _trackPoints;
    if (points == null || points.length < 2) {
      return const SizedBox.shrink();
    }

    if (_isMapExpanded) {
      return const SizedBox.shrink();
    }

    return WorkoutMapPreview(
      child: WorkoutMapExpandButton(
        isExpanded: false,
        onToggle: _expandMap,
        child: _buildWorkoutRecordMap(points),
      ),
    );
  }

  Widget _buildExpandedMapOverlay(List<LatLng> points) {
    return Material(
      color: Theme.of(context).scaffoldBackgroundColor,
      child: WorkoutMapExpandButton(
        isExpanded: true,
        onToggle: _collapseMap,
        child: _buildWorkoutRecordMap(points),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final points = _trackPoints;

    return PopScope(
      canPop: !_isMapExpanded,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop && _isMapExpanded) {
          _collapseMap();
        }
      },
      child: LayoutBuilder(
        builder: (context, constraints) {
          return Stack(
            children: [
              IgnorePointer(
                ignoring: _isMapExpanded,
                child: Opacity(
                  opacity: _isMapExpanded ? 0 : 1,
                  child: SizedBox(
                    height: constraints.maxHeight,
                    child: SingleChildScrollView(
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Padding(
                            padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  widget.workout.name,
                                  style: theme.textTheme.titleMedium?.copyWith(
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                                const SizedBox(height: 12),
                                WorkoutInfoSection(workout: widget.workout),
                              ],
                            ),
                          ),
                          if (_hasTrack)
                            Padding(
                              padding: const EdgeInsets.fromLTRB(16, 16, 16, 16),
                              child: _buildTrackSection(l10n),
                            ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              if (_isMapExpanded && _hasInteractiveMap && points != null)
                Positioned.fill(
                  child: _buildExpandedMapOverlay(points),
                ),
            ],
          );
        },
      ),
    );
  }
}
