/// Web does not expose Strava API import; OAuth is Android-only.
Future<String> authenticateStravaOAuth(String authorizationUrl) async {
  throw UnsupportedError('Strava OAuth is Android-only');
}
