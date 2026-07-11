import 'package:flutter_foreground_task/flutter_foreground_task.dart';

import 'track_recording_service.dart';

Future<void> bootstrapTrackRecording() async {
  FlutterForegroundTask.initCommunicationPort();
  await TrackRecordingService.instance.initialize();
}
