import 'dart:io';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';

import 'track_file_picker_stub.dart';

export 'track_file_picker_stub.dart' show TrackPickResult;

Future<TrackPickResult?> pickTrackFile() async {
  final files = await pickTrackFiles(allowMultiple: false);
  if (files.isEmpty) {
    return null;
  }
  return files.first;
}

Future<List<TrackPickResult>> pickTrackFiles({bool allowMultiple = true}) async {
  final result = await FilePicker.platform.pickFiles(
    type: FileType.custom,
    allowedExtensions: const ['gpx', 'fit'],
    withData: false,
    withReadStream: true,
    allowMultiple: allowMultiple,
  );

  if (result == null || result.files.isEmpty) {
    return const [];
  }

  final picked = <TrackPickResult>[];
  for (final file in result.files) {
    final filename = file.name;
    if (filename.isEmpty) {
      continue;
    }
    final bytes = await _readPlatformFileBytes(file);
    if (bytes == null || bytes.isEmpty) {
      continue;
    }
    picked.add((filename: filename, bytes: bytes));
  }
  return picked;
}

Future<Uint8List?> _readPlatformFileBytes(PlatformFile file) async {
  if (file.bytes != null && file.bytes!.isNotEmpty) {
    return Uint8List.fromList(file.bytes!);
  }

  final path = file.path;
  if (path != null && path.isNotEmpty) {
    try {
      return await File(path).readAsBytes();
    } catch (_) {
      // Fall through to stream.
    }
  }

  final stream = file.readStream;
  if (stream != null) {
    final builder = BytesBuilder(copy: false);
    await for (final chunk in stream) {
      builder.add(chunk);
    }
    if (builder.length > 0) {
      return builder.takeBytes();
    }
  }

  return null;
}
