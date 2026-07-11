import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';

import 'track_file_picker_stub.dart';

Future<TrackPickResult?> pickTrackFile() async {
  final result = await FilePicker.platform.pickFiles(
    type: FileType.custom,
    allowedExtensions: const ['gpx', 'fit'],
    withData: true,
    allowMultiple: false,
  );

  if (result == null || result.files.isEmpty) {
    return null;
  }

  final file = result.files.first;
  final bytes = file.bytes;
  final filename = file.name;
  if (bytes == null || filename.isEmpty) {
    return null;
  }

  return (filename: filename, bytes: Uint8List.fromList(bytes));
}
