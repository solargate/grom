import 'dart:async';
import 'dart:js_interop';

import 'package:web/web.dart';

import 'track_file_picker_stub.dart';

Future<TrackPickResult?> pickTrackFile() async {
  final completer = Completer<TrackPickResult?>();
  var completed = false;
  late final JSFunction onWindowFocusJs;

  final input = HTMLInputElement()
    ..type = 'file'
    ..accept = '.gpx,.fit,application/gpx+xml,application/xml'
    ..multiple = false
    ..style.display = 'none';

  void complete(TrackPickResult? value) {
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
        if (buffer != null && buffer.isA<JSArrayBuffer>()) {
          final bytes = (buffer as JSArrayBuffer).toDart.asUint8List();
          complete((filename: file.name, bytes: bytes));
        } else {
          complete(null);
        }
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
