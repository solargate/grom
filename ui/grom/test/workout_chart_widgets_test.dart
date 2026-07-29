import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/workout_heartrate.dart';
import 'package:grom/models/workout_speed.dart';
import 'package:grom/widgets/workout_heartrate_chart.dart';
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
