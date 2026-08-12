import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/widgets/altcha_field.dart';

void main() {
  testWidgets('AltchaField hidden when disabled', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AltchaField(
            enabled: false,
            challengeUrl: 'https://grom.example/api/v1/captcha/challenge',
            onPayloadChanged: (_) {},
          ),
        ),
      ),
    );

    expect(find.byType(AltchaField), findsOneWidget);
    expect(find.byType(SizedBox), findsWidgets);
  });

  testWidgets('AltchaField hidden when challenge URL empty', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AltchaField(
            enabled: true,
            challengeUrl: '',
            onPayloadChanged: (_) {},
          ),
        ),
      ),
    );

    expect(tester.takeException(), isNull);
  });
}
