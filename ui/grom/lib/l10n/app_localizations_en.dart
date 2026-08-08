// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'Grom';

  @override
  String get home => 'Home';

  @override
  String get homeTabFeed => 'Feed';

  @override
  String get homeTabMyWorkouts => 'My workouts';

  @override
  String get signIn => 'Sign in';

  @override
  String get register => 'Register';

  @override
  String get signOut => 'Sign out';

  @override
  String get settings => 'Settings';

  @override
  String get add => 'Add';

  @override
  String welcomeUser(String nickname) {
    return 'Welcome, $nickname!';
  }

  @override
  String get registrationSuccessful =>
      'Registration successful. Please sign in.';

  @override
  String get signedOut => 'You have signed out';

  @override
  String signedInAs(String nickname) {
    return 'Signed in as $nickname';
  }

  @override
  String get failedToSignIn => 'Failed to sign in';

  @override
  String get failedToRegister => 'Failed to register';

  @override
  String get enterEmail => 'Enter email';

  @override
  String get enterValidEmail => 'Enter a valid email';

  @override
  String get emailLabel => 'Email *';

  @override
  String get enterPassword => 'Enter password';

  @override
  String get passwordLabel => 'Password *';

  @override
  String get enterNickname => 'Enter nickname';

  @override
  String get nicknameLabel => 'Nickname *';

  @override
  String get nameLabel => 'Full name';

  @override
  String get passwordMinLength => 'Password must be at least 8 characters';

  @override
  String get confirmPasswordLabel => 'Confirm password *';

  @override
  String get confirmPassword => 'Confirm password';

  @override
  String get passwordsDoNotMatch => 'Passwords do not match';

  @override
  String get forgotPasswordLink => 'Forgot password?';

  @override
  String get forgotPasswordTitle => 'Reset password';

  @override
  String get forgotPasswordHint =>
      'Enter your account email. If it is registered, we will send a reset link. Open the link in a browser to choose a new password.';

  @override
  String get forgotPasswordSubmit => 'Send reset link';

  @override
  String get forgotPasswordCheckEmail =>
      'If an account exists for that email, a reset link has been sent. Open it in a browser, then sign in here.';

  @override
  String get forgotPasswordFailed => 'Failed to request password reset';

  @override
  String get captchaRequired => 'Complete the captcha check';

  @override
  String get captchaNotRobot => 'I\'m not a robot';

  @override
  String get resetPasswordTitle => 'Choose a new password';

  @override
  String get resetPasswordHint => 'Enter a new password for your account.';

  @override
  String get resetPasswordSubmit => 'Update password';

  @override
  String get resetPasswordSuccess => 'Password updated. Please sign in.';

  @override
  String get resetPasswordFailed => 'Failed to reset password';

  @override
  String get resetPasswordInvalidToken =>
      'This reset link is missing or invalid.';

  @override
  String get serverUrlLabel => 'Server URL *';

  @override
  String get enterServerUrl => 'Enter server URL';

  @override
  String get enterValidServerUrl => 'Enter a valid server host or URL';

  @override
  String get serverUrlHint => 'example.com';

  @override
  String get language => 'Language';

  @override
  String get languageEnglish => 'English';

  @override
  String get languageRussian => 'Russian';

  @override
  String get languageGerman => 'German';

  @override
  String get addWorkout => 'Add workout';

  @override
  String get workoutName => 'Workout name';

  @override
  String get workoutDescription => 'Description';

  @override
  String get workoutType => 'Workout type';

  @override
  String get workoutDate => 'Date';

  @override
  String get workoutStartTime => 'Start time';

  @override
  String get workoutDuration => 'Time';

  @override
  String get workoutDistance => 'Distance';

  @override
  String get workoutPace => 'Pace';

  @override
  String get workoutElevationGain => 'Elevation gain';

  @override
  String get workoutSpeedAvg => 'Avg. speed';

  @override
  String get workoutSpeedMax => 'Max. speed';

  @override
  String get workoutSpeedChartTitle => 'Speed';

  @override
  String get workoutHeartRateChartTitle => 'Heart rate';

  @override
  String get workoutTotalTime => 'Total time';

  @override
  String get workoutHeartRateAvg => 'Avg. heart rate';

  @override
  String get workoutHeartRateMax => 'Max. heart rate';

  @override
  String chartMinutes(String value) {
    return '$value min';
  }

  @override
  String get workoutSteps => 'Steps';

  @override
  String get workoutCalories => 'Calories';

  @override
  String elevationMeters(String value) {
    return '$value m';
  }

  @override
  String heartRateBpm(String value) {
    return '$value';
  }

  @override
  String stepsCount(String value) {
    return '$value';
  }

  @override
  String caloriesKcal(String value) {
    return '$value';
  }

  @override
  String get save => 'Save';

  @override
  String get cancel => 'Cancel';

  @override
  String get selectWorkoutType => 'Select workout type';

  @override
  String get enterWorkoutName => 'Enter workout name';

  @override
  String get workoutSaved => 'Workout saved';

  @override
  String get failedToSaveWorkout => 'Failed to save workout';

  @override
  String get failedToLoadWorkouts => 'Failed to load workouts';

  @override
  String get failedToLoadWorkoutLikes => 'Failed to load workout likes';

  @override
  String get failedToUpdateWorkoutLike => 'Failed to update workout like';

  @override
  String get workoutLikeAction => 'Like workout';

  @override
  String get workoutNoLikesYet => 'No likes yet';

  @override
  String workoutLikesTitle(String count) {
    return 'Likes ($count)';
  }

  @override
  String get failedToLoadWorkoutComments => 'Failed to load workout comments';

  @override
  String get failedToAddWorkoutComment => 'Failed to add comment';

  @override
  String get failedToDeleteWorkoutComment => 'Failed to delete comment';

  @override
  String get workoutCommentAction => 'Comments';

  @override
  String get workoutNoCommentsYet => 'No comments yet';

  @override
  String workoutCommentsTitle(String count) {
    return 'Comments ($count)';
  }

  @override
  String get workoutCommentHint => 'Write a comment';

  @override
  String get addWorkoutCommentAction => 'Add comment';

  @override
  String get deleteWorkoutCommentAction => 'Delete comment';

  @override
  String get deleteWorkoutCommentTitle => 'Delete comment?';

  @override
  String get deleteWorkoutCommentConfirm => 'Delete this comment?';

  @override
  String get retry => 'Retry';

  @override
  String get expandMap => 'Expand map';

  @override
  String get addPhotos => 'Add photos';

  @override
  String photosSelected(int count) {
    return '$count photos selected';
  }

  @override
  String get removePhoto => 'Remove photo';

  @override
  String get failedToUploadPhotos => 'Failed to upload photos';

  @override
  String get closePhotoViewer => 'Close';

  @override
  String get collapseMap => 'Collapse map';

  @override
  String get noWorkoutsYet => 'You have no workouts yet';

  @override
  String get durationZero => '0s';

  @override
  String get distanceZero => '0 km';

  @override
  String durationHours(int hours) {
    return '${hours}h';
  }

  @override
  String durationMinutes(int minutes) {
    return '${minutes}m';
  }

  @override
  String durationSeconds(int seconds) {
    return '${seconds}s';
  }

  @override
  String distanceKilometers(String value) {
    return '$value km';
  }

  @override
  String distanceMeters(int value) {
    return '$value m';
  }

  @override
  String get selectDuration => 'Select duration';

  @override
  String get selectDistance => 'Select distance';

  @override
  String get hoursLabel => 'Hours';

  @override
  String get minutesLabel => 'Minutes';

  @override
  String get secondsLabel => 'Seconds';

  @override
  String get kilometersLabel => 'Kilometers';

  @override
  String get sportCategoryFoot => 'Foot Sports';

  @override
  String get sportCategoryCycle => 'Cycle Sports';

  @override
  String get sportCategoryStrength => 'Strength Sports';

  @override
  String get sportCategoryWater => 'Water Sports';

  @override
  String get sportCategoryWinter => 'Winter Sports';

  @override
  String get sportCategoryTeam => 'Team Sports';

  @override
  String get sportCategoryRacket => 'Racket Sports';

  @override
  String get sportCategoryOther => 'Other Sports';

  @override
  String get sportTypeRun => 'Run';

  @override
  String get sportTypeHike => 'Hiking';

  @override
  String get sportTypeTrailRun => 'Trail Running';

  @override
  String get sportTypeWheelchair => 'Wheelchair';

  @override
  String get sportTypeWalk => 'Walk';

  @override
  String get sportTypeNordicWalk => 'Nordic Walk';

  @override
  String get sportTypeRide => 'Ride';

  @override
  String get sportTypeEBikeRide => 'E-Bike Ride';

  @override
  String get sportTypeMountainBikeRide => 'Mountain Bike Ride';

  @override
  String get sportTypeEMountainBikeRide => 'E-Mountain Bike Ride';

  @override
  String get sportTypeGravelRide => 'Gravel Ride';

  @override
  String get sportTypeVelomobile => 'Velomobile';

  @override
  String get sportTypeHandcycle => 'Handcycle';

  @override
  String get sportTypeCanoeing => 'Canoe';

  @override
  String get sportTypeStandUpPaddling => 'SUP';

  @override
  String get sportTypeKayaking => 'Kayak';

  @override
  String get sportTypePackraft => 'Packraft';

  @override
  String get sportTypeSurfing => 'Surf';

  @override
  String get sportTypeKitesurf => 'Kitesurf';

  @override
  String get sportTypeSwim => 'Swim';

  @override
  String get sportTypeRowing => 'Rowing';

  @override
  String get sportTypeWindsurf => 'Windsurf';

  @override
  String get sportTypeSail => 'Sailing';

  @override
  String get sportTypeIceSkate => 'Ice Skate';

  @override
  String get sportTypeNordicSki => 'Nordic Ski';

  @override
  String get sportTypeAlpineSki => 'Alpine Ski';

  @override
  String get sportTypeSnowboard => 'Snowboard';

  @override
  String get sportTypeBackcountrySki => 'Backcountry Ski';

  @override
  String get sportTypeIceHockey => 'Ice Hockey';

  @override
  String get sportTypeSnowshoe => 'Snowshoe';

  @override
  String get sportTypeWorkout => 'Workout';

  @override
  String get sportTypeGolf => 'Golf';

  @override
  String get sportTypeBadminton => 'Badminton';

  @override
  String get sportTypeElliptical => 'Eliptical';

  @override
  String get sportTypeBasketball => 'Basketball';

  @override
  String get sportTypeInlineSkate => 'Inline Skate';

  @override
  String get sportTypeSkateboard => 'Skateboarding';

  @override
  String get sportTypeTennis => 'Tennis';

  @override
  String get sportTypeStairStepper => 'Stair Stepper';

  @override
  String get sportTypePadel => 'Padel';

  @override
  String get sportTypeRockClimbing => 'Rock Climb';

  @override
  String get sportTypeSoccer => 'Football (Soccer)';

  @override
  String get sportTypePickleball => 'Pickleball';

  @override
  String get sportTypeWeightTraining => 'Weight Training';

  @override
  String get sportTypeVolleyball => 'Volleyball';

  @override
  String get sportTypeRollerSki => 'Roller Ski';

  @override
  String get sportTypeSquash => 'Squash';

  @override
  String get sportTypeCrossfit => 'Crossfit';

  @override
  String get sportTypeYoga => 'Yoga';

  @override
  String get sportTypeDance => 'Dance';

  @override
  String get sportTypeTableTennis => 'Table Tennis';

  @override
  String get sportTypePilates => 'Pilates';

  @override
  String get sportTypeRacquetball => 'Racquetball';

  @override
  String get sportTypeHiit => 'HIIT';

  @override
  String get sportTypeCricket => 'Cricket';

  @override
  String get workoutTrack => 'Track';

  @override
  String get selectTrackFile => 'Select FIT or GPX file';

  @override
  String trackFileSelected(String filename) {
    return '$filename';
  }

  @override
  String get removeTrack => 'Remove track';

  @override
  String get invalidTrackFormat => 'Only FIT and GPX files are supported';

  @override
  String get failedToParseTrack => 'Failed to read track file';

  @override
  String get trackMetadataApplied => 'Values updated from track';

  @override
  String get shareTrackLoginRequired => 'Log in to import a shared track';

  @override
  String get shareTrackReadFailed => 'Could not read the shared file';

  @override
  String get tabRecord => 'Record';

  @override
  String get tabManual => 'Manual';

  @override
  String get recordStart => 'Record';

  @override
  String get recordPause => 'Pause';

  @override
  String get recordFinish => 'Finish';

  @override
  String get recordingDuration => 'Duration';

  @override
  String get currentSpeed => 'Speed';

  @override
  String speedKmh(String speed) {
    return '$speed km/h';
  }

  @override
  String get speedUnavailable => '—';

  @override
  String get locationPermissionDenied =>
      'Location permission is required to record a workout track';

  @override
  String get notificationPermissionDenied =>
      'Notification permission is required to show the recording status';

  @override
  String get locationServicesDisabled =>
      'Enable location services to record a workout track';

  @override
  String get openSettings => 'Open settings';

  @override
  String get discardRecordingTitle => 'Discard recording?';

  @override
  String get discardRecordingMessage =>
      'The current track recording will be lost.';

  @override
  String get discardRecordingConfirm => 'Discard';

  @override
  String get recordingNotificationTitle => 'Recording workout';

  @override
  String get recordingNotificationText => 'Tap to return to Grom';

  @override
  String get recordingNotificationChannelName => 'Workout recording';

  @override
  String get recordingPausedNotificationText => 'Workout recording paused';

  @override
  String get recordingAutoPausedNotificationText =>
      'Workout recording auto-paused';

  @override
  String get autoPauseEnabled => 'Auto-pause on';

  @override
  String get autoPauseDisabled => 'Auto-pause off';

  @override
  String get backgroundLocationRationale =>
      'Background location lets Grom keep recording your workout when you switch apps.';

  @override
  String get doNotDismissNotification =>
      'Do not dismiss the notification while recording';

  @override
  String get recordingInProgress => 'Recording in progress';

  @override
  String get restoreRecordingTitle => 'Resume recording?';

  @override
  String get restoreRecordingMessage =>
      'An unfinished workout recording was found. Resume it or discard the saved track.';

  @override
  String get restoreRecordingConfirm => 'Resume';

  @override
  String get restoreRecordingDiscard => 'Discard';

  @override
  String get editWorkout => 'Edit';

  @override
  String get editWorkoutTitle => 'Edit workout';

  @override
  String get deleteWorkout => 'Delete';

  @override
  String get deleteWorkoutConfirm =>
      'The workout will be permanently deleted and cannot be restored.';

  @override
  String get workoutDeleted => 'Workout deleted';

  @override
  String get failedToDeleteWorkout => 'Failed to delete workout';

  @override
  String get workoutActions => 'Workout actions';

  @override
  String get downloadTrackAsGpx => 'Download track as GPX';

  @override
  String get downloadTrackOriginal => 'Download track (original)';

  @override
  String get downloadingTrack => 'Downloading track…';

  @override
  String get failedToDownloadTrack => 'Failed to download track';

  @override
  String get trackSaved => 'Track saved';

  @override
  String get failedToLoadWorkoutTrack => 'Failed to load workout track';

  @override
  String get userSearch => 'User search';

  @override
  String get profile => 'Profile';

  @override
  String get searchUsersHint => 'Nickname or @user@server';

  @override
  String get search => 'Search';

  @override
  String get follow => 'Follow';

  @override
  String get unfollow => 'Unfollow';

  @override
  String get following => 'Following';

  @override
  String get followers => 'Followers';

  @override
  String get followPending => 'Pending';

  @override
  String get noUsersFound => 'No users found';

  @override
  String get noFollowingYet => 'You are not following anyone yet';

  @override
  String get noFollowersYet => 'No one is following you yet';

  @override
  String get searchByNicknameOrHandle =>
      'Search by nickname or federated handle (@user@server)';

  @override
  String workoutByAuthor(String author) {
    return 'By $author';
  }

  @override
  String get failedToSearchUsers => 'Failed to search users';

  @override
  String get failedToLoadProfile => 'Failed to load profile';

  @override
  String get editProfile => 'Edit profile';

  @override
  String get profileSaved => 'Profile saved';

  @override
  String get avatarUpdated => 'Avatar updated';

  @override
  String get failedToUploadAvatar => 'Failed to upload avatar';

  @override
  String get cropAvatarTitle => 'Crop avatar';

  @override
  String get cropAvatarDone => 'Done';

  @override
  String get failedToSaveProfile => 'Failed to save profile';

  @override
  String get equipment => 'Equipment';

  @override
  String get addEquipment => 'Add';

  @override
  String get selectEquipment => 'Select equipment';

  @override
  String get equipmentType => 'Type';

  @override
  String get equipmentName => 'Name';

  @override
  String get equipmentBrand => 'Brand';

  @override
  String get equipmentModel => 'Model';

  @override
  String get equipmentWeight => 'Weight (kg)';

  @override
  String get equipmentNotes => 'Notes';

  @override
  String get workoutEquipment => 'Equipment';

  @override
  String get workoutDevice => 'Device';

  @override
  String get bikeType => 'Bike type';

  @override
  String get waterEquipmentType => 'Water equipment type';

  @override
  String get deleteEquipment => 'Delete';

  @override
  String get deleteEquipmentConfirm =>
      'Delete this equipment? It will be removed from all workouts.';

  @override
  String get noEquipmentYet => 'You have no equipment yet';

  @override
  String get equipmentSaved => 'Equipment saved';

  @override
  String get equipmentDeleted => 'Equipment deleted';

  @override
  String get failedToLoadEquipment => 'Failed to load equipment';

  @override
  String get failedToSaveEquipment => 'Failed to save equipment';

  @override
  String get enterEquipmentName => 'Enter equipment name';

  @override
  String get equipmentTypeBike => 'Bike';

  @override
  String get equipmentTypeShoes => 'Shoes';

  @override
  String get equipmentTypeWater => 'Water equipment';

  @override
  String get equipmentTypeOther => 'Other';

  @override
  String get equipmentSubtypeEmpty => 'Not selected';

  @override
  String get bikeTypeMountain => 'Mountain';

  @override
  String get bikeTypeGravel => 'Gravel';

  @override
  String get bikeTypeRoad => 'Road';

  @override
  String get bikeTypeTouring => 'Touring';

  @override
  String get bikeTypeTriathlon => 'Triathlon';

  @override
  String get bikeTypeCyclocross => 'Cyclocross';

  @override
  String get bikeTypeFixie => 'Fixie';

  @override
  String get bikeTypeFolding => 'Folding';

  @override
  String get bikeTypeBmx => 'BMX';

  @override
  String get waterTypeSup => 'SUP';

  @override
  String get waterTypeKayak => 'Kayak';

  @override
  String get waterTypeCanoe => 'Canoe';

  @override
  String get waterTypeCanoeDouble => 'Double canoe';

  @override
  String get waterTypePackraft => 'Packraft';

  @override
  String get waterTypeSurf => 'Surf';

  @override
  String get about => 'About';

  @override
  String get aboutAuthorLabel => 'Author';

  @override
  String get aboutSourceCodeLabel => 'Source code';

  @override
  String get aboutLicenseLabel => 'License';

  @override
  String get mapDataAttributionTitle => 'Map data';

  @override
  String get openStreetMapAttribution => '© OpenStreetMap contributors';

  @override
  String get openStreetMapLicense =>
      'Map previews and interactive maps use data from OpenStreetMap, available under the Open Database License (ODbL).';

  @override
  String get openStreetMapCopyrightLink =>
      'OpenStreetMap copyright and license';

  @override
  String get integration => 'Integration';

  @override
  String get strava => 'Strava';

  @override
  String get stravaImportDescriptionBefore =>
      'You can download an archive of your workouts from the ';

  @override
  String get stravaImportDescriptionLink => 'Strava website';

  @override
  String get stravaDownloadArchiveUrl =>
      'https://www.strava.com/athlete/download_my_account';

  @override
  String get stravaImportDescriptionAfter =>
      '. Upload the resulting ZIP archive to Grom. All workouts will be imported with tracks, equipment, and photos.';

  @override
  String get importStravaArchive => 'Import Strava archive';

  @override
  String get uploading => 'Uploading';

  @override
  String get importing => 'Importing';

  @override
  String stravaImportCompleted(int imported, int skipped, int parseSkipped,
      int mediaMissing, int errors) {
    return 'Strava import completed: $imported imported, $skipped skipped, $parseSkipped CSV parse skipped, $mediaMissing media files missing from archive, $errors errors';
  }

  @override
  String stravaImportFailed(String message) {
    return 'Strava import failed: $message';
  }

  @override
  String get stravaImportInProgress => 'Another import is already in progress';

  @override
  String get healthSyncGoogleDrive => 'Health Sync + Google Drive';

  @override
  String get healthSyncImportDescriptionBefore => 'You can use the ';

  @override
  String get healthSyncImportDescriptionLink => 'Health Sync';

  @override
  String get healthSyncImportDescriptionAfter =>
      ' app to sync workouts from various services to your Google Drive. Workouts stored in Google Drive can be imported into Grom.';

  @override
  String get healthSyncPlayStoreUrl =>
      'https://play.google.com/store/apps/details?id=nl.appyhapps.healthsync';

  @override
  String get healthSyncSyncToggle => 'Health Sync + Google Drive sync';

  @override
  String get healthSyncFolderLabel => 'Health Sync folder';

  @override
  String get healthSyncSync => 'Sync';

  @override
  String get healthSyncSynchronizing => 'Synchronizing…';

  @override
  String healthSyncImported(int count) {
    return 'Imported $count workouts';
  }

  @override
  String get healthSyncNoNewWorkouts => 'No new workouts found';

  @override
  String get healthSyncFolderNotFound =>
      'Health Sync folder not found on Google Drive';

  @override
  String get healthSyncFolderEmpty => 'Health Sync folder is empty';

  @override
  String get healthSyncFolderNameRequired => 'Enter a Health Sync folder name';

  @override
  String get healthSyncFindFolder => 'Find Health Sync folder';

  @override
  String get healthSyncGoogleSignInCancelled => 'Google sign-in was cancelled';

  @override
  String get healthSyncGoogleSignInFailed => 'Google sign-in failed';

  @override
  String get healthSyncDriveAccessDenied => 'Google Drive access was denied';

  @override
  String healthSyncSyncError(String message) {
    return 'Health Sync import failed: $message';
  }
}
