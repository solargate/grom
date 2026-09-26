import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/workout.dart';
import 'package:grom/navigation/grom_shell_scope.dart';
import 'package:grom/navigation/open_user_profile.dart';

void main() {
  group('isSelfProfile', () {
    test('matches nickname or handle to self', () {
      expect(
        isSelfProfile(
          handle: 'alice',
          nickname: 'alice',
          selfNickname: 'alice',
        ),
        isTrue,
      );
      expect(
        isSelfProfile(
          handle: 'alice@example.com',
          nickname: 'alice',
          selfNickname: 'alice',
        ),
        isTrue,
      );
      expect(
        isSelfProfile(
          handle: 'bob',
          nickname: 'bob',
          selfNickname: 'alice',
        ),
        isFalse,
      );
      expect(
        isSelfProfile(
          handle: 'bob',
          nickname: 'bob',
          selfNickname: null,
        ),
        isFalse,
      );
    });
  });

  testWidgets('openUserProfile uses GromShellScope', (tester) async {
    String? openedHandle;
    String? openedNickname;

    await tester.pumpWidget(
      MaterialApp(
        home: GromShellScope(
          openUserProfile: ({required handle, required nickname}) {
            openedHandle = handle;
            openedNickname = nickname;
          },
          openWorkout: (_) {},
          child: Builder(
            builder: (context) {
              return TextButton(
                onPressed: () {
                  openUserProfile(
                    context,
                    handle: 'bob',
                    nickname: 'bob',
                    selfNickname: 'alice',
                  );
                },
                child: const Text('open'),
              );
            },
          ),
        ),
      ),
    );

    await tester.tap(find.text('open'));
    expect(openedHandle, 'bob');
    expect(openedNickname, 'bob');
  });

  testWidgets('openUserProfile no-ops for self', (tester) async {
    var called = false;

    await tester.pumpWidget(
      MaterialApp(
        home: GromShellScope(
          openUserProfile: ({required handle, required nickname}) {
            called = true;
          },
          openWorkout: (_) {},
          child: Builder(
            builder: (context) {
              return TextButton(
                onPressed: () {
                  openUserProfile(
                    context,
                    handle: 'alice',
                    nickname: 'alice',
                    selfNickname: 'alice',
                  );
                },
                child: const Text('open'),
              );
            },
          ),
        ),
      ),
    );

    await tester.tap(find.text('open'));
    expect(called, isFalse);
  });

  testWidgets('openWorkoutFromProfile uses GromShellScope', (tester) async {
    Workout? opened;

    final workout = Workout(
      id: 'w1',
      name: 'Run',
      description: '',
      sportType: 'running',
      startDate: DateTime.utc(2024, 1, 1),
      durationSeconds: 60,
      distance: 1,
      owner: 'bob',
      device: '',
      track: '',
    );

    await tester.pumpWidget(
      MaterialApp(
        home: GromShellScope(
          openUserProfile: ({required handle, required nickname}) {},
          openWorkout: (w) => opened = w,
          child: Builder(
            builder: (context) {
              return TextButton(
                onPressed: () {
                  openWorkoutFromProfile(
                    context,
                    workout: workout,
                    authToken: 'token',
                  );
                },
                child: const Text('open'),
              );
            },
          ),
        ),
      ),
    );

    await tester.tap(find.text('open'));
    expect(opened?.id, 'w1');
  });
}
