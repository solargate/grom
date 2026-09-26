import 'package:flutter/material.dart';

import '../models/workout.dart';

/// Entry on the in-shell other-user profile stack.
class ViewingUser {
  const ViewingUser({
    required this.handle,
    required this.nickname,
  });

  final String handle;
  final String nickname;
}

/// Navigation callbacks for content inside [GromShell].
class GromShellScope extends InheritedWidget {
  const GromShellScope({
    super.key,
    required this.openUserProfile,
    required this.openWorkout,
    required super.child,
  });

  final void Function({
    required String handle,
    required String nickname,
  }) openUserProfile;

  final ValueChanged<Workout> openWorkout;

  static GromShellScope of(BuildContext context) {
    final scope = context.dependOnInheritedWidgetOfExactType<GromShellScope>();
    assert(scope != null, 'GromShellScope not found in context');
    return scope!;
  }

  static GromShellScope? maybeOf(BuildContext context) {
    return context.getInheritedWidgetOfExactType<GromShellScope>();
  }

  @override
  bool updateShouldNotify(GromShellScope oldWidget) => false;
}
