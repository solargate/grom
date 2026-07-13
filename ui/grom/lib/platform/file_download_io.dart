import 'dart:io';

import 'package:path_provider/path_provider.dart';

Future<void> saveDownloadedFile({
  required List<int> bytes,
  required String filename,
}) async {
  final directory =
      await getDownloadsDirectory() ?? await getApplicationDocumentsDirectory();
  final file = File('${directory.path}/$filename');
  await file.writeAsBytes(bytes, flush: true);
}
