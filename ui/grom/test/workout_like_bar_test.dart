import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/workout.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/widgets/workout_like_bar.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

Widget wrap(Widget child) {
  return MaterialApp(
    theme: buildAppTheme(),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(body: child),
  );
}

Workout sampleWorkout({
  bool canLike = true,
  bool likedByMe = false,
  int likesCount = 2,
  int commentsCount = 3,
  String owner = 'bob',
}) {
  return Workout(
    id: 'wid',
    name: 'Run',
    description: '',
    sportType: 'Run',
    startDate: DateTime.utc(2026, 7, 8, 10),
    durationSeconds: 1800,
    distance: 5000,
    owner: owner,
    canLike: canLike,
    likedByMe: likedByMe,
    likesCount: likesCount,
    commentsCount: commentsCount,
  );
}

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ServerStorage.clear();
    await ServerStorage.saveBaseUrl('https://grom.example');
  });

  tearDown(() async {
    await ServerStorage.clear();
  });

  testWidgets('disables like button when canLike is false', (tester) async {
    await tester.pumpWidget(
      wrap(
        WorkoutLikeBar(
          workout: sampleWorkout(canLike: false),
          authToken: 'tok',
          api: ApiRequest(client: MockClient((_) async => http.Response('', 500))),
        ),
      ),
    );
    final button = tester.widget<IconButton>(find.byType(IconButton).first);
    expect(button.onPressed, isNull);
    expect(find.text('2'), findsOneWidget);
    expect(find.text('3'), findsOneWidget);
  });

  testWidgets('toggle like updates count from API', (tester) async {
    var liked = false;
    final client = MockClient((request) async {
      expect(request.url.path, '/api/v1/workouts/wid/likes');
      expect(request.url.queryParameters['owner'], 'bob');
      if (request.method == 'POST') {
        liked = true;
        return http.Response(
          jsonEncode({'count': 5, 'liked_by_me': true}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      liked = false;
      return http.Response(
        jsonEncode({'count': 2, 'liked_by_me': false}),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    await tester.pumpWidget(
      wrap(
        WorkoutLikeBar(
          workout: sampleWorkout(likesCount: 2, commentsCount: 0, likedByMe: false),
          authToken: 'tok',
          api: ApiRequest(client: client),
        ),
      ),
    );

    await tester.tap(find.byTooltip('Like workout'));
    await tester.pumpAndSettle();
    expect(liked, isTrue);
    expect(find.text('5'), findsOneWidget);
  });

  testWidgets('comments sheet lists, adds, and deletes comments', (tester) async {
    final comments = <Map<String, dynamic>>[
      {
        'id': 'c1',
        'datetime': '2026-08-06T12:00:00Z',
        'text': 'Nice pace',
        'can_delete': true,
        'user': {
          'handle': 'alice@localhost',
          'nickname': 'alice',
          'name': 'Alice',
          'is_local': true,
          'has_avatar': false,
        },
      },
    ];

    final client = MockClient((request) async {
      if (request.url.path.endsWith('/comments') && request.method == 'GET') {
        return http.Response(
          jsonEncode({'count': comments.length, 'comments': comments}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.url.path.endsWith('/comments') && request.method == 'POST') {
        final text = jsonDecode(request.body)['text'] as String;
        final created = {
          'id': 'c2',
          'datetime': '2026-08-06T13:00:00Z',
          'text': text,
          'can_delete': true,
          'user': {
            'handle': 'alice@localhost',
            'nickname': 'alice',
            'name': 'Alice',
            'is_local': true,
            'has_avatar': false,
          },
        };
        comments.add(created);
        return http.Response(
          jsonEncode({'count': comments.length, 'comment': created}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      if (request.method == 'DELETE') {
        comments.removeWhere((c) => c['id'] == 'c1');
        return http.Response(
          jsonEncode({'count': comments.length}),
          200,
          headers: {'content-type': 'application/json'},
        );
      }
      return http.Response('unexpected', 500);
    });

    await tester.pumpWidget(
      wrap(
        WorkoutLikeBar(
          workout: sampleWorkout(commentsCount: 1, likesCount: 0),
          authToken: 'tok',
          api: ApiRequest(client: client),
        ),
      ),
    );

    // Open comments via the comment IconButton on the right.
    await tester.tap(find.byTooltip('Comments'));
    await tester.pumpAndSettle();
    expect(find.text('Nice pace'), findsOneWidget);

    await tester.enterText(find.byType(TextField), 'Thanks!');
    await tester.tap(find.byTooltip('Add comment'));
    await tester.pumpAndSettle();
    expect(find.text('Thanks!'), findsOneWidget);

    final niceTile = find.ancestor(
      of: find.text('Nice pace'),
      matching: find.byType(ListTile),
    );
    await tester.ensureVisible(find.descendant(of: niceTile, matching: find.byIcon(Icons.delete)));
    await tester.tap(find.descendant(of: niceTile, matching: find.byIcon(Icons.delete)));
    await tester.pumpAndSettle();
    expect(find.text('Delete this comment?'), findsOneWidget);
    await tester.tap(find.widgetWithText(TextButton, 'Delete comment'));
    await tester.pumpAndSettle();
    expect(find.text('Nice pace'), findsNothing);
    expect(find.text('Thanks!'), findsOneWidget);
  });
}
