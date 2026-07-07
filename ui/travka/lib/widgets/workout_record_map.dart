import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_map_location_marker/flutter_map_location_marker.dart';
import 'package:latlong2/latlong.dart';

class WorkoutRecordMap extends StatefulWidget {
  const WorkoutRecordMap({
    super.key,
    required this.trackPoints,
    required this.followUser,
    this.initialCenter,
    this.showCurrentLocation = true,
    this.fitToTrack = false,
  });

  final List<LatLng> trackPoints;
  final bool followUser;
  final LatLng? initialCenter;
  final bool showCurrentLocation;
  final bool fitToTrack;

  @override
  State<WorkoutRecordMap> createState() => _WorkoutRecordMapState();
}

class _WorkoutRecordMapState extends State<WorkoutRecordMap> {
  final MapController _mapController = MapController();
  LatLng? _lastCentered;

  CameraFit? get _initialCameraFit {
    if (!widget.fitToTrack || widget.trackPoints.length < 2) {
      return null;
    }
    return CameraFit.bounds(
      bounds: LatLngBounds.fromPoints(widget.trackPoints),
      padding: const EdgeInsets.all(32),
    );
  }

  void _refreshTilesAfterCameraChange() {
    if (!widget.fitToTrack) {
      return;
    }
    final camera = _mapController.camera;
    _mapController.move(camera.center, camera.zoom + 0.0001);
  }

  @override
  void didUpdateWidget(covariant WorkoutRecordMap oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.followUser && widget.trackPoints.isNotEmpty) {
      final latest = widget.trackPoints.last;
      if (_lastCentered != latest) {
        _lastCentered = latest;
        _mapController.move(latest, _mapController.camera.zoom);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final center = widget.initialCenter ??
        (widget.trackPoints.isNotEmpty
            ? widget.trackPoints.last
            : const LatLng(0, 0));
    final hasValidCenter = widget.initialCenter != null ||
        widget.trackPoints.isNotEmpty ||
        (center.latitude != 0 || center.longitude != 0);

    if (!hasValidCenter) {
      return const Center(child: CircularProgressIndicator());
    }

    return FlutterMap(
      mapController: _mapController,
      options: MapOptions(
        initialCenter: center,
        initialZoom: 16,
        initialCameraFit: _initialCameraFit,
        onMapReady: _refreshTilesAfterCameraChange,
        interactionOptions: const InteractionOptions(
          flags: InteractiveFlag.all,
        ),
      ),
      children: [
        TileLayer(
          urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
          userAgentPackageName: 'com.example.travka',
          maxZoom: 19,
        ),
        if (widget.trackPoints.length >= 2)
          PolylineLayer(
            polylines: [
              Polyline(
                points: widget.trackPoints,
                strokeWidth: 5,
                color: colorScheme.primary,
                strokeCap: StrokeCap.round,
                strokeJoin: StrokeJoin.round,
              ),
            ],
          ),
        if (widget.showCurrentLocation)
          CurrentLocationLayer(
            alignPositionOnUpdate: widget.followUser
                ? AlignOnUpdate.always
                : AlignOnUpdate.never,
            alignDirectionOnUpdate: AlignOnUpdate.never,
          ),
        RichAttributionWidget(
          attributions: [
            TextSourceAttribution(
              'OpenStreetMap contributors',
              onTap: () {},
            ),
          ],
        ),
      ],
    );
  }
}
