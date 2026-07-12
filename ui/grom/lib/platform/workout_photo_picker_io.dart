import 'package:image_picker/image_picker.dart';

import 'workout_photo_picker_stub.dart';

Future<List<WorkoutPhotoPick>> pickWorkoutPhotos() async {
  final picked = await ImagePicker().pickMultiImage(imageQuality: 100);
  if (picked.isEmpty) {
    return const [];
  }

  final results = <WorkoutPhotoPick>[];
  for (final file in picked) {
    final bytes = await file.readAsBytes();
    final filename = file.name.isNotEmpty ? file.name : 'photo.jpg';
    results.add((filename: filename, bytes: bytes));
  }
  return results;
}
