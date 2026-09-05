import 'package:flutter/foundation.dart';
import 'package:flutter_web_auth_2/flutter_web_auth_2.dart';

import '../services/strava_api_constants.dart';

/// Opens Strava OAuth (Custom Tabs / Auth Tab) and returns the redirect URL.
Future<String> authenticateStravaOAuth(String authorizationUrl) async {
  if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) {
    throw UnsupportedError('Strava OAuth is Android-only');
  }

  final redirect = Uri.parse(kStravaOAuthRedirectUri);
  // Custom scheme (not http/https) so Android Auth Tab hands the URL back to
  // CallbackActivity instead of loading it as a web page.
  return FlutterWebAuth2.authenticate(
    url: authorizationUrl,
    callbackUrlScheme: redirect.scheme,
  );
}
