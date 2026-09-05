/// Default / fallback for Home sync activity count (user-editable).
const kStravaApiSyncLimitDefault = 10;

/// Inclusive bounds for the sync limit field (Strava `per_page` max is 200).
const kStravaApiSyncLimitMin = 1;
const kStravaApiSyncLimitMax = 200;

/// Clamp a user-entered sync limit; invalid values fall back to the default.
int clampStravaApiSyncLimit(int? value) {
  if (value == null || value < kStravaApiSyncLimitMin) {
    return kStravaApiSyncLimitDefault;
  }
  if (value > kStravaApiSyncLimitMax) {
    return kStravaApiSyncLimitMax;
  }
  return value;
}

/// `external_id.name` for Strava API and ZIP imports (shared namespace).
const kStravaExternalIDName = 'strava';

/// OAuth redirect for Android (custom scheme so Auth Tab returns to the app).
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
