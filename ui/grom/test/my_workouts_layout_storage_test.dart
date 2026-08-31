import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/my_workouts_layout.dart';
import 'package:grom/services/my_workouts_layout_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  test('getLayout defaults to cards', () async {
    final layout = await MyWorkoutsLayoutStorage.getLayout();
    expect(layout, MyWorkoutsLayout.cards);
  });

  test('setLayout persists list and cards', () async {
    await MyWorkoutsLayoutStorage.setLayout(MyWorkoutsLayout.list);
    expect(await MyWorkoutsLayoutStorage.getLayout(), MyWorkoutsLayout.list);

    await MyWorkoutsLayoutStorage.setLayout(MyWorkoutsLayout.cards);
    expect(await MyWorkoutsLayoutStorage.getLayout(), MyWorkoutsLayout.cards);
  });

  test('getLayout honors custom defaultValue', () async {
    final layout = await MyWorkoutsLayoutStorage.getLayout(
      defaultValue: MyWorkoutsLayout.list,
    );
    expect(layout, MyWorkoutsLayout.list);
  });
}
