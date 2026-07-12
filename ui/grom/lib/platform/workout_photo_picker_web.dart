import 'dart:async';
import 'dart:js_interop';

import 'package:web/web.dart';

import 'workout_photo_picker_stub.dart';

Future<List<WorkoutPhotoPick>> pickWorkoutPhotos() async {
  final completer = Completer<List<WorkoutPhotoPick>>();
  var completed = false;
  late final JSFunction onWindowFocusJs;

  final input = HTMLInputElement()
    ..type = 'file'
    ..accept = 'image/*'
    ..multiple = true
    ..style.display = 'none';

  void complete(List<WorkoutPhotoPick> value) {
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

  Future<WorkoutPhotoPick?> readFile(File file) {
    final itemCompleter = Completer<WorkoutPhotoPick?>();
    final reader = FileReader();
    reader.addEventListener(
      'loadend',
      (Event _) {
        final buffer = reader.result;
        if (buffer == null || !buffer.isA<JSArrayBuffer>()) {
          itemCompleter.complete(null);
          return;
        }
        final bytes = (buffer as JSArrayBuffer).toDart.asUint8List();
        final filename = file.name.isNotEmpty ? file.name : 'photo.png';
        itemCompleter.complete((filename: filename, bytes: bytes));
      }.toJS,
    );
    reader.readAsArrayBuffer(file);
    return itemCompleter.future;
  }

  void onChange(Event _) async {
    final files = input.files;
    if (files == null || files.length == 0) {
      complete(const []);
      return;
    }

    final results = <WorkoutPhotoPick>[];
    for (var i = 0; i < files.length; i++) {
      final file = files.item(i);
      if (file == null) {
        continue;
      }
      final item = await readFile(file);
      if (item != null) {
        results.add(item);
      }
    }
    complete(results);
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
