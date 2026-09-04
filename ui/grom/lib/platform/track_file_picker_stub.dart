import 'dart:typed_data';

typedef TrackPickResult = ({String filename, Uint8List bytes});

Future<TrackPickResult?> pickTrackFile() {
  throw UnsupportedError('pickTrackFile is not supported on this platform');
}

Future<List<TrackPickResult>> pickTrackFiles() {
  throw UnsupportedError('pickTrackFiles is not supported on this platform');
}
