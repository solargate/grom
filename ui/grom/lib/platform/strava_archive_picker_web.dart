import 'dart:async';
import 'dart:js_interop';

import 'package:web/web.dart';

import 'strava_archive_types.dart';

Future<StravaArchivePick?> pickStravaArchiveFile() async {
  final completer = Completer<StravaArchivePick?>();
  var completed = false;
  late final JSFunction onWindowFocusJs;

  final input = HTMLInputElement()
    ..type = 'file'
    ..accept = '.zip,application/zip'
    ..multiple = false
    ..style.display = 'none';

  void complete(StravaArchivePick? value) {
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

  void onChange(Event _) {
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

    complete((
      filename: file.name.isNotEmpty ? file.name : 'strava-export.zip',
      path: null,
      stream: null,
      size: file.size,
      bytes: null,
      nativeFile: file,
    ));
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

Future<StravaArchivePick?> pickStravaArchiveFileImpl() => pickStravaArchiveFile();
