import 'auth_storage.dart';

/// Clears the Grom JWT from local storage.
Future<void> clearLocalSession() async {
  await AuthStorage.clear();
}
