import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/services/track_recording_service.dart';
import 'package:grom/widgets/workout_record_controls.dart';

void main() {
  testWidgets('WorkoutRecordControls shows finish only when paused', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        home: Scaffold(
          body: WorkoutRecordControls(
            state: TrackRecordingState.paused,
            pauseVariant: PauseButtonVariant.none,
            onPlay: () {},
            onPause: () {},
            onFinish: () {},
            playLabel: 'Record',
            pauseLabel: 'Pause',
            finishLabel: 'Finish',
          ),
        ),
      ),
    );

    expect(find.text('Finish'), findsOneWidget);
  });

  testWidgets('WorkoutRecordControls hides finish when recording', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        home: Scaffold(
          body: WorkoutRecordControls(
            state: TrackRecordingState.recording,
            pauseVariant: PauseButtonVariant.manual,
            onPlay: () {},
            onPause: () {},
            onFinish: () {},
            playLabel: 'Record',
            pauseLabel: 'Pause',
            finishLabel: 'Finish',
          ),
        ),
      ),
    );

    expect(find.text('Finish'), findsNothing);
  });

  testWidgets('WorkoutRecordControls uses auto pause color when auto paused', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        home: Scaffold(
          body: WorkoutRecordControls(
            state: TrackRecordingState.autoPaused,
            pauseVariant: PauseButtonVariant.auto,
            onPlay: () {},
            onPause: () {},
            onFinish: () {},
            playLabel: 'Record',
            pauseLabel: 'Pause',
            finishLabel: 'Finish',
          ),
        ),
      ),
    );

    final materials = tester.widgetList<Material>(find.byType(Material));
    expect(materials.any((material) => material.color == kAutoPauseColor), isTrue);
    expect(find.text('Finish'), findsNothing);
  });
}
