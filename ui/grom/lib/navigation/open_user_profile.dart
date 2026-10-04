import 'package:flutter/material.dart';

import '../models/workout.dart';
import 'grom_shell_scope.dart';

bool isSelfProfile({
  required String handle,
  required String nickname,
  String? selfNickname,
}) {
  if (selfNickname == null || selfNickname.isEmpty) {
    return false;
  }
  return nickname == selfNickname || handle == selfNickname;
}

/// Opens a user profile inside [GromShell] (side nav stays).
///
/// Self vs other and "already on own Profile" no-ops are handled by the shell.
void openUserProfile(
  BuildContext context, {
  required String handle,
  required String nickname,
  String? selfNickname,
  bool federationEnabled = false,
}) {
  final scope = GromShellScope.maybeOf(context);
  if (scope == null) {
    return;
  }
  scope.openUserProfile(handle: handle, nickname: nickname);
}

/// Opens a workout detail inside [GromShell] (used from other-user profiles).
void openWorkoutFromProfile(
  BuildContext context, {
  required Workout workout,
  required String authToken,
  bool federationEnabled = false,
  String? selfNickname,
}) {
  final scope = GromShellScope.maybeOf(context);
  if (scope == null) {
    return;
  }
  scope.openWorkout(workout);
}
