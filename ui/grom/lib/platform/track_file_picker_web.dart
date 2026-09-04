import 'dart:async';
import 'dart:js_interop';

import 'package:web/web.dart';

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
  final completer = Completer<List<TrackPickResult>>();
  var completed = false;
  late final JSFunction onWindowFocusJs;

  final input = HTMLInputElement()
    ..type = 'file'
    ..accept = '.gpx,.fit,application/gpx+xml,application/xml'
    ..multiple = allowMultiple
    ..style.display = 'none';

  void complete(List<TrackPickResult> value) {
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
        complete(const []);
      }
    });
  }

  onWindowFocusJs = onWindowFocus.toJS;

  void onChange(Event event) {
    final files = input.files;
    if (files == null || files.length == 0) {
      complete(const []);
      return;
    }

    final pending = <Future<TrackPickResult?>>[];
    for (var i = 0; i < files.length; i++) {
      final file = files.item(i);
      if (file == null) {
        continue;
      }
      pending.add(_readWebFile(file));
    }

    Future.wait(pending).then((results) {
      complete([
        for (final item in results)
          if (item != null) item,
      ]);
    });
  }

  void onCancel(Event _) {
    Future<void>.delayed(const Duration(milliseconds: 500), () {
      if (!completed) {
        complete(const []);
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

Future<TrackPickResult?> _readWebFile(File file) {
  final completer = Completer<TrackPickResult?>();
  final reader = FileReader();
  reader.addEventListener(
    'loadend',
    (Event _) {
      final buffer = reader.result;
      if (buffer != null && buffer.isA<JSArrayBuffer>()) {
        final bytes = (buffer as JSArrayBuffer).toDart.asUint8List();
        completer.complete((filename: file.name, bytes: bytes));
      } else {
        completer.complete(null);
      }
    }.toJS,
  );
  reader.readAsArrayBuffer(file);
  return completer.future;
}
