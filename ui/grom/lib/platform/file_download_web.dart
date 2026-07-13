import 'dart:js_interop';
import 'dart:typed_data';

import 'package:web/web.dart' as web;

Future<void> saveDownloadedFile({
  required List<int> bytes,
  required String filename,
}) async {
  final data = Uint8List.fromList(bytes);
  final blobParts = <web.BlobPart>[data.toJS].toJS;
  final blob = web.Blob(blobParts);
  final url = web.URL.createObjectURL(blob);
  final anchor = web.HTMLAnchorElement()
    ..href = url
    ..download = filename;
  web.document.body?.append(anchor);
  anchor.click();
  anchor.remove();
  web.URL.revokeObjectURL(url);
}
