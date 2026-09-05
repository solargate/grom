/// Max activities fetched from Strava API per Home sync.
const kStravaApiSyncLimit = 10;

/// `external_id.name` for Strava API and ZIP imports (shared namespace).
const kStravaExternalIDName = 'strava';

/// OAuth redirect for Android (custom scheme so Auth Tab returns to the app).
/// Authorization Callback Domain in the user's Strava API app: `localhost`.
const kStravaOAuthRedirectUri = 'grom://localhost';

/// Scopes requested for public / followers activities (no private-only).
/// `read` is the baseline Strava scope; `activity:read` covers activity data.
const kStravaOAuthScope = 'read,activity:read';

/// Web authorize endpoint (works in Android Auth Tab / Custom Tabs).
const kStravaAuthorizeUrl = 'https://www.strava.com/oauth/authorize';
const kStravaTokenUrl = 'https://www.strava.com/oauth/token';
const kStravaApiBase = 'https://www.strava.com/api/v3';

/// Strava/Cloudflare often reject Dart's default User-Agent with HTTP 403.
const kStravaHttpUserAgent = 'GromAndroid/1.0 (Strava API client)';

const kStravaConnectButtonAsset =
    'assets/strava/btn_strava_connect_with_orange.svg';
