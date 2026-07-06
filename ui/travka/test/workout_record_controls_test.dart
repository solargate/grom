import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travka/services/track_recording_service.dart';
import 'package:travka/widgets/workout_record_controls.dart';

void main() {
  testWidgets('WorkoutRecordControls shows finish only when paused', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: WorkoutRecordControls(
            state: TrackRecordingState.paused,
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
        home: Scaffold(
          body: WorkoutRecordControls(
            state: TrackRecordingState.recording,
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
}
