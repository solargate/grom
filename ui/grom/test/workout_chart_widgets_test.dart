import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/workout_heartrate.dart';
import 'package:grom/models/workout_pace.dart';
import 'package:grom/models/workout_speed.dart';
import 'package:grom/widgets/workout_heartrate_chart.dart';
import 'package:grom/widgets/workout_pace_chart.dart';
import 'package:grom/widgets/workout_speed_chart.dart';

Widget wrap(Widget child) {
  return MaterialApp(
    theme: buildAppTheme(),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(body: child),
  );
}

void main() {
  testWidgets('WorkoutSpeedChart hides when fewer than 2 samples', (tester) async {
    await tester.pumpWidget(
      wrap(
        WorkoutSpeedChart(
          samples: [
            WorkoutSpeedSample(
              time: DateTime.utc(2026, 7, 8, 10),
              speedKmh: 10,
              distanceM: 0,
            ),
          ],
        ),
      ),
    );
    expect(tester.takeException(), isNull);
    expect(find.byType(SizedBox), findsWidgets);
  });

  testWidgets('WorkoutPaceChart hides when fewer than 2 samples', (tester) async {
    await tester.pumpWidget(
      wrap(
        WorkoutPaceChart(
          samples: [
            WorkoutPaceSample(
              time: DateTime.utc(2026, 7, 8, 10),
              paceSecPerKm: 360,
              distanceM: 0,
            ),
          ],
        ),
      ),
    );
    expect(tester.takeException(), isNull);
    expect(find.text('Pace'), findsNothing);
  });

  testWidgets('WorkoutPaceChart renders samples and avg/best rows', (tester) async {
    await tester.pumpWidget(
      wrap(
        WorkoutPaceChart(
          samples: [
            WorkoutPaceSample(
              time: DateTime.utc(2026, 7, 8, 10),
              paceSecPerKm: 360,
              distanceM: 0,
            ),
            WorkoutPaceSample(
              time: DateTime.utc(2026, 7, 8, 10, 1),
              paceSecPerKm: 300,
              distanceM: 200,
            ),
          ],
          paceAvgSec: 330,
          paceBestSec: 300,
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(find.byType(WorkoutPaceChart), findsOneWidget);
    expect(find.text('Pace'), findsOneWidget);
    expect(find.text('Avg. pace'), findsOneWidget);
    expect(find.text('Best pace'), findsOneWidget);
    expect(find.text('5:30 /km'), findsOneWidget);
    expect(find.text('5:00 /km'), findsOneWidget);
  });

  testWidgets('WorkoutHeartRateChart hides when fewer than 2 samples', (tester) async {
    await tester.pumpWidget(
      wrap(
        const WorkoutHeartRateChart(
          samples: [],
          hasGps: false,
        ),
      ),
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets('WorkoutHeartRateChart renders with GPS samples', (tester) async {
    await tester.pumpWidget(
      wrap(
        WorkoutHeartRateChart(
          samples: [
            WorkoutHeartRateSample(
              time: DateTime.utc(2026, 7, 8, 10),
              heartRateBpm: 120,
              distanceM: 0,
            ),
            WorkoutHeartRateSample(
              time: DateTime.utc(2026, 7, 8, 10, 1),
              heartRateBpm: 140,
              distanceM: 200,
            ),
          ],
          hasGps: true,
          heartRateAvg: 130,
          heartRateMax: 140,
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(find.byType(WorkoutHeartRateChart), findsOneWidget);
  });
}
