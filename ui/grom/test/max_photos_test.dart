import 'package:flutter_test/flutter_test.dart';
import 'package:grom/widgets/manual_workout_form.dart';

void main() {
  test('kMaxPhotosPerWorkout matches server MaxPhotosPerWorkout', () {
    expect(kMaxPhotosPerWorkout, 20);
  });
}
