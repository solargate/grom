/// Max activities fetched from Strava API per Home sync.
const kStravaApiSyncLimit = 10;

/// `external_id.name` for Strava API and ZIP imports (shared namespace).
const kStravaExternalIDName = 'strava';

/// OAuth redirect for Android (custom scheme so Auth Tab returns to the app).
/// Authorization Callback Domain in the user's Strava API app: `localhost`.
const kStravaOAuthRedirectUri = 'grom://localhost/exchange_token';

/// Scopes requested for public / followers activities (no private-only).
const kStravaOAuthScope = 'activity:read';

const kStravaAuthorizeUrl = 'https://www.strava.com/oauth/mobile/authorize';
const kStravaTokenUrl = 'https://www.strava.com/oauth/token';
const kStravaApiBase = 'https://www.strava.com/api/v3';

const kStravaConnectButtonAsset =
    'assets/strava/btn_strava_connect_with_orange.svg';
