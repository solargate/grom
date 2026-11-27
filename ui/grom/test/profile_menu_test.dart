import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/navigation/grom_shell.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/session.dart';
import 'package:grom/widgets/profile_menu.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

Widget wrap(Widget child) {
  return MaterialApp(
    theme: buildAppTheme(),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(
      appBar: AppBar(actions: [child]),
    ),
  );
}

void main() {
  testWidgets('ProfileMenu puts both destructive actions after Edit profile',
      (tester) async {
    ProfileMenuAction? selected;

    await tester.pumpWidget(
      wrap(
        ProfileMenu(
          onSelected: (action) => selected = action,
        ),
      ),
    );

    await tester.tap(find.byType(PopupMenuButton<ProfileMenuAction>));
    await tester.pumpAndSettle();

    expect(find.text('Edit profile'), findsOneWidget);
    final deleteData = find.text('Delete workout data');
    final deleteAccount = find.text('Delete account');
    expect(tester.getTopLeft(find.text('Edit profile')).dy,
        lessThan(tester.getTopLeft(deleteData).dy));
    expect(tester.getTopLeft(deleteData).dy,
        lessThan(tester.getTopLeft(deleteAccount).dy));
    final errorColor =
        Theme.of(tester.element(deleteAccount)).colorScheme.error;
    expect(tester.widget<Text>(deleteData).style?.color, errorColor);
    expect(tester.widget<Text>(deleteAccount).style?.color, errorColor);

    await tester.tap(deleteData);
    await tester.pumpAndSettle();
    expect(selected, ProfileMenuAction.deleteWorkoutData);

    await tester.tap(find.byType(PopupMenuButton<ProfileMenuAction>));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Delete account'));
    await tester.pumpAndSettle();
    expect(selected, ProfileMenuAction.deleteAccount);
  });

  testWidgets('Deleting workouts refreshes the profile and keeps the session',
      (tester) async {
    SharedPreferences.setMockInitialValues({});
    SessionCoordinator.instance.debugReset();
    addTearDown(SessionCoordinator.instance.debugReset);
    await ServerStorage.saveBaseUrl('https://grom.example');
    await AuthStorage.saveToken('session');

    final deletionResponse = Completer<http.Response>();
    final passwords = <String>[];
    var total = 2;
    final client = MockClient((request) async {
      switch (request.url.path) {
        case '/api/v1/server-info':
          return http.Response('{"name":"Test Grom"}', 200);
        case '/api/v1/auth/me':
          return http.Response(
            '{"id":"u1","nickname":"alice","name":"Alice","email":"alice@example.com"}',
            200,
          );
        case '/api/v1/social/following':
        case '/api/v1/social/followers':
          return http.Response('[]', 200);
        case '/api/v1/profile':
          return http.Response(jsonEncode({'total_workouts': total}), 200);
        case '/api/v1/workouts':
          if (request.method == 'GET') {
            return http.Response('{"items":[],"has_more":false}', 200);
          }
          expect(request.method, 'DELETE');
          expect(request.headers['authorization'], 'Bearer session');
          expect(request.headers['content-type'], 'application/json');
          final password = jsonDecode(request.body)['password'] as String;
          passwords.add(password);
          if (password == 'incorrect') {
            return http.Response('{"error":"invalid password"}', 403);
          }
          return deletionResponse.future;
      }
      fail('Unexpected request ${request.method} ${request.url}');
    });
    final messenger = GlobalKey<ScaffoldMessengerState>();

    await http.runWithClient(() async {
      await tester.pumpWidget(MaterialApp(
        scaffoldMessengerKey: messenger,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: GromShell(locale: const Locale('en'), onLocaleChanged: (_) {}),
      ));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Profile').first);
      await tester.pumpAndSettle();
      expect(find.text('Total workouts: 2'), findsOneWidget);

      Future<void> openConfirmation() async {
        await tester.tap(find.byType(ProfileMenu));
        await tester.pumpAndSettle();
        await tester.tap(find.text('Delete workout data'));
        await tester.pumpAndSettle();
      }

      await openConfirmation();
      expect(
          tester.widget<TextField>(find.byType(TextField)).obscureText, isTrue);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(passwords, isEmpty);

      await openConfirmation();
      await tester.enterText(find.byType(TextField), 'incorrect');
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect(find.text('Invalid password'), findsOneWidget);
      expect(find.text('Total workouts: 2'), findsOneWidget);
      messenger.currentState!.hideCurrentSnackBar();
      await tester.pumpAndSettle();

      await openConfirmation();
      await tester.enterText(find.byType(TextField), 'correct password');
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect(
          tester
              .widget<ProfileMenu>(find.byType(ProfileMenu))
              .deletingWorkoutData,
          isTrue);

      total = 0;
      deletionResponse.complete(http.Response('', 204));
      await tester.pumpAndSettle();
      expect(passwords, ['incorrect', 'correct password']);
      expect(find.text('Workout data deleted'), findsOneWidget);
      expect(find.text('Total workouts: 0'), findsOneWidget);
      expect(await AuthStorage.getToken(), 'session');
      await tester.pumpWidget(const SizedBox.shrink());
    }, () => client);
  }, variant: const TargetPlatformVariant({TargetPlatform.linux}));
}
