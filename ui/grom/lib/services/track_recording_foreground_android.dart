import 'package:flutter/foundation.dart';
import 'package:flutter_foreground_task/flutter_foreground_task.dart';

import 'track_recording_foreground_platform.dart';
import 'track_recording_task_handler.dart';

class TrackRecordingForegroundAndroid implements TrackRecordingForegroundPlatform {
  @override
  bool get isSupported =>
      !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

  @override
  Future<void> initialize() async {
    if (!isSupported) {
      return;
    }

    FlutterForegroundTask.init(
      androidNotificationOptions: AndroidNotificationOptions(
        channelId: trackRecordingChannelId,
        channelName: 'Grom workout recording',
        channelDescription:
            'Shows while Grom records a workout track in the background.',
        channelImportance: NotificationChannelImportance.LOW,
        priority: NotificationPriority.LOW,
        visibility: NotificationVisibility.VISIBILITY_PUBLIC,
        onlyAlertOnce: true,
      ),
      iosNotificationOptions: const IOSNotificationOptions(
        showNotification: false,
        playSound: false,
      ),
      foregroundTaskOptions: ForegroundTaskOptions(
        eventAction: ForegroundTaskEventAction.repeat(60000),
        autoRunOnBoot: false,
        autoRunOnMyPackageReplaced: false,
        allowWakeLock: true,
        allowWifiLock: false,
      ),
    );
  }

  @override
  Future<bool> ensureNotificationPermission() async {
    if (!isSupported) {
      return true;
    }

    final permission = await FlutterForegroundTask.checkNotificationPermission();
    if (permission == NotificationPermission.granted) {
      return true;
    }
    final result = await FlutterForegroundTask.requestNotificationPermission();
    return result == NotificationPermission.granted;
  }

  @override
  Future<bool> isRunning() {
    if (!isSupported) {
      return Future.value(false);
    }
    return FlutterForegroundTask.isRunningService;
  }

  @override
  Future<ForegroundServiceResult> start({
    required String title,
    required String text,
    required String channelName,
    String? pausedText,
  }) async {
    if (!isSupported) {
      return const ForegroundServiceFailure(error: 'unsupported_platform');
    }

    await FlutterForegroundTask.saveData(key: 'notificationText', value: text);
    if (pausedText != null) {
      await FlutterForegroundTask.saveData(
        key: 'pausedNotificationText',
        value: pausedText,
      );
    }

    final result = await FlutterForegroundTask.startService(
      serviceId: trackRecordingServiceId,
      notificationTitle: title,
      notificationText: text,
      notificationIcon: const NotificationIcon(
        metaDataName: 'com.solargate.grom.service.TRACK_RECORDING_ICON',
      ),
      serviceTypes: const [ForegroundServiceTypes.location],
      callback: trackRecordingStartCallback,
    );

    return _mapResult(result);
  }

  @override
  Future<ForegroundServiceResult> updateNotification({
    required String title,
    required String text,
  }) async {
    if (!isSupported) {
      return const ForegroundServiceFailure(error: 'unsupported_platform');
    }

    final result = await FlutterForegroundTask.updateService(
      notificationTitle: title,
      notificationText: text,
    );
    return _mapResult(result);
  }

  @override
  Future<void> sendCommand(String type) async {
    if (!isSupported) {
      return;
    }
    FlutterForegroundTask.sendDataToTask({'type': type});
  }

  @override
  Future<ForegroundServiceResult> stop() async {
    if (!isSupported) {
      return const ForegroundServiceFailure(error: 'unsupported_platform');
    }

    final result = await FlutterForegroundTask.stopService();
    return _mapResult(result);
  }

  @override
  void registerTaskDataCallback(void Function(Object data) callback) {
    if (!isSupported) {
      return;
    }
    FlutterForegroundTask.addTaskDataCallback(callback);
  }

  ForegroundServiceResult _mapResult(ServiceRequestResult result) {
    return switch (result) {
      ServiceRequestSuccess() => const ForegroundServiceSuccess(),
      ServiceRequestFailure(:final error) => ForegroundServiceFailure(error: error),
    };
  }
}

TrackRecordingForegroundPlatform createTrackRecordingForegroundPlatform() =>
    TrackRecordingForegroundAndroid();
