import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/app_theme.dart';
import 'package:grom/widgets/user_avatar.dart';

void main() {
  testWidgets('UserAvatar without avatar shows placeholder icon', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        home: const Scaffold(
          body: UserAvatar(
            nickname: 'alice',
            hasAvatar: false,
          ),
        ),
      ),
    );

    expect(tester.takeException(), isNull);
    expect(find.byIcon(Icons.person), findsOneWidget);
    expect(find.byType(CircleAvatar), findsOneWidget);
  });

  testWidgets('UserAvatar with hasAvatar false does not throw without network', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: buildAppTheme(),
        home: const Scaffold(
          body: UserAvatar(
            nickname: 'bob',
            hasAvatar: false,
            avatarUrl: null,
          ),
        ),
      ),
    );

    await tester.pump();
    expect(tester.takeException(), isNull);
    expect(find.byIcon(Icons.person), findsOneWidget);
  });
}
