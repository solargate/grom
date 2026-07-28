import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../models/workout.dart';
import '../models/workout_speed.dart';
import '../services/track_parser.dart';
import '../widgets/workout_header_section.dart';
import '../widgets/workout_map_expand_button.dart';
import '../widgets/workout_map_preview.dart';
import '../widgets/workout_media_strip.dart';
import '../widgets/workout_photo_viewer.dart';
import '../widgets/workout_record_map.dart';
import '../widgets/workout_speed_chart.dart';

class WorkoutDetailView extends StatefulWidget {
  const WorkoutDetailView({
    super.key,
    required this.workout,
    required this.authToken,
    this.federationEnabled = false,
    this.isMapExpanded = false,
    this.onMapExpandedChanged,
    this.photoViewerIndex,
    this.onPhotoViewerIndexChanged,
  });

  final Workout workout;
  final String authToken;
  final bool federationEnabled;
  final bool isMapExpanded;
  final ValueChanged<bool>? onMapExpandedChanged;
  final int? photoViewerIndex;
  final ValueChanged<int?>? onPhotoViewerIndexChanged;

  @override
  State<WorkoutDetailView> createState() => _WorkoutDetailViewState();
}

class _WorkoutDetailViewState extends State<WorkoutDetailView> {
  final ApiRequest _api = ApiRequest();
  final MapController _mapController = MapController();

  List<LatLng>? _trackPoints;
  bool _isLoadingTrack = false;
  String? _trackError;

  List<WorkoutSpeedSample>? _speedSamples;
  double? _speedAvgKmh;
  double? _speedMaxKmh;
  bool _isLoadingSpeed = false;

  bool get _hasGpsMap => widget.workout.hasMapPreview;

  bool get _hasSpeedChart =>
      _speedSamples != null && _speedSamples!.length >= 2;

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
    if (_hasGpsMap) {
      _loadTrack();
    }
    _loadSpeed();
  }

  Future<void> _loadSpeed() async {
    setState(() {
      _isLoadingSpeed = true;
    });

    try {
      final series = await _api.getWorkoutSpeed(
        token: widget.authToken,
        workoutId: widget.workout.id,
        owner: widget.workout.ownerNickname.isNotEmpty
            ? widget.workout.ownerNickname
            : null,
      );
      if (!mounted) {
        return;
      }
      final samples = series.samples;
      setState(() {
        _speedSamples = samples;
        _speedAvgKmh = resolveSpeedAvgKmh(
          widget.workout.speedAvgKmh ?? series.speedAvgKmh,
          samples,
        );
        _speedMaxKmh = resolveSpeedMaxKmh(
          widget.workout.speedMaxKmh ?? series.speedMaxKmh,
          samples,
        );
        _isLoadingSpeed = false;
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _speedSamples = const [];
        _isLoadingSpeed = false;
      });
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
        owner: widget.workout.ownerNickname.isNotEmpty
            ? widget.workout.ownerNickname
            : null,
        format: 'gpx',
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
    if (widget.isMapExpanded) {
      widget.onMapExpandedChanged?.call(false);
    }
  }

  void _expandMap() {
    if (!widget.isMapExpanded) {
      widget.onMapExpandedChanged?.call(true);
    }
  }

  void _openPhotoViewer(int index) {
    widget.onPhotoViewerIndexChanged?.call(index);
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
    if (!_hasGpsMap) {
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

    if (widget.isMapExpanded) {
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
    final points = _trackPoints;
    final author = widget.workout.author;

    return LayoutBuilder(
      builder: (context, constraints) {
        final showPhotoViewer = widget.photoViewerIndex != null &&
            widget.workout.mediaFiles.isNotEmpty;

        return Stack(
          children: [
            IgnorePointer(
              ignoring: widget.isMapExpanded || showPhotoViewer,
              child: Opacity(
                opacity: (widget.isMapExpanded || showPhotoViewer) ? 0 : 1,
                child: SizedBox(
                  height: constraints.maxHeight,
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Padding(
                          padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                          child: WorkoutHeaderSection(
                            workout: widget.workout,
                            authToken: widget.authToken,
                            author: author,
                            federationEnabled: widget.federationEnabled,
                            showEquipment: true,
                          ),
                        ),
                        if (_hasGpsMap)
                          Padding(
                            padding: const EdgeInsets.fromLTRB(16, 16, 16, 16),
                            child: _buildTrackSection(l10n),
                          ),
                        if (widget.workout.hasMedia)
                          Padding(
                            padding: EdgeInsets.fromLTRB(
                              16,
                              _hasGpsMap ? 0 : 16,
                              16,
                              16,
                            ),
                            child: WorkoutMediaStrip(
                              workout: widget.workout,
                              authToken: widget.authToken,
                              onPhotoTap: _openPhotoViewer,
                            ),
                          ),
                        if (_hasSpeedChart)
                          Padding(
                            padding: EdgeInsets.fromLTRB(
                              16,
                              (widget.workout.hasMedia || _hasGpsMap) ? 0 : 16,
                              16,
                              16,
                            ),
                            child: WorkoutSpeedChart(
                              samples: _speedSamples!,
                              speedAvgKmh: _speedAvgKmh,
                              speedMaxKmh: _speedMaxKmh,
                            ),
                          )
                        else if (_isLoadingSpeed)
                          const Padding(
                            padding: EdgeInsets.fromLTRB(16, 0, 16, 16),
                            child: SizedBox(
                              height: 40,
                              child: Center(
                                child: SizedBox(
                                  width: 24,
                                  height: 24,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                ),
                              ),
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
            if (widget.isMapExpanded && _hasInteractiveMap && points != null)
              Positioned.fill(
                child: _buildExpandedMapOverlay(points),
              ),
            if (showPhotoViewer)
              Positioned.fill(
                child: WorkoutPhotoViewer(
                  workout: widget.workout,
                  authToken: widget.authToken,
                  initialIndex: widget.photoViewerIndex!.clamp(
                    0,
                    widget.workout.mediaFiles.length - 1,
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}
