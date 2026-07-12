import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:grom/api_request.dart';
import 'package:grom/platform/strava_archive_picker.dart';
import 'package:grom/platform/strava_archive_upload.dart';

import '../auth_storage.dart';

class StravaImportState {
  const StravaImportState({
    this.active = false,
    this.phase = 'idle',
    this.uploadProgress = 0,
    this.importProgress = 0,
    this.importCurrent = 0,
    this.importTotal = 0,
    this.message = '',
    this.resultImported = 0,
    this.resultSkipped = 0,
    this.resultParseSkipped = 0,
    this.resultErrors = 0,
    this.completed = false,
    this.failed = false,
  });

  final bool active;
  final String phase;
  final double uploadProgress;
  final double importProgress;
  final int importCurrent;
  final int importTotal;
  final String message;
  final int resultImported;
  final int resultSkipped;
  final int resultParseSkipped;
  final int resultErrors;
  final bool completed;
  final bool failed;

  bool get showUploadProgress =>
      active && (phase == 'uploading' || uploadProgress < 1.0);

  bool get showImportProgress =>
      active && (phase == 'extracting' || phase == 'importing');

  StravaImportState copyWith({
    bool? active,
    String? phase,
    double? uploadProgress,
    double? importProgress,
    int? importCurrent,
    int? importTotal,
    String? message,
    int? resultImported,
    int? resultSkipped,
    int? resultParseSkipped,
    int? resultErrors,
    bool? completed,
    bool? failed,
  }) {
    return StravaImportState(
      active: active ?? this.active,
      phase: phase ?? this.phase,
      uploadProgress: uploadProgress ?? this.uploadProgress,
      importProgress: importProgress ?? this.importProgress,
      importCurrent: importCurrent ?? this.importCurrent,
      importTotal: importTotal ?? this.importTotal,
      message: message ?? this.message,
      resultImported: resultImported ?? this.resultImported,
      resultSkipped: resultSkipped ?? this.resultSkipped,
      resultParseSkipped: resultParseSkipped ?? this.resultParseSkipped,
      resultErrors: resultErrors ?? this.resultErrors,
      completed: completed ?? this.completed,
      failed: failed ?? this.failed,
    );
  }
}

class StravaImportService extends ChangeNotifier {
  StravaImportService._();

  static final StravaImportService instance = StravaImportService._();

  final ApiRequest _api = ApiRequest();
  StravaImportState _state = const StravaImportState();
  Timer? _pollTimer;

  StravaImportState get state => _state;

  Future<void> syncFromServer() async {
    final token = await AuthStorage.getToken();
    if (token == null) {
      return;
    }
    try {
      final status = await _api.getStravaImportStatus(token);
      _applyStatus(status);
    } catch (_) {
      // Keep current state on network errors.
    }
  }

  Future<void> pickAndImport() async {
    if (_state.active) {
      return;
    }

    final pick = await pickStravaArchiveFile();
    if (pick == null) {
      return;
    }

    final token = await AuthStorage.getToken();
    if (token == null) {
      return;
    }

    _state = const StravaImportState(
      active: true,
      phase: 'uploading',
      uploadProgress: 0,
    );
    notifyListeners();

    try {
      await uploadStravaArchivePick(
        api: _api,
        token: token,
        pick: pick,
        onProgress: (progress) {
          _state = _state.copyWith(uploadProgress: progress);
          notifyListeners();
        },
      );
      _state = _state.copyWith(
        uploadProgress: 1,
        phase: 'extracting',
      );
      notifyListeners();
      _startPolling(token);
    } on ApiException catch (e) {
      _state = StravaImportState(
        failed: true,
        message: e.message,
      );
      notifyListeners();
    } catch (e) {
      _state = StravaImportState(
        failed: true,
        message: e.toString(),
      );
      notifyListeners();
    }
  }

  void _startPolling(String token) {
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(const Duration(seconds: 1), (_) async {
      try {
        final status = await _api.getStravaImportStatus(token);
        _applyStatus(status);
      } catch (_) {
        // Ignore transient polling errors.
      }
    });
  }

  void _applyStatus(Map<String, dynamic> status) {
    final active = status['active'] as bool? ?? false;
    final phase = status['phase'] as String? ?? 'idle';
    final uploadProgress =
        (status['upload_progress'] as num?)?.toDouble() ?? 0;
    final importProgress =
        (status['import_progress'] as num?)?.toDouble() ?? 0;
    final importCurrent = status['import_current'] as int? ?? 0;
    final importTotal = status['import_total'] as int? ?? 0;
    final message = status['message'] as String? ?? '';

    var resultImported = 0;
    var resultSkipped = 0;
    var resultParseSkipped = 0;
    var resultErrors = 0;
    final result = status['result'];
    if (result is Map<String, dynamic>) {
      resultImported = result['imported'] as int? ?? 0;
      resultSkipped = result['skipped'] as int? ?? 0;
      resultParseSkipped = result['parse_skipped'] as int? ?? 0;
      resultErrors = result['errors'] as int? ?? 0;
    }

    final completed = phase == 'completed';
    final failed = phase == 'failed';

    _state = StravaImportState(
      active: active,
      phase: phase,
      uploadProgress: uploadProgress,
      importProgress: importProgress,
      importCurrent: importCurrent,
      importTotal: importTotal,
      message: message,
      resultImported: resultImported,
      resultSkipped: resultSkipped,
      resultParseSkipped: resultParseSkipped,
      resultErrors: resultErrors,
      completed: completed,
      failed: failed,
    );
    notifyListeners();

    if (!active || completed || failed) {
      _pollTimer?.cancel();
      _pollTimer = null;
      if (completed || failed) {
        Future<void>.delayed(const Duration(seconds: 4), () {
          if (_state.completed || _state.failed) {
            _state = const StravaImportState();
            notifyListeners();
          }
        });
      }
    }
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }
}
