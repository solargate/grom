import 'dart:async';
import 'dart:js_interop';

import 'package:web/web.dart';

import 'avatar_image_picker_stub.dart';

Future<AvatarPickResult?> pickAvatarImage() async {
  final completer = Completer<AvatarPickResult?>();
  var completed = false;
  late final JSFunction onWindowFocusJs;

  final input = HTMLInputElement()
    ..type = 'file'
    ..accept = 'image/*'
    ..multiple = false
    ..style.display = 'none';

  void complete(AvatarPickResult? value) {
    if (completed) {
      return;
    }
    completed = true;
    input.remove();
    window.removeEventListener('focus', onWindowFocusJs);
    if (!completer.isCompleted) {
      completer.complete(value);
    }
  }

  void onWindowFocus(Event _) {
    Future<void>.delayed(const Duration(milliseconds: 500), () {
      if (!completed) {
        complete(null);
      }
    });
  }

  onWindowFocusJs = onWindowFocus.toJS;

  void onChange(Event event) {
    final files = input.files;
    if (files == null || files.length == 0) {
      complete(null);
      return;
    }

    final file = files.item(0);
    if (file == null) {
      complete(null);
      return;
    }

    final reader = FileReader();
    reader.addEventListener(
      'loadend',
      (Event _) {
        final buffer = reader.result;
        if (buffer == null || !buffer.isA<JSArrayBuffer>()) {
          complete(null);
          return;
        }

        final bytes = (buffer as JSArrayBuffer).toDart.asUint8List();
        final mimeType = file.type.isNotEmpty ? file.type : 'image/png';
        final blob = Blob(
          <JSUint8Array>[bytes.toJS].toJS,
          BlobPropertyBag(type: mimeType),
        );
        final path = URL.createObjectURL(blob);
        complete((path: path, bytes: bytes));
      }.toJS,
    );
    reader.readAsArrayBuffer(file);
  }

  void onCancel(Event _) {
    Future<void>.delayed(const Duration(milliseconds: 500), () {
      if (!completed) {
        complete(null);
      }
    });
  }

  input.addEventListener('change', onChange.toJS);
  input.addEventListener('cancel', onCancel.toJS);
  window.addEventListener('focus', onWindowFocusJs);

  document.body?.append(input);
  input.click();

  return completer.future;
}
