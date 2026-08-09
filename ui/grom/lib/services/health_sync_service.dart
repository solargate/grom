import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:grom/api_request.dart';
import 'package:grom/platform/google_drive.dart';
import 'package:grom/services/health_sync_csv.dart';
import 'package:grom/services/health_sync_storage.dart';

import '../auth_storage.dart';

enum HealthSyncResultKind {
  imported,
  noNewWorkouts,
  folderNotFound,
  folderEmpty,
  signInCancelled,
  signInFailed,
  accessDenied,
  error,
}

class HealthSyncResult {
  const HealthSyncResult({
    required this.kind,
    this.importedCount = 0,
    this.message = '',
  });

  final HealthSyncResultKind kind;
  final int importedCount;
  final String message;
}

class HealthSyncService extends ChangeNotifier {
  HealthSyncService._();

  static final HealthSyncService instance = HealthSyncService._();

  final ApiRequest _api = ApiRequest();

  bool _enabled = false;
  String _folderName = '';
  String _folderId = '';
  bool _syncing = false;

  bool get enabled => _enabled;
  String get folderName => _folderName;
  String get folderId => _folderId;
  bool get syncing => _syncing;

  Future<void> loadFromStorage() async {
    _enabled = await HealthSyncStorage.loadEnabled();
    final folder = await HealthSyncStorage.loadFolder();
    _folderName = folder.name;
    _folderId = folder.id;
    notifyListeners();
  }

  Future<void> setEnabled(bool value) async {
    _enabled = value;
    await HealthSyncStorage.saveEnabled(value);
    notifyListeners();
  }

  Future<void> saveFolder({required String name, required String id}) async {
    _folderName = name;
    _folderId = id;
    await HealthSyncStorage.saveFolder(name: name, id: id);
    notifyListeners();
  }

  Future<void> updateFolderName(String name) async {
    _folderName = name;
    _folderId = '';
    await HealthSyncStorage.saveFolderName(name);
    await HealthSyncStorage.clearFolderId();
    notifyListeners();
  }

  Future<HealthSyncResult> enableAndSetup() async {
    try {
      await ensureGoogleDriveSignedIn();
    } on GoogleDriveException catch (error) {
      return _resultFromDriveError(error);
    } catch (error) {
      return HealthSyncResult(
        kind: HealthSyncResultKind.error,
        message: error.toString(),
      );
    }

    if (_folderId.isNotEmpty && _folderName.isNotEmpty) {
      await setEnabled(true);
      return const HealthSyncResult(kind: HealthSyncResultKind.noNewWorkouts);
    }

    try {
      final folder = await findHealthSyncFolderByNameContains();
      if (folder == null) {
        return const HealthSyncResult(kind: HealthSyncResultKind.folderNotFound);
      }
      await saveFolder(name: folder.name, id: folder.id);
      await setEnabled(true);
      return const HealthSyncResult(kind: HealthSyncResultKind.noNewWorkouts);
    } on GoogleDriveException catch (error) {
      return _resultFromDriveError(error);
    } catch (error) {
      return HealthSyncResult(
        kind: HealthSyncResultKind.error,
        message: error.toString(),
      );
    }
  }

  Future<HealthSyncResult> refreshFolderFromDrive() async {
    try {
      await ensureGoogleDriveSignedIn();
      final folder = await findHealthSyncFolderByNameContains();
      if (folder == null) {
        return const HealthSyncResult(kind: HealthSyncResultKind.folderNotFound);
      }
      await saveFolder(name: folder.name, id: folder.id);
      return const HealthSyncResult(kind: HealthSyncResultKind.noNewWorkouts);
    } on GoogleDriveException catch (error) {
      return _resultFromDriveError(error);
    } catch (error) {
      return HealthSyncResult(
        kind: HealthSyncResultKind.error,
        message: error.toString(),
      );
    }
  }

  Future<HealthSyncResult> syncWorkouts() async {
    if (_syncing) {
      return const HealthSyncResult(kind: HealthSyncResultKind.error);
    }

    final folderName = _folderName.trim();
    if (folderName.isEmpty) {
      return const HealthSyncResult(kind: HealthSyncResultKind.folderEmpty);
    }

    _syncing = true;
    notifyListeners();

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        return const HealthSyncResult(
          kind: HealthSyncResultKind.error,
          message: 'not authenticated',
        );
      }

      await ensureGoogleDriveSignedIn();

      final folderId = await _resolveFolderId(folderName);
      if (folderId == null) {
        return const HealthSyncResult(kind: HealthSyncResultKind.folderNotFound);
      }

      final files = await listFolderFiles(folderId);
      if (files.isEmpty) {
        return const HealthSyncResult(kind: HealthSyncResultKind.folderEmpty);
      }

      final filenames = files.map((file) => file.name).toList();
      final csvFiles = files.where((file) => isHealthSyncCsvFilename(file.name));
      var imported = 0;

      for (final csvFile in csvFiles) {
        try {
          final importedOne = await _importCsvFile(
            token: token,
            csvFile: csvFile,
            filenames: filenames,
            files: files,
          );
          if (importedOne) {
            imported++;
          }
        } catch (_) {
          // Skip individual file failures and continue syncing others.
        }
      }

      if (imported == 0) {
        return const HealthSyncResult(kind: HealthSyncResultKind.noNewWorkouts);
      }
      return HealthSyncResult(
        kind: HealthSyncResultKind.imported,
        importedCount: imported,
      );
    } on GoogleDriveException catch (error) {
      return _resultFromDriveError(error);
    } on ApiException catch (error) {
      return HealthSyncResult(
        kind: HealthSyncResultKind.error,
        message: error.message,
      );
    } catch (error) {
      return HealthSyncResult(
        kind: HealthSyncResultKind.error,
        message: error.toString(),
      );
    } finally {
      _syncing = false;
      notifyListeners();
    }
  }

  Future<String?> _resolveFolderId(String folderName) async {
    if (_folderId.isNotEmpty) {
      try {
        await listFolderFiles(_folderId);
        return _folderId;
      } catch (_) {
        _folderId = '';
        await HealthSyncStorage.clearFolderId();
        notifyListeners();
      }
    }

    final folder = await findFolderByExactName(folderName);
    if (folder == null) {
      return null;
    }
    await saveFolder(name: folder.name, id: folder.id);
    return folder.id;
  }

  Future<bool> _importCsvFile({
    required String token,
    required GoogleDriveFileEntry csvFile,
    required List<String> filenames,
    required List<GoogleDriveFileEntry> files,
  }) async {
    final bytes = await downloadDriveFile(csvFile.id);
    final content = utf8.decode(bytes);
    final row = parseHealthSyncCsvContent(content);
    if (row == null) {
      return false;
    }

    final exists = await _api.hasExternalID(
      token: token,
      name: row.externalIdName,
      id: csvFile.name,
    );
    if (exists) {
      return false;
    }

    final trackFilename = matchHealthSyncTrackFilename(csvFile.name, filenames);
    List<int>? trackBytes;
    if (trackFilename != null) {
      GoogleDriveFileEntry? trackFile;
      for (final file in files) {
        if (file.name == trackFilename) {
          trackFile = file;
          break;
        }
      }
      if (trackFile != null) {
        trackBytes = await downloadDriveFile(trackFile.id);
      }
    }

    final fields = <String, String>{
      'name': row.name.isNotEmpty ? row.name : row.sport,
      'sport_type': row.sportTypeId,
      'start_date': row.startDate.toUtc().toIso8601String(),
      'duration_seconds': row.durationMovingSeconds.toString(),
      'duration_total_seconds': row.durationTotalSeconds.toString(),
      'distance': (row.distanceKm * 1000).toString(),
      'external_id_name': row.externalIdName,
      'external_id_id': csvFile.name,
    };

    await _api.createWorkoutMultipart(
      token: token,
      fields: fields,
      trackBytes: trackBytes,
      trackFilename: trackFilename,
    );
    return true;
  }

  HealthSyncResult _resultFromDriveError(GoogleDriveException error) {
    return healthSyncResultFromDriveError(error);
  }
}

/// Maps [GoogleDriveException] to a [HealthSyncResult] for UI handling.
HealthSyncResult healthSyncResultFromDriveError(GoogleDriveException error) {
  final detail = error.message ?? '';
  if (error.error == GoogleDriveError.cancelled) {
    return HealthSyncResult(
      kind: HealthSyncResultKind.signInCancelled,
      message: detail,
    );
  }
  if (error.error == GoogleDriveError.signInFailed) {
    return HealthSyncResult(
      kind: HealthSyncResultKind.signInFailed,
      message: detail,
    );
  }
  if (error.error == GoogleDriveError.accessDenied) {
    return HealthSyncResult(
      kind: HealthSyncResultKind.accessDenied,
      message: detail,
    );
  }
  return HealthSyncResult(
    kind: HealthSyncResultKind.error,
    message: detail.isEmpty ? 'unsupported platform' : detail,
  );
}

/// Localized snackbar text for a [HealthSyncResult], with optional error detail.
String healthSyncResultSnackBarMessage(
  HealthSyncResult result, {
  required String imported,
  required String noNewWorkouts,
  required String folderNotFound,
  required String folderEmpty,
  required String signInCancelled,
  required String signInFailed,
  required String accessDenied,
  required String Function(String message) syncError,
}) {
  String withDetail(String base) {
    final detail = result.message.trim();
    if (detail.isEmpty) {
      return base;
    }
    return '$base: $detail';
  }

  return switch (result.kind) {
    HealthSyncResultKind.imported => imported,
    HealthSyncResultKind.noNewWorkouts => noNewWorkouts,
    HealthSyncResultKind.folderNotFound => folderNotFound,
    HealthSyncResultKind.folderEmpty => folderEmpty,
    HealthSyncResultKind.signInCancelled => withDetail(signInCancelled),
    HealthSyncResultKind.signInFailed => withDetail(signInFailed),
    HealthSyncResultKind.accessDenied => withDetail(accessDenied),
    HealthSyncResultKind.error => syncError(result.message),
  };
}
