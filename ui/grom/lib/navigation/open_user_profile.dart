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

/// Opens another user's profile inside [GromShell] (side nav stays).
void openUserProfile(
  BuildContext context, {
  required String handle,
  required String nickname,
  String? selfNickname,
  bool federationEnabled = false,
}) {
  if (isSelfProfile(
    handle: handle,
    nickname: nickname,
    selfNickname: selfNickname,
  )) {
    return;
  }
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
