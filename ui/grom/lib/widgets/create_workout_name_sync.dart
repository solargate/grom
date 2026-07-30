/// Tracks whether the create-workout name field still follows the sport type.
///
/// Edit mode does not use this helper. Once the user edits the name (including
/// clearing it), [synced] stays false for the form session.
class CreateWorkoutNameSync {
  CreateWorkoutNameSync({this.synced = true});

  bool synced;

  /// Returns the sport label to apply, or null if the name is locked.
  String? nameForSportChange(String sportLabel) {
    if (!synced) {
      return null;
    }
    return sportLabel;
  }

  void onUserEdited() {
    synced = false;
  }
}
