import 'package:file_picker/file_picker.dart';

import 'strava_archive_types.dart';

Future<StravaArchivePick?> pickStravaArchiveFile() async {
  final result = await FilePicker.platform.pickFiles(
    type: FileType.custom,
    allowedExtensions: const ['zip'],
    withData: false,
    withReadStream: true,
    allowMultiple: false,
  );
  if (result == null || result.files.isEmpty) {
    return null;
  }

  final file = result.files.first;
  if (file.path == null && file.readStream == null && file.bytes == null) {
    return null;
  }

  return (
    filename: file.name,
    path: file.path,
    stream: file.readStream,
    size: file.size,
    bytes: file.bytes,
    nativeFile: null,
  );
}

Future<StravaArchivePick?> pickStravaArchiveFileImpl() => pickStravaArchiveFile();
