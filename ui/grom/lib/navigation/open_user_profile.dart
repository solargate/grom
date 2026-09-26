import 'package:flutter/material.dart';

import '../models/workout.dart';
import '../pages/user_profile_page.dart';
import '../pages/workout_detail_page.dart';

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
  Navigator.of(context, rootNavigator: true).push<void>(
    MaterialPageRoute<void>(
      builder: (_) => UserProfilePage(
        handle: handle,
        viewerNickname: selfNickname,
        federationEnabled: federationEnabled,
      ),
    ),
  );
}

void openWorkoutFromProfile(
  BuildContext context, {
  required Workout workout,
  required String authToken,
  bool federationEnabled = false,
  String? selfNickname,
}) {
  Navigator.of(context).push<void>(
    MaterialPageRoute<void>(
      builder: (_) => Scaffold(
        appBar: AppBar(),
        body: WorkoutDetailView(
          workout: workout,
          authToken: authToken,
          federationEnabled: federationEnabled,
          selfNickname: selfNickname,
        ),
      ),
    ),
  );
}
