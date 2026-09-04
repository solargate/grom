import 'package:flutter/foundation.dart';
import 'package:grom/api_request.dart';
import 'package:grom/models/parsed_track_metadata.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/platform/track_file_picker.dart';
import 'package:grom/services/device_track_import_id.dart';

import '../auth_storage.dart';

class DeviceTrackImportState {
  const DeviceTrackImportState({
    this.active = false,
    this.importCurrent = 0,
    this.importTotal = 0,
    this.created = 0,
    this.skipped = 0,
    this.invalid = 0,
    this.failed = 0,
    this.completed = false,
  });

  final bool active;
  final int importCurrent;
  final int importTotal;
  final int created;
  final int skipped;
  final int invalid;
  final int failed;
  final bool completed;

  double get importProgress {
    if (importTotal <= 0) {
      return 0;
    }
    return importCurrent / importTotal;
  }

  bool get showImportProgress => active && importTotal > 0;

  DeviceTrackImportState copyWith({
    bool? active,
    int? importCurrent,
    int? importTotal,
    int? created,
    int? skipped,
    int? invalid,
    int? failed,
    bool? completed,
  }) {
    return DeviceTrackImportState(
      active: active ?? this.active,
      importCurrent: importCurrent ?? this.importCurrent,
      importTotal: importTotal ?? this.importTotal,
      created: created ?? this.created,
      skipped: skipped ?? this.skipped,
      invalid: invalid ?? this.invalid,
      failed: failed ?? this.failed,
      completed: completed ?? this.completed,
    );
  }
}

class DeviceTrackImportService extends ChangeNotifier {
  DeviceTrackImportService._({
    ApiRequest? api,
    Future<String?> Function()? tokenProvider,
  })  : _api = api ?? ApiRequest(),
        _tokenProvider = tokenProvider ?? AuthStorage.getToken;

  static final DeviceTrackImportService instance = DeviceTrackImportService._();

  /// Test-only constructor with injectable API and auth token.
  @visibleForTesting
  factory DeviceTrackImportService.forTesting({
    required ApiRequest api,
    required Future<String?> Function() tokenProvider,
  }) {
    return DeviceTrackImportService._(api: api, tokenProvider: tokenProvider);
  }

  final ApiRequest _api;
  final Future<String?> Function() _tokenProvider;

  DeviceTrackImportState _state = const DeviceTrackImportState();
  DeviceTrackImportState get state => _state;

  Future<DeviceTrackImportState> pickAndImport() async {
    if (_state.active) {
      return _state;
    }

    final files = await pickTrackFiles();
    if (files.isEmpty) {
      return _state;
    }
    return importPickedFiles(files);
  }

  /// Imports already-picked track files (also used by tests).
  Future<DeviceTrackImportState> importPickedFiles(
    List<TrackPickResult> files,
  ) async {
    if (_state.active) {
      return _state;
    }

    final token = await _tokenProvider();
    if (token == null || files.isEmpty) {
      return _state;
    }

    var fallbackSport = defaultSportTypeId;
    try {
      final profile = await _api.getProfile(token);
      fallbackSport = resolveDefaultSportTypeId(profile.lastSportType);
    } catch (_) {
      // Keep defaultSportTypeId.
    }

    _state = DeviceTrackImportState(
      active: true,
      importTotal: files.length,
    );
    notifyListeners();

    var created = 0;
    var skipped = 0;
    var invalid = 0;
    var failed = 0;

    for (var i = 0; i < files.length; i++) {
      final file = files[i];
      final outcome = await _importOne(
        token: token,
        filename: file.filename,
        bytes: file.bytes,
        fallbackSport: fallbackSport,
      );
      switch (outcome) {
        case _ImportOutcome.created:
          created++;
        case _ImportOutcome.skipped:
          skipped++;
        case _ImportOutcome.invalid:
          invalid++;
        case _ImportOutcome.failed:
          failed++;
      }
      _state = _state.copyWith(
        importCurrent: i + 1,
        created: created,
        skipped: skipped,
        invalid: invalid,
        failed: failed,
      );
      notifyListeners();
    }

    _state = _state.copyWith(active: false, completed: true);
    notifyListeners();
    return _state;
  }

  Future<_ImportOutcome> _importOne({
    required String token,
    required String filename,
    required List<int> bytes,
    required String fallbackSport,
  }) async {
    if (!isTrackFilename(filename) || bytes.isEmpty) {
      return _ImportOutcome.invalid;
    }

    final externalId = deviceImportExternalID(filename: filename, bytes: bytes);
    try {
      final exists = await _api.hasExternalID(
        token: token,
        name: deviceImportExternalIDName,
        id: externalId,
      );
      if (exists) {
        return _ImportOutcome.skipped;
      }

      final ParsedTrackMetadata metadata;
      try {
        metadata = await _api.parseTrack(
          token: token,
          trackBytes: bytes,
          trackFilename: filename,
        );
      } catch (_) {
        return _ImportOutcome.failed;
      }

      final sportType = _resolveSportType(metadata.sportType, fallbackSport);
      final name = (metadata.name != null && metadata.name!.isNotEmpty)
          ? metadata.name!
          : trackBasenameWithoutExtension(filename);

      final fields = <String, String>{
        'name': name,
        'sport_type': sportType,
        'external_id_name': deviceImportExternalIDName,
        'external_id_id': externalId,
        if (metadata.startDate != null)
          'start_date': metadata.startDate!.toUtc().toIso8601String(),
        'duration_seconds': (metadata.durationSeconds ?? 0).toString(),
        if (metadata.durationTotalSeconds != null)
          'duration_total_seconds': metadata.durationTotalSeconds!.toString(),
        if (metadata.distanceMeters != null)
          'distance': metadata.distanceMeters!.toString(),
        if (metadata.speedMaxKmh != null)
          'speed_max_kmh': metadata.speedMaxKmh!.toStringAsFixed(2),
        if (metadata.speedAvgKmh != null)
          'speed_avg_kmh': metadata.speedAvgKmh!.toStringAsFixed(2),
      };

      // Form binding requires start_date; track attach fills duration/distance when absent.
      if (!fields.containsKey('start_date')) {
        return _ImportOutcome.failed;
      }

      await _api.createWorkoutMultipart(
        token: token,
        fields: fields,
        trackBytes: bytes,
        trackFilename: filename,
      );
      return _ImportOutcome.created;
    } catch (_) {
      return _ImportOutcome.failed;
    }
  }

  String _resolveSportType(String? fromTrack, String fallback) {
    final trimmed = fromTrack?.trim() ?? '';
    if (trimmed.isNotEmpty && sportTypeById(trimmed) != null) {
      return trimmed;
    }
    return resolveDefaultSportTypeId(fallback);
  }
}

enum _ImportOutcome { created, skipped, invalid, failed }
