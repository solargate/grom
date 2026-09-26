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

  group('isCrossOriginAvatarUrl', () {
    test('relative and empty are same-origin', () {
      expect(isCrossOriginAvatarUrl(''), isFalse);
      expect(
        isCrossOriginAvatarUrl(
          '/api/v1/federation/authors/bob_remote/avatar',
          localBase: 'https://grom.example',
        ),
        isFalse,
      );
    });

    test('absolute local base is same-origin', () {
      expect(
        isCrossOriginAvatarUrl(
          'https://grom.example/api/v1/users/bob/avatar',
          localBase: 'https://grom.example',
        ),
        isFalse,
      );
    });

    test('absolute other host is cross-origin', () {
      expect(
        isCrossOriginAvatarUrl(
          'https://other.example/users/bob/avatar',
          localBase: 'https://grom.example',
        ),
        isTrue,
      );
    });

    test('absolute URL without known base is treated as cross-origin', () {
      expect(
        isCrossOriginAvatarUrl(
          'https://other.example/users/bob/avatar',
          localBase: null,
        ),
        isTrue,
      );
    });
  });
}
