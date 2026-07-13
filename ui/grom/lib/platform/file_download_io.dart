import 'dart:io';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:path_provider/path_provider.dart';

import 'file_download_exceptions.dart';

Future<void> saveDownloadedFile({
  required List<int> bytes,
  required String filename,
}) async {
  if (Platform.isAndroid || Platform.isIOS) {
    final ext = filename.contains('.')
        ? filename.split('.').last.toLowerCase()
        : null;

    final savedPath = await FilePicker.platform.saveFile(
      fileName: filename,
      bytes: Uint8List.fromList(bytes),
      type: ext == null ? FileType.any : FileType.custom,
      allowedExtensions: ext == null ? null : [ext],
    );

    if (savedPath == null) {
      throw SaveDownloadedFileCancelled();
    }
    return;
  }

  final directory =
      await getDownloadsDirectory() ?? await getApplicationDocumentsDirectory();
  final file = File('${directory.path}/$filename');
  await file.writeAsBytes(bytes, flush: true);
}
