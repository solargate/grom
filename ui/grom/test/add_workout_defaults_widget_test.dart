import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/widgets/add_workout_sheet.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await AuthStorage.saveToken('tok');
  });

  testWidgets('AddWorkoutSheet applies profile sport and equipment defaults', (tester) async {
    final client = MockClient((request) async {
      switch (request.url.path) {
        case '/api/v1/profile':
          return http.Response(
            jsonEncode({
              'last_sport_type': 'Ride',
              'last_equipment_by_sport': {
                'Ride': ['bike-1', 'missing-bike'],
              },
            }),
            200,
            headers: {'content-type': 'application/json'},
          );
        case '/api/v1/equipment':
          return http.Response(
            jsonEncode([
              {
                'id': 'bike-1',
                'type': 'Bike',
                'name': 'Road bike',
                'bike_type': 'road',
              },
            ]),
            200,
            headers: {'content-type': 'application/json'},
          );
        default:
          return http.Response('not found', 404);
      }
    });

    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: AddWorkoutSheet(api: ApiRequest(client: client)),
        ),
      ),
    );

    await tester.pump();
    await tester.pumpAndSettle();

    expect(find.text('Ride'), findsWidgets);
    expect(find.text('Road bike'), findsOneWidget);
  });

  testWidgets('AddWorkoutSheet falls back to default sport when profile empty', (tester) async {
    final client = MockClient((request) async {
      switch (request.url.path) {
        case '/api/v1/profile':
          return http.Response(
            jsonEncode({}),
            200,
            headers: {'content-type': 'application/json'},
          );
        case '/api/v1/equipment':
          return http.Response(
            jsonEncode([]),
            200,
            headers: {'content-type': 'application/json'},
          );
        default:
          return http.Response('not found', 404);
      }
    });

    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          body: AddWorkoutSheet(api: ApiRequest(client: client)),
        ),
      ),
    );

    await tester.pump();
    await tester.pumpAndSettle();

    final defaultLabel = sportTypeLabel(
      AppLocalizationsEn(),
      defaultSportTypeId,
    );
    expect(find.text(defaultLabel), findsWidgets);
  });
}
