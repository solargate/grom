import 'package:flutter/foundation.dart';
import 'package:grom/api_request.dart';

import '../auth_storage.dart';
import 'strava_api_auth.dart';
import 'strava_api_client.dart';
import 'strava_api_constants.dart';
import 'strava_api_gpx.dart';
import 'strava_api_sport_map.dart';
import 'strava_api_storage.dart';

enum StravaApiSyncResultKind {
  imported,
  noNewWorkouts,
  notConnected,
  notEnabled,
  authFailed,
  cancelled,
  error,
}

class StravaApiSyncResult {
  const StravaApiSyncResult({
    required this.kind,
    this.importedCount = 0,
    this.message = '',
  });

  final StravaApiSyncResultKind kind;
  final int importedCount;
  final String message;
}

/// SnackBar text mapping (mirrors former Health Sync helper).
String stravaApiSyncResultSnackBarMessage(
  StravaApiSyncResult result, {
  required String Function(int count) imported,
  required String noNewWorkouts,
  required String notConnected,
  required String notEnabled,
  required String authFailed,
  required String cancelled,
  required String Function(String message) syncError,
}) {
  switch (result.kind) {
    case StravaApiSyncResultKind.imported:
      return imported(result.importedCount);
    case StravaApiSyncResultKind.noNewWorkouts:
      return noNewWorkouts;
    case StravaApiSyncResultKind.notConnected:
      return notConnected;
    case StravaApiSyncResultKind.notEnabled:
      return notEnabled;
    case StravaApiSyncResultKind.authFailed:
      return authFailed;
    case StravaApiSyncResultKind.cancelled:
      return cancelled;
    case StravaApiSyncResultKind.error:
      return syncError(result.message);
  }
}

class StravaApiSyncService extends ChangeNotifier {
  StravaApiSyncService._({
    ApiRequest? api,
    StravaApiAuth? auth,
    StravaApiClient? client,
    Future<String?> Function()? tokenProvider,
  })  : _api = api ?? ApiRequest(),
        _auth = auth ?? StravaApiAuth(),
        _client = client ?? StravaApiClient(),
        _gromTokenProvider = tokenProvider ?? AuthStorage.getToken;

  static final StravaApiSyncService instance = StravaApiSyncService._();

  @visibleForTesting
  factory StravaApiSyncService.forTesting({
    required ApiRequest api,
    required StravaApiAuth auth,
    required StravaApiClient client,
    required Future<String?> Function() tokenProvider,
  }) {
    return StravaApiSyncService._(
      api: api,
      auth: auth,
      client: client,
      tokenProvider: tokenProvider,
    );
  }

  final ApiRequest _api;
  final StravaApiAuth _auth;
  final StravaApiClient _client;
  final Future<String?> Function() _gromTokenProvider;

  bool _enabled = false;
  bool _connected = false;
  bool _syncing = false;
  String _clientId = '';
  String _clientSecret = '';
  int _syncLimit = kStravaApiSyncLimitDefault;

  bool get enabled => _enabled;
  bool get connected => _connected;
  bool get syncing => _syncing;
  String get clientId => _clientId;
  String get clientSecret => _clientSecret;
  int get syncLimit => _syncLimit;

  /// Home sync button: Android-only callers also gate on platform.
  bool get canSync => _enabled && _connected && !_syncing;

  Future<void> loadFromStorage() async {
    _enabled = await StravaApiStorage.loadEnabled();
    _clientId = await StravaApiStorage.loadClientId();
    _clientSecret = await StravaApiStorage.loadClientSecret();
    _syncLimit = await StravaApiStorage.loadSyncLimit();
    final tokens = await StravaApiStorage.loadTokens();
    _connected = tokens.connected && tokens.refreshToken.isNotEmpty;
    notifyListeners();
  }

  Future<void> setEnabled(bool value) async {
    _enabled = value;
    await StravaApiStorage.saveEnabled(value);
    notifyListeners();
  }

  Future<void> saveCredentialsDraft({
    required String clientId,
    required String clientSecret,
  }) async {
    _clientId = clientId.trim();
    _clientSecret = clientSecret.trim();
    await StravaApiStorage.saveCredentials(
      clientId: _clientId,
      clientSecret: _clientSecret,
    );
  }

  Future<void> saveSyncLimitDraft(int? raw) async {
    _syncLimit = clampStravaApiSyncLimit(raw);
    await StravaApiStorage.saveSyncLimit(_syncLimit);
  }

  Future<StravaConnectResult> connect() async {
    final result = await _auth.connect(
      clientId: _clientId,
      clientSecret: _clientSecret,
    );
    if (result.kind == StravaConnectResultKind.connected) {
      _connected = true;
      notifyListeners();
    } else if (result.kind == StravaConnectResultKind.missingScope ||
        result.kind == StravaConnectResultKind.denied ||
        result.kind == StravaConnectResultKind.error) {
      _connected = false;
      notifyListeners();
    }
    return result;
  }

  Future<StravaApiSyncResult> syncWorkouts() async {
    if (_syncing) {
      return const StravaApiSyncResult(kind: StravaApiSyncResultKind.error);
    }
    if (!_enabled) {
      return const StravaApiSyncResult(kind: StravaApiSyncResultKind.notEnabled);
    }
    if (!_connected) {
      return const StravaApiSyncResult(
        kind: StravaApiSyncResultKind.notConnected,
      );
    }

    _syncing = true;
    notifyListeners();

    try {
      final gromToken = await _gromTokenProvider();
      if (gromToken == null || gromToken.isEmpty) {
        return const StravaApiSyncResult(
          kind: StravaApiSyncResultKind.authFailed,
          message: 'not authenticated',
        );
      }

      final accessToken = await _auth.ensureAccessToken();
      if (accessToken == null || accessToken.isEmpty) {
        _connected = false;
        notifyListeners();
        return const StravaApiSyncResult(
          kind: StravaApiSyncResultKind.notConnected,
        );
      }

      _syncLimit = await StravaApiStorage.loadSyncLimit();

      final activities = await _client.listAthleteActivities(
        accessToken: accessToken,
        perPage: _syncLimit,
        page: 1,
      );

      var imported = 0;
      for (final summary in activities) {
        final externalId = '${summary.id}';
        final exists = await _api.hasExternalID(
          token: gromToken,
          name: kStravaExternalIDName,
          id: externalId,
        );
        if (exists) {
          // Newest-first: stop at first already-imported Strava workout.
          break;
        }

        try {
          await _importOne(
            gromToken: gromToken,
            accessToken: accessToken,
            summary: summary,
          );
          imported++;
        } catch (_) {
          // Continue with remaining newer activities; stop only on existing id.
        }
      }

      if (imported == 0) {
        return const StravaApiSyncResult(
          kind: StravaApiSyncResultKind.noNewWorkouts,
        );
      }
      return StravaApiSyncResult(
        kind: StravaApiSyncResultKind.imported,
        importedCount: imported,
      );
    } on StravaApiException catch (error) {
      if (error.statusCode == 401 || error.statusCode == 403) {
        return StravaApiSyncResult(
          kind: StravaApiSyncResultKind.authFailed,
          message: error.message,
        );
      }
      return StravaApiSyncResult(
        kind: StravaApiSyncResultKind.error,
        message: error.message,
      );
    } on ApiException catch (error) {
      return StravaApiSyncResult(
        kind: StravaApiSyncResultKind.error,
        message: error.message,
      );
    } catch (error) {
      return StravaApiSyncResult(
        kind: StravaApiSyncResultKind.error,
        message: error.toString(),
      );
    } finally {
      _syncing = false;
      notifyListeners();
    }
  }

  Future<void> _importOne({
    required String gromToken,
    required String accessToken,
    required StravaSummaryActivity summary,
  }) async {
    // Prefer detail for description / photo counts when available.
    var activity = summary;
    try {
      activity = await _client.getActivity(
        accessToken: accessToken,
        activityId: summary.id,
      );
    } catch (_) {
      // Keep list summary.
    }

    final sportType = mapStravaSportType(
      activity.sportType.isNotEmpty ? activity.sportType : activity.type,
      activityName: activity.name,
    );
    final name = activity.name.isNotEmpty ? activity.name : sportType;

    List<int>? trackBytes;
    String? trackFilename;
    try {
      final points = await _client.getActivityTrackPoints(
        accessToken: accessToken,
        activityId: activity.id,
      );
      if (points.length >= 2) {
        trackBytes = buildGpxFromStravaStreams(
          name: name,
          startDate: activity.startDate,
          points: points,
        );
        trackFilename = 'strava_${activity.id}.gpx';
      }
    } catch (_) {
      // Import without track.
    }

    final photos = <({String filename, List<int> bytes})>[];
    if (activity.totalPhotoCount > 0) {
      final remotePhotos = await _client.listActivityPhotos(
        accessToken: accessToken,
        activityId: activity.id,
      );
      for (var i = 0; i < remotePhotos.length; i++) {
        final bytes = await _client.downloadBytes(remotePhotos[i].url);
        if (bytes == null || bytes.isEmpty) {
          continue;
        }
        final ext = _guessImageExtension(remotePhotos[i].url);
        photos.add((
          filename: 'strava_${activity.id}_${i + 1}$ext',
          bytes: bytes,
        ));
      }
    }

    final fields = <String, String>{
      'name': name,
      'sport_type': sportType,
      'start_date': activity.startDate.toUtc().toIso8601String(),
      'duration_seconds': '${activity.movingTime}',
      'duration_total_seconds': '${activity.elapsedTime}',
      'distance': '${activity.distanceMeters}',
      'external_id_name': kStravaExternalIDName,
      'external_id_id': '${activity.id}',
      if (activity.description != null && activity.description!.isNotEmpty)
        'description': activity.description!,
      if (activity.maxSpeedMs != null && activity.maxSpeedMs! > 0)
        'speed_max_kmh': (activity.maxSpeedMs! * 3.6).toStringAsFixed(2),
      if (activity.averageSpeedMs != null && activity.averageSpeedMs! > 0)
        'speed_avg_kmh': (activity.averageSpeedMs! * 3.6).toStringAsFixed(2),
    };

    await _api.createWorkoutMultipart(
      token: gromToken,
      fields: fields,
      trackBytes: trackBytes,
      trackFilename: trackFilename,
      photos: photos.isEmpty ? null : photos,
    );
  }

  String _guessImageExtension(String url) {
    final path = Uri.tryParse(url)?.path.toLowerCase() ?? '';
    if (path.endsWith('.png')) {
      return '.png';
    }
    if (path.endsWith('.webp')) {
      return '.webp';
    }
    if (path.endsWith('.gif')) {
      return '.gif';
    }
    return '.jpg';
  }
}
