import 'dart:convert';
import 'dart:io';

import 'package:path_provider/path_provider.dart';

import 'track_recording_session.dart';
import 'track_recording_store_platform.dart';

class TrackRecordingStoreIo implements TrackRecordingStorePlatform {
  static const _fileName = 'track_recording_session.json';

  File? _file;
  TrackRecordingSession? _cachedSession;

  @override
  Future<void> init() async {
    final directory = await getApplicationDocumentsDirectory();
    _file = File('${directory.path}/$_fileName');
    _cachedSession = await _readFromDisk();
  }

  @override
  TrackRecordingSession? get cachedSession => _cachedSession;

  @override
  bool get hasActiveSession => _cachedSession?.isActive ?? false;

  @override
  Future<TrackRecordingSession?> load() async {
    _cachedSession = await _readFromDisk();
    return _cachedSession;
  }

  @override
  Future<void> save(TrackRecordingSession session) async {
    _cachedSession = session;
    final file = _file ?? await _resolveFile();
    await file.writeAsString(
      const JsonEncoder.withIndent('  ').convert(session.toJson()),
    );
  }

  @override
  Future<void> clear() async {
    _cachedSession = null;
    final file = _file ?? await _resolveFile();
    if (await file.exists()) {
      await file.delete();
    }
  }

  Future<File> _resolveFile() async {
    final directory = await getApplicationDocumentsDirectory();
    _file = File('${directory.path}/$_fileName');
    return _file!;
  }

  Future<TrackRecordingSession?> _readFromDisk() async {
    final file = _file ?? await _resolveFile();
    if (!await file.exists()) {
      return null;
    }
    try {
      final raw = await file.readAsString();
      if (raw.isEmpty) {
        return null;
      }
      final json = jsonDecode(raw) as Map<String, dynamic>;
      return TrackRecordingSession.fromJson(json);
    } catch (_) {
      return null;
    }
  }
}

TrackRecordingStorePlatform createTrackRecordingStorePlatform() =>
    TrackRecordingStoreIo();
