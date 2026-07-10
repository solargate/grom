// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'Travka';

  @override
  String get home => 'Home';

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
  String get serverUrlLabel => 'Server URL *';

  @override
  String get enterServerUrl => 'Enter server URL';

  @override
  String get enterValidServerUrl => 'Enter a valid URL (https://...)';

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
  String get workoutDuration => 'Duration';

  @override
  String get workoutDistance => 'Distance';

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
  String get retry => 'Retry';

  @override
  String get expandMap => 'Expand map';

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
  String get sportCategoryWater => 'Water Sports';

  @override
  String get sportCategoryWinter => 'Winter Sports';

  @override
  String get sportCategoryOther => 'Other Sports';

  @override
  String get sportTypeRun => 'Run';

  @override
  String get sportTypeHike => 'Hike';

  @override
  String get sportTypeTrailRun => 'Trail Run';

  @override
  String get sportTypeWheelchair => 'Wheelchair';

  @override
  String get sportTypeWalk => 'Walk';

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
  String get sportTypeStandUpPaddling => 'Stand Up Paddling';

  @override
  String get sportTypeKayaking => 'Kayak';

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
  String get recordingNotificationText => 'Tap to return to Travka';

  @override
  String get recordingNotificationChannelName => 'Workout recording';

  @override
  String get recordingPausedNotificationText => 'Workout recording paused';

  @override
  String get backgroundLocationRationale =>
      'Background location lets Travka keep recording your workout when you switch apps.';

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
  String get deleteWorkout => 'Delete';

  @override
  String get workoutActions => 'Workout actions';

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
}
