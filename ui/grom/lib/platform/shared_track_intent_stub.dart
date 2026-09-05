import 'dart:typed_data';

typedef SharedTrackPayload = ({String filename, Uint8List bytes});

typedef SharedTrackReceiveResult = ({
  SharedTrackPayload? payload,
  bool readFailed,
  bool unsupportedFormat,
});

Future<SharedTrackReceiveResult> takePendingSharedTrack() async {
  return (payload: null, readFailed: false, unsupportedFormat: false);
}

Stream<SharedTrackReceiveResult> watchSharedTracks() {
  return const Stream.empty();
}
