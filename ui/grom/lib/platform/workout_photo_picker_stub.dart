import 'dart:typed_data';

typedef WorkoutPhotoPick = ({String filename, Uint8List bytes});

Future<List<WorkoutPhotoPick>> pickWorkoutPhotos() =>
    throw UnsupportedError('Cannot pick workout photos without platform implementation');
