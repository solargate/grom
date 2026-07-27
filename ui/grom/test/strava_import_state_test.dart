import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_import_service.dart';

void main() {
  test('fromStatus parses importing progress', () {
    final state = StravaImportState.fromStatus({
      'active': true,
      'phase': 'importing',
      'upload_progress': 1,
      'import_progress': 0.4,
      'import_current': 4,
      'import_total': 10,
      'message': 'Working',
    });

    expect(state.active, isTrue);
    expect(state.phase, 'importing');
    expect(state.uploadProgress, 1);
    expect(state.importProgress, 0.4);
    expect(state.importCurrent, 4);
    expect(state.importTotal, 10);
    expect(state.message, 'Working');
    expect(state.completed, isFalse);
    expect(state.failed, isFalse);
    expect(state.showImportProgress, isTrue);
    expect(state.showUploadProgress, isFalse);
  });

  test('fromStatus parses completed result counters', () {
    final state = StravaImportState.fromStatus({
      'active': false,
      'phase': 'completed',
      'result': {
        'imported': 12,
        'skipped': 3,
        'parse_skipped': 1,
        'media_missing': 47,
        'errors': 2,
      },
    });

    expect(state.completed, isTrue);
    expect(state.failed, isFalse);
    expect(state.resultImported, 12);
    expect(state.resultSkipped, 3);
    expect(state.resultParseSkipped, 1);
    expect(state.resultMediaMissing, 47);
    expect(state.resultErrors, 2);
  });

  test('fromStatus marks failed phase and defaults missing fields', () {
    final state = StravaImportState.fromStatus({
      'phase': 'failed',
      'message': 'bad zip',
    });

    expect(state.active, isFalse);
    expect(state.failed, isTrue);
    expect(state.completed, isFalse);
    expect(state.message, 'bad zip');
    expect(state.uploadProgress, 0);
    expect(state.importProgress, 0);
    expect(state.resultImported, 0);
  });

  test('showUploadProgress is true while uploading', () {
    final state = StravaImportState.fromStatus({
      'active': true,
      'phase': 'uploading',
      'upload_progress': 0.2,
    });
    expect(state.showUploadProgress, isTrue);
    expect(state.showImportProgress, isFalse);
  });
}
