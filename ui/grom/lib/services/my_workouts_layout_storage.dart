import 'package:shared_preferences/shared_preferences.dart';

import '../models/my_workouts_layout.dart';

const myWorkoutsLayoutStorageKey = 'my_workouts_layout';

class MyWorkoutsLayoutStorage {
  static Future<MyWorkoutsLayout> getLayout({
    MyWorkoutsLayout defaultValue = MyWorkoutsLayout.cards,
  }) async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(myWorkoutsLayoutStorageKey);
    return switch (raw) {
      'list' => MyWorkoutsLayout.list,
      'cards' => MyWorkoutsLayout.cards,
      _ => defaultValue,
    };
  }

  static Future<void> setLayout(MyWorkoutsLayout layout) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(
      myWorkoutsLayoutStorageKey,
      layout == MyWorkoutsLayout.list ? 'list' : 'cards',
    );
  }
}
