import 'dart:convert';
import 'dart:io';

import 'package:path_provider/path_provider.dart';

import 'track_recording_session.dart';
import 'track_recording_store_platform.dart';

class TrackRecordingStoreIo implements TrackRecordingStorePlatform {
  static const _fileName = 'track_recording_session.json';

  TrackRecordingStoreIo({File? file}) : _file = file;

  File? _file;
  TrackRecordingSession? _cachedSession;
  Future<void> _pending = Future<void>.value();

  @override
  Future<void> init() async {
    await load();
  }

  @override
  TrackRecordingSession? get cachedSession => _cachedSession;

  @override
  bool get hasActiveSession => _cachedSession?.isActive ?? false;

  @override
  Future<TrackRecordingSession?> load() => _enqueue(() async {
        _cachedSession = await _readFromDisk();
        return _cachedSession;
      });

  @override
  Future<void> save(TrackRecordingSession session) => _enqueue(() async {
        final file = _file ?? await _resolveFile();
        final temp = File('${file.path}.tmp');
        try {
          await temp.writeAsString(jsonEncode(session.toJson()), flush: true);
          await temp.rename(file.path);
          _cachedSession = session;
        } finally {
          if (await temp.exists()) {
            await temp.delete();
          }
        }
      });

  @override
  Future<void> clear() => _enqueue(() async {
        final file = _file ?? await _resolveFile();
        if (await file.exists()) {
          await file.delete();
        }
        _cachedSession = null;
      });

  Future<T> _enqueue<T>(Future<T> Function() operation) {
    final result = _pending.then((_) => operation());
    _pending = result.then<void>((_) {}, onError: (Object _) {});
    return result;
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
