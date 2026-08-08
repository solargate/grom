import 'package:altcha_widget/altcha_widget.dart';
import 'package:flutter/material.dart';

import '../api_request.dart';

/// ALTCHA checkbox shown when the server has captcha enabled.
class AltchaField extends StatelessWidget {
  const AltchaField({
    super.key,
    required this.enabled,
    required this.challengeUrl,
    required this.onPayloadChanged,
  });

  final bool enabled;
  final String challengeUrl;
  final ValueChanged<String?> onPayloadChanged;

  @override
  Widget build(BuildContext context) {
    if (!widgetEnabled) {
      return const SizedBox.shrink();
    }

    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: AltchaWidget(
        key: ValueKey(challengeUrl),
        challenge: challengeUrl,
        hideFooter: true,
        onVerified: (payload) => onPayloadChanged(payload),
        onFailed: (_) => onPayloadChanged(null),
      ),
    );
  }

  bool get widgetEnabled => enabled && challengeUrl.isNotEmpty;
}

/// Resolves the absolute captcha challenge URL for the current API base.
String captchaChallengeUrl(ApiRequest api) {
  final uri = api.resolveUri('/api/v1/captcha/challenge');
  if (uri.hasScheme) {
    return uri.toString();
  }
  return Uri.base.resolveUri(uri).toString();
}

/// Label overrides for [AltchaLocalizationsDelegate] (includes RU).
const altchaCustomTranslations = <String, Map<String, String>>{
  'en': {'label': "I'm not a robot"},
  'de': {'label': 'Ich bin kein Roboter'},
  'ru': {'label': 'Я не робот'},
};
