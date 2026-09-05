import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_de.dart';
import 'app_localizations_en.dart';
import 'app_localizations_ru.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('de'),
    Locale('en'),
    Locale('ru')
  ];

  /// No description provided for @appTitle.
  ///
  /// In en, this message translates to:
  /// **'Grom'**
  String get appTitle;

  /// No description provided for @home.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get home;

  /// No description provided for @homeTabFeed.
  ///
  /// In en, this message translates to:
  /// **'Feed'**
  String get homeTabFeed;

  /// No description provided for @homeTabMyWorkouts.
  ///
  /// In en, this message translates to:
  /// **'My workouts'**
  String get homeTabMyWorkouts;

  /// No description provided for @filterWorkouts.
  ///
  /// In en, this message translates to:
  /// **'Filter'**
  String get filterWorkouts;

  /// No description provided for @myWorkoutsLayoutList.
  ///
  /// In en, this message translates to:
  /// **'List'**
  String get myWorkoutsLayoutList;

  /// No description provided for @myWorkoutsLayoutCards.
  ///
  /// In en, this message translates to:
  /// **'Cards'**
  String get myWorkoutsLayoutCards;

  /// No description provided for @noWorkoutsMatchSportFilter.
  ///
  /// In en, this message translates to:
  /// **'No workouts match the selected sport types'**
  String get noWorkoutsMatchSportFilter;

  /// No description provided for @welcomeDescription.
  ///
  /// In en, this message translates to:
  /// **'Workouts, equipment, and a friends feed on your own server.'**
  String get welcomeDescription;

  /// No description provided for @welcomeInstructions.
  ///
  /// In en, this message translates to:
  /// **'To get started, sign in or register.'**
  String get welcomeInstructions;

  /// No description provided for @welcomeMobileServerHint.
  ///
  /// In en, this message translates to:
  /// **'On a mobile phone, enter the Grom server address.'**
  String get welcomeMobileServerHint;

  /// No description provided for @signIn.
  ///
  /// In en, this message translates to:
  /// **'Sign in'**
  String get signIn;

  /// No description provided for @register.
  ///
  /// In en, this message translates to:
  /// **'Register'**
  String get register;

  /// No description provided for @signOut.
  ///
  /// In en, this message translates to:
  /// **'Sign out'**
  String get signOut;

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @add.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get add;

  /// No description provided for @welcomeUser.
  ///
  /// In en, this message translates to:
  /// **'Welcome, {nickname}!'**
  String welcomeUser(String nickname);

  /// No description provided for @registrationSuccessful.
  ///
  /// In en, this message translates to:
  /// **'Registration successful. Please sign in.'**
  String get registrationSuccessful;

  /// No description provided for @signedOut.
  ///
  /// In en, this message translates to:
  /// **'You have signed out'**
  String get signedOut;

  /// No description provided for @signedInAs.
  ///
  /// In en, this message translates to:
  /// **'Signed in as {nickname}'**
  String signedInAs(String nickname);

  /// No description provided for @failedToSignIn.
  ///
  /// In en, this message translates to:
  /// **'Failed to sign in'**
  String get failedToSignIn;

  /// No description provided for @failedToRegister.
  ///
  /// In en, this message translates to:
  /// **'Failed to register'**
  String get failedToRegister;

  /// No description provided for @enterEmail.
  ///
  /// In en, this message translates to:
  /// **'Enter email'**
  String get enterEmail;

  /// No description provided for @enterValidEmail.
  ///
  /// In en, this message translates to:
  /// **'Enter a valid email'**
  String get enterValidEmail;

  /// No description provided for @emailLabel.
  ///
  /// In en, this message translates to:
  /// **'Email *'**
  String get emailLabel;

  /// No description provided for @enterPassword.
  ///
  /// In en, this message translates to:
  /// **'Enter password'**
  String get enterPassword;

  /// No description provided for @passwordLabel.
  ///
  /// In en, this message translates to:
  /// **'Password *'**
  String get passwordLabel;

  /// No description provided for @enterNickname.
  ///
  /// In en, this message translates to:
  /// **'Enter nickname'**
  String get enterNickname;

  /// No description provided for @nicknameLabel.
  ///
  /// In en, this message translates to:
  /// **'Nickname *'**
  String get nicknameLabel;

  /// No description provided for @nameLabel.
  ///
  /// In en, this message translates to:
  /// **'Full name'**
  String get nameLabel;

  /// No description provided for @passwordMinLength.
  ///
  /// In en, this message translates to:
  /// **'Password must be at least 8 characters'**
  String get passwordMinLength;

  /// No description provided for @confirmPasswordLabel.
  ///
  /// In en, this message translates to:
  /// **'Confirm password *'**
  String get confirmPasswordLabel;

  /// No description provided for @confirmPassword.
  ///
  /// In en, this message translates to:
  /// **'Confirm password'**
  String get confirmPassword;

  /// No description provided for @passwordsDoNotMatch.
  ///
  /// In en, this message translates to:
  /// **'Passwords do not match'**
  String get passwordsDoNotMatch;

  /// No description provided for @forgotPasswordLink.
  ///
  /// In en, this message translates to:
  /// **'Forgot password?'**
  String get forgotPasswordLink;

  /// No description provided for @forgotPasswordTitle.
  ///
  /// In en, this message translates to:
  /// **'Reset password'**
  String get forgotPasswordTitle;

  /// No description provided for @forgotPasswordHint.
  ///
  /// In en, this message translates to:
  /// **'Enter your account email. If it is registered, we will send a reset link. Open the link in a browser to choose a new password.'**
  String get forgotPasswordHint;

  /// No description provided for @forgotPasswordSubmit.
  ///
  /// In en, this message translates to:
  /// **'Send reset link'**
  String get forgotPasswordSubmit;

  /// No description provided for @forgotPasswordCheckEmail.
  ///
  /// In en, this message translates to:
  /// **'If an account exists for that email, a reset link has been sent. Open it in a browser, then sign in here.'**
  String get forgotPasswordCheckEmail;

  /// No description provided for @forgotPasswordFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to request password reset'**
  String get forgotPasswordFailed;

  /// No description provided for @captchaRequired.
  ///
  /// In en, this message translates to:
  /// **'Complete the captcha check'**
  String get captchaRequired;

  /// No description provided for @captchaNotRobot.
  ///
  /// In en, this message translates to:
  /// **'I\'m not a robot'**
  String get captchaNotRobot;

  /// No description provided for @resetPasswordTitle.
  ///
  /// In en, this message translates to:
  /// **'Choose a new password'**
  String get resetPasswordTitle;

  /// No description provided for @resetPasswordHint.
  ///
  /// In en, this message translates to:
  /// **'Enter a new password for your account.'**
  String get resetPasswordHint;

  /// No description provided for @resetPasswordSubmit.
  ///
  /// In en, this message translates to:
  /// **'Update password'**
  String get resetPasswordSubmit;

  /// No description provided for @resetPasswordSuccess.
  ///
  /// In en, this message translates to:
  /// **'Password updated. Please sign in.'**
  String get resetPasswordSuccess;

  /// No description provided for @resetPasswordFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to reset password'**
  String get resetPasswordFailed;

  /// No description provided for @resetPasswordInvalidToken.
  ///
  /// In en, this message translates to:
  /// **'This reset link is missing or invalid.'**
  String get resetPasswordInvalidToken;

  /// No description provided for @serverUrlLabel.
  ///
  /// In en, this message translates to:
  /// **'Server URL *'**
  String get serverUrlLabel;

  /// No description provided for @enterServerUrl.
  ///
  /// In en, this message translates to:
  /// **'Enter server URL'**
  String get enterServerUrl;

  /// No description provided for @enterValidServerUrl.
  ///
  /// In en, this message translates to:
  /// **'Enter a valid server host or URL'**
  String get enterValidServerUrl;

  /// No description provided for @serverUrlHint.
  ///
  /// In en, this message translates to:
  /// **'example.com'**
  String get serverUrlHint;

  /// No description provided for @chooseServerTooltip.
  ///
  /// In en, this message translates to:
  /// **'Choose a server'**
  String get chooseServerTooltip;

  /// No description provided for @chooseServerTitle.
  ///
  /// In en, this message translates to:
  /// **'Choose a server'**
  String get chooseServerTitle;

  /// No description provided for @approvedServersSection.
  ///
  /// In en, this message translates to:
  /// **'Approved servers'**
  String get approvedServersSection;

  /// No description provided for @recentServersSection.
  ///
  /// In en, this message translates to:
  /// **'Recent servers'**
  String get recentServersSection;

  /// No description provided for @serverPickerEmpty.
  ///
  /// In en, this message translates to:
  /// **'No servers yet. Enter a URL, or sign in to remember one.'**
  String get serverPickerEmpty;

  /// No description provided for @language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;

  /// No description provided for @languageEnglish.
  ///
  /// In en, this message translates to:
  /// **'English'**
  String get languageEnglish;

  /// No description provided for @languageRussian.
  ///
  /// In en, this message translates to:
  /// **'Russian'**
  String get languageRussian;

  /// No description provided for @languageGerman.
  ///
  /// In en, this message translates to:
  /// **'German'**
  String get languageGerman;

  /// No description provided for @addWorkout.
  ///
  /// In en, this message translates to:
  /// **'Add workout'**
  String get addWorkout;

  /// No description provided for @workoutName.
  ///
  /// In en, this message translates to:
  /// **'Workout name'**
  String get workoutName;

  /// No description provided for @workoutDescription.
  ///
  /// In en, this message translates to:
  /// **'Description'**
  String get workoutDescription;

  /// No description provided for @workoutType.
  ///
  /// In en, this message translates to:
  /// **'Workout type'**
  String get workoutType;

  /// No description provided for @workoutDate.
  ///
  /// In en, this message translates to:
  /// **'Date'**
  String get workoutDate;

  /// No description provided for @workoutStartTime.
  ///
  /// In en, this message translates to:
  /// **'Start time'**
  String get workoutStartTime;

  /// No description provided for @workoutDuration.
  ///
  /// In en, this message translates to:
  /// **'Time'**
  String get workoutDuration;

  /// No description provided for @workoutDistance.
  ///
  /// In en, this message translates to:
  /// **'Distance'**
  String get workoutDistance;

  /// No description provided for @workoutPace.
  ///
  /// In en, this message translates to:
  /// **'Pace'**
  String get workoutPace;

  /// No description provided for @workoutElevationGain.
  ///
  /// In en, this message translates to:
  /// **'Elevation gain'**
  String get workoutElevationGain;

  /// No description provided for @workoutSpeedAvg.
  ///
  /// In en, this message translates to:
  /// **'Avg. speed'**
  String get workoutSpeedAvg;

  /// No description provided for @workoutSpeedMax.
  ///
  /// In en, this message translates to:
  /// **'Max. speed'**
  String get workoutSpeedMax;

  /// No description provided for @workoutSpeedChartTitle.
  ///
  /// In en, this message translates to:
  /// **'Speed'**
  String get workoutSpeedChartTitle;

  /// No description provided for @workoutHeartRateChartTitle.
  ///
  /// In en, this message translates to:
  /// **'Heart rate'**
  String get workoutHeartRateChartTitle;

  /// No description provided for @workoutTotalTime.
  ///
  /// In en, this message translates to:
  /// **'Total time'**
  String get workoutTotalTime;

  /// No description provided for @workoutHeartRateAvg.
  ///
  /// In en, this message translates to:
  /// **'Avg. heart rate'**
  String get workoutHeartRateAvg;

  /// No description provided for @workoutHeartRateMax.
  ///
  /// In en, this message translates to:
  /// **'Max. heart rate'**
  String get workoutHeartRateMax;

  /// No description provided for @chartMinutes.
  ///
  /// In en, this message translates to:
  /// **'{value} min'**
  String chartMinutes(String value);

  /// No description provided for @workoutSteps.
  ///
  /// In en, this message translates to:
  /// **'Steps'**
  String get workoutSteps;

  /// No description provided for @workoutCalories.
  ///
  /// In en, this message translates to:
  /// **'Calories'**
  String get workoutCalories;

  /// No description provided for @elevationMeters.
  ///
  /// In en, this message translates to:
  /// **'{value} m'**
  String elevationMeters(String value);

  /// No description provided for @heartRateBpm.
  ///
  /// In en, this message translates to:
  /// **'{value}'**
  String heartRateBpm(String value);

  /// No description provided for @stepsCount.
  ///
  /// In en, this message translates to:
  /// **'{value}'**
  String stepsCount(String value);

  /// No description provided for @caloriesKcal.
  ///
  /// In en, this message translates to:
  /// **'{value}'**
  String caloriesKcal(String value);

  /// No description provided for @save.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get save;

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @ok.
  ///
  /// In en, this message translates to:
  /// **'OK'**
  String get ok;

  /// No description provided for @selectWorkoutType.
  ///
  /// In en, this message translates to:
  /// **'Select workout type'**
  String get selectWorkoutType;

  /// No description provided for @enterWorkoutName.
  ///
  /// In en, this message translates to:
  /// **'Enter workout name'**
  String get enterWorkoutName;

  /// No description provided for @workoutSaved.
  ///
  /// In en, this message translates to:
  /// **'Workout saved'**
  String get workoutSaved;

  /// No description provided for @failedToSaveWorkout.
  ///
  /// In en, this message translates to:
  /// **'Failed to save workout'**
  String get failedToSaveWorkout;

  /// No description provided for @failedToLoadWorkouts.
  ///
  /// In en, this message translates to:
  /// **'Failed to load workouts'**
  String get failedToLoadWorkouts;

  /// No description provided for @failedToLoadWorkoutLikes.
  ///
  /// In en, this message translates to:
  /// **'Failed to load workout likes'**
  String get failedToLoadWorkoutLikes;

  /// No description provided for @failedToUpdateWorkoutLike.
  ///
  /// In en, this message translates to:
  /// **'Failed to update workout like'**
  String get failedToUpdateWorkoutLike;

  /// No description provided for @workoutLikeAction.
  ///
  /// In en, this message translates to:
  /// **'Like workout'**
  String get workoutLikeAction;

  /// No description provided for @workoutNoLikesYet.
  ///
  /// In en, this message translates to:
  /// **'No likes yet'**
  String get workoutNoLikesYet;

  /// No description provided for @workoutLikesTitle.
  ///
  /// In en, this message translates to:
  /// **'Likes ({count})'**
  String workoutLikesTitle(String count);

  /// No description provided for @failedToLoadWorkoutComments.
  ///
  /// In en, this message translates to:
  /// **'Failed to load workout comments'**
  String get failedToLoadWorkoutComments;

  /// No description provided for @failedToAddWorkoutComment.
  ///
  /// In en, this message translates to:
  /// **'Failed to add comment'**
  String get failedToAddWorkoutComment;

  /// No description provided for @failedToDeleteWorkoutComment.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete comment'**
  String get failedToDeleteWorkoutComment;

  /// No description provided for @workoutCommentAction.
  ///
  /// In en, this message translates to:
  /// **'Comments'**
  String get workoutCommentAction;

  /// No description provided for @workoutNoCommentsYet.
  ///
  /// In en, this message translates to:
  /// **'No comments yet'**
  String get workoutNoCommentsYet;

  /// No description provided for @workoutCommentsTitle.
  ///
  /// In en, this message translates to:
  /// **'Comments ({count})'**
  String workoutCommentsTitle(String count);

  /// No description provided for @workoutCommentHint.
  ///
  /// In en, this message translates to:
  /// **'Write a comment'**
  String get workoutCommentHint;

  /// No description provided for @addWorkoutCommentAction.
  ///
  /// In en, this message translates to:
  /// **'Add comment'**
  String get addWorkoutCommentAction;

  /// No description provided for @deleteWorkoutCommentAction.
  ///
  /// In en, this message translates to:
  /// **'Delete comment'**
  String get deleteWorkoutCommentAction;

  /// No description provided for @deleteWorkoutCommentTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete comment?'**
  String get deleteWorkoutCommentTitle;

  /// No description provided for @deleteWorkoutCommentConfirm.
  ///
  /// In en, this message translates to:
  /// **'Delete this comment?'**
  String get deleteWorkoutCommentConfirm;

  /// No description provided for @retry.
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get retry;

  /// No description provided for @expandMap.
  ///
  /// In en, this message translates to:
  /// **'Expand map'**
  String get expandMap;

  /// No description provided for @addPhotos.
  ///
  /// In en, this message translates to:
  /// **'Add photos'**
  String get addPhotos;

  /// No description provided for @photosSelected.
  ///
  /// In en, this message translates to:
  /// **'{count} photos selected'**
  String photosSelected(int count);

  /// No description provided for @removePhoto.
  ///
  /// In en, this message translates to:
  /// **'Remove photo'**
  String get removePhoto;

  /// No description provided for @failedToUploadPhotos.
  ///
  /// In en, this message translates to:
  /// **'Failed to upload photos'**
  String get failedToUploadPhotos;

  /// No description provided for @closePhotoViewer.
  ///
  /// In en, this message translates to:
  /// **'Close'**
  String get closePhotoViewer;

  /// No description provided for @collapseMap.
  ///
  /// In en, this message translates to:
  /// **'Collapse map'**
  String get collapseMap;

  /// No description provided for @noWorkoutsYet.
  ///
  /// In en, this message translates to:
  /// **'You have no workouts yet'**
  String get noWorkoutsYet;

  /// No description provided for @durationZero.
  ///
  /// In en, this message translates to:
  /// **'0s'**
  String get durationZero;

  /// No description provided for @distanceZero.
  ///
  /// In en, this message translates to:
  /// **'0 km'**
  String get distanceZero;

  /// No description provided for @durationHours.
  ///
  /// In en, this message translates to:
  /// **'{hours}h'**
  String durationHours(int hours);

  /// No description provided for @durationMinutes.
  ///
  /// In en, this message translates to:
  /// **'{minutes}m'**
  String durationMinutes(int minutes);

  /// No description provided for @durationSeconds.
  ///
  /// In en, this message translates to:
  /// **'{seconds}s'**
  String durationSeconds(int seconds);

  /// No description provided for @distanceKilometers.
  ///
  /// In en, this message translates to:
  /// **'{value} km'**
  String distanceKilometers(String value);

  /// No description provided for @distanceMeters.
  ///
  /// In en, this message translates to:
  /// **'{value} m'**
  String distanceMeters(int value);

  /// No description provided for @selectDuration.
  ///
  /// In en, this message translates to:
  /// **'Select duration'**
  String get selectDuration;

  /// No description provided for @selectDistance.
  ///
  /// In en, this message translates to:
  /// **'Select distance'**
  String get selectDistance;

  /// No description provided for @hoursLabel.
  ///
  /// In en, this message translates to:
  /// **'Hours'**
  String get hoursLabel;

  /// No description provided for @minutesLabel.
  ///
  /// In en, this message translates to:
  /// **'Minutes'**
  String get minutesLabel;

  /// No description provided for @secondsLabel.
  ///
  /// In en, this message translates to:
  /// **'Seconds'**
  String get secondsLabel;

  /// No description provided for @kilometersLabel.
  ///
  /// In en, this message translates to:
  /// **'Kilometers'**
  String get kilometersLabel;

  /// No description provided for @sportCategoryFoot.
  ///
  /// In en, this message translates to:
  /// **'Foot Sports'**
  String get sportCategoryFoot;

  /// No description provided for @sportCategoryCycle.
  ///
  /// In en, this message translates to:
  /// **'Cycle Sports'**
  String get sportCategoryCycle;

  /// No description provided for @sportCategoryStrength.
  ///
  /// In en, this message translates to:
  /// **'Strength Sports'**
  String get sportCategoryStrength;

  /// No description provided for @sportCategoryWater.
  ///
  /// In en, this message translates to:
  /// **'Water Sports'**
  String get sportCategoryWater;

  /// No description provided for @sportCategoryWinter.
  ///
  /// In en, this message translates to:
  /// **'Winter Sports'**
  String get sportCategoryWinter;

  /// No description provided for @sportCategoryTeam.
  ///
  /// In en, this message translates to:
  /// **'Team Sports'**
  String get sportCategoryTeam;

  /// No description provided for @sportCategoryRacket.
  ///
  /// In en, this message translates to:
  /// **'Racket Sports'**
  String get sportCategoryRacket;

  /// No description provided for @sportCategoryOther.
  ///
  /// In en, this message translates to:
  /// **'Other Sports'**
  String get sportCategoryOther;

  /// No description provided for @sportTypeRun.
  ///
  /// In en, this message translates to:
  /// **'Run'**
  String get sportTypeRun;

  /// No description provided for @sportTypeHike.
  ///
  /// In en, this message translates to:
  /// **'Hiking'**
  String get sportTypeHike;

  /// No description provided for @sportTypeTrailRun.
  ///
  /// In en, this message translates to:
  /// **'Trail Running'**
  String get sportTypeTrailRun;

  /// No description provided for @sportTypeWheelchair.
  ///
  /// In en, this message translates to:
  /// **'Wheelchair'**
  String get sportTypeWheelchair;

  /// No description provided for @sportTypeWalk.
  ///
  /// In en, this message translates to:
  /// **'Walk'**
  String get sportTypeWalk;

  /// No description provided for @sportTypeNordicWalk.
  ///
  /// In en, this message translates to:
  /// **'Nordic Walk'**
  String get sportTypeNordicWalk;

  /// No description provided for @sportTypeRide.
  ///
  /// In en, this message translates to:
  /// **'Ride'**
  String get sportTypeRide;

  /// No description provided for @sportTypeEBikeRide.
  ///
  /// In en, this message translates to:
  /// **'E-Bike Ride'**
  String get sportTypeEBikeRide;

  /// No description provided for @sportTypeMountainBikeRide.
  ///
  /// In en, this message translates to:
  /// **'Mountain Bike Ride'**
  String get sportTypeMountainBikeRide;

  /// No description provided for @sportTypeEMountainBikeRide.
  ///
  /// In en, this message translates to:
  /// **'E-Mountain Bike Ride'**
  String get sportTypeEMountainBikeRide;

  /// No description provided for @sportTypeGravelRide.
  ///
  /// In en, this message translates to:
  /// **'Gravel Ride'**
  String get sportTypeGravelRide;

  /// No description provided for @sportTypeVelomobile.
  ///
  /// In en, this message translates to:
  /// **'Velomobile'**
  String get sportTypeVelomobile;

  /// No description provided for @sportTypeHandcycle.
  ///
  /// In en, this message translates to:
  /// **'Handcycle'**
  String get sportTypeHandcycle;

  /// No description provided for @sportTypeCanoeing.
  ///
  /// In en, this message translates to:
  /// **'Canoe'**
  String get sportTypeCanoeing;

  /// No description provided for @sportTypeStandUpPaddling.
  ///
  /// In en, this message translates to:
  /// **'SUP'**
  String get sportTypeStandUpPaddling;

  /// No description provided for @sportTypeKayaking.
  ///
  /// In en, this message translates to:
  /// **'Kayak'**
  String get sportTypeKayaking;

  /// No description provided for @sportTypePackraft.
  ///
  /// In en, this message translates to:
  /// **'Packraft'**
  String get sportTypePackraft;

  /// No description provided for @sportTypeSurfing.
  ///
  /// In en, this message translates to:
  /// **'Surf'**
  String get sportTypeSurfing;

  /// No description provided for @sportTypeKitesurf.
  ///
  /// In en, this message translates to:
  /// **'Kitesurf'**
  String get sportTypeKitesurf;

  /// No description provided for @sportTypeSwim.
  ///
  /// In en, this message translates to:
  /// **'Swim'**
  String get sportTypeSwim;

  /// No description provided for @sportTypeRowing.
  ///
  /// In en, this message translates to:
  /// **'Rowing'**
  String get sportTypeRowing;

  /// No description provided for @sportTypeWindsurf.
  ///
  /// In en, this message translates to:
  /// **'Windsurf'**
  String get sportTypeWindsurf;

  /// No description provided for @sportTypeSail.
  ///
  /// In en, this message translates to:
  /// **'Sailing'**
  String get sportTypeSail;

  /// No description provided for @sportTypeIceSkate.
  ///
  /// In en, this message translates to:
  /// **'Ice Skate'**
  String get sportTypeIceSkate;

  /// No description provided for @sportTypeNordicSki.
  ///
  /// In en, this message translates to:
  /// **'Nordic Ski'**
  String get sportTypeNordicSki;

  /// No description provided for @sportTypeAlpineSki.
  ///
  /// In en, this message translates to:
  /// **'Alpine Ski'**
  String get sportTypeAlpineSki;

  /// No description provided for @sportTypeSnowboard.
  ///
  /// In en, this message translates to:
  /// **'Snowboard'**
  String get sportTypeSnowboard;

  /// No description provided for @sportTypeBackcountrySki.
  ///
  /// In en, this message translates to:
  /// **'Backcountry Ski'**
  String get sportTypeBackcountrySki;

  /// No description provided for @sportTypeIceHockey.
  ///
  /// In en, this message translates to:
  /// **'Ice Hockey'**
  String get sportTypeIceHockey;

  /// No description provided for @sportTypeSnowshoe.
  ///
  /// In en, this message translates to:
  /// **'Snowshoe'**
  String get sportTypeSnowshoe;

  /// No description provided for @sportTypeWorkout.
  ///
  /// In en, this message translates to:
  /// **'Workout'**
  String get sportTypeWorkout;

  /// No description provided for @sportTypeGolf.
  ///
  /// In en, this message translates to:
  /// **'Golf'**
  String get sportTypeGolf;

  /// No description provided for @sportTypeBadminton.
  ///
  /// In en, this message translates to:
  /// **'Badminton'**
  String get sportTypeBadminton;

  /// No description provided for @sportTypeElliptical.
  ///
  /// In en, this message translates to:
  /// **'Eliptical'**
  String get sportTypeElliptical;

  /// No description provided for @sportTypeBasketball.
  ///
  /// In en, this message translates to:
  /// **'Basketball'**
  String get sportTypeBasketball;

  /// No description provided for @sportTypeInlineSkate.
  ///
  /// In en, this message translates to:
  /// **'Inline Skate'**
  String get sportTypeInlineSkate;

  /// No description provided for @sportTypeSkateboard.
  ///
  /// In en, this message translates to:
  /// **'Skateboarding'**
  String get sportTypeSkateboard;

  /// No description provided for @sportTypeTennis.
  ///
  /// In en, this message translates to:
  /// **'Tennis'**
  String get sportTypeTennis;

  /// No description provided for @sportTypeStairStepper.
  ///
  /// In en, this message translates to:
  /// **'Stair Stepper'**
  String get sportTypeStairStepper;

  /// No description provided for @sportTypePadel.
  ///
  /// In en, this message translates to:
  /// **'Padel'**
  String get sportTypePadel;

  /// No description provided for @sportTypeRockClimbing.
  ///
  /// In en, this message translates to:
  /// **'Rock Climb'**
  String get sportTypeRockClimbing;

  /// No description provided for @sportTypeSoccer.
  ///
  /// In en, this message translates to:
  /// **'Football (Soccer)'**
  String get sportTypeSoccer;

  /// No description provided for @sportTypePickleball.
  ///
  /// In en, this message translates to:
  /// **'Pickleball'**
  String get sportTypePickleball;

  /// No description provided for @sportTypeWeightTraining.
  ///
  /// In en, this message translates to:
  /// **'Weight Training'**
  String get sportTypeWeightTraining;

  /// No description provided for @sportTypeVolleyball.
  ///
  /// In en, this message translates to:
  /// **'Volleyball'**
  String get sportTypeVolleyball;

  /// No description provided for @sportTypeRollerSki.
  ///
  /// In en, this message translates to:
  /// **'Roller Ski'**
  String get sportTypeRollerSki;

  /// No description provided for @sportTypeSquash.
  ///
  /// In en, this message translates to:
  /// **'Squash'**
  String get sportTypeSquash;

  /// No description provided for @sportTypeCrossfit.
  ///
  /// In en, this message translates to:
  /// **'Crossfit'**
  String get sportTypeCrossfit;

  /// No description provided for @sportTypeYoga.
  ///
  /// In en, this message translates to:
  /// **'Yoga'**
  String get sportTypeYoga;

  /// No description provided for @sportTypeDance.
  ///
  /// In en, this message translates to:
  /// **'Dance'**
  String get sportTypeDance;

  /// No description provided for @sportTypeTableTennis.
  ///
  /// In en, this message translates to:
  /// **'Table Tennis'**
  String get sportTypeTableTennis;

  /// No description provided for @sportTypePilates.
  ///
  /// In en, this message translates to:
  /// **'Pilates'**
  String get sportTypePilates;

  /// No description provided for @sportTypeRacquetball.
  ///
  /// In en, this message translates to:
  /// **'Racquetball'**
  String get sportTypeRacquetball;

  /// No description provided for @sportTypeHiit.
  ///
  /// In en, this message translates to:
  /// **'HIIT'**
  String get sportTypeHiit;

  /// No description provided for @sportTypeCricket.
  ///
  /// In en, this message translates to:
  /// **'Cricket'**
  String get sportTypeCricket;

  /// No description provided for @workoutTrack.
  ///
  /// In en, this message translates to:
  /// **'Track'**
  String get workoutTrack;

  /// No description provided for @selectTrackFile.
  ///
  /// In en, this message translates to:
  /// **'Select FIT or GPX file'**
  String get selectTrackFile;

  /// No description provided for @trackFileSelected.
  ///
  /// In en, this message translates to:
  /// **'{filename}'**
  String trackFileSelected(String filename);

  /// No description provided for @removeTrack.
  ///
  /// In en, this message translates to:
  /// **'Remove track'**
  String get removeTrack;

  /// No description provided for @invalidTrackFormat.
  ///
  /// In en, this message translates to:
  /// **'Only FIT and GPX files are supported'**
  String get invalidTrackFormat;

  /// No description provided for @failedToParseTrack.
  ///
  /// In en, this message translates to:
  /// **'Failed to read track file'**
  String get failedToParseTrack;

  /// No description provided for @trackMetadataApplied.
  ///
  /// In en, this message translates to:
  /// **'Values updated from track'**
  String get trackMetadataApplied;

  /// No description provided for @shareTrackLoginRequired.
  ///
  /// In en, this message translates to:
  /// **'Log in to import a shared track'**
  String get shareTrackLoginRequired;

  /// No description provided for @shareTrackReadFailed.
  ///
  /// In en, this message translates to:
  /// **'Could not read the shared file'**
  String get shareTrackReadFailed;

  /// No description provided for @tabRecord.
  ///
  /// In en, this message translates to:
  /// **'Record'**
  String get tabRecord;

  /// No description provided for @tabManual.
  ///
  /// In en, this message translates to:
  /// **'Manual'**
  String get tabManual;

  /// No description provided for @recordStart.
  ///
  /// In en, this message translates to:
  /// **'Record'**
  String get recordStart;

  /// No description provided for @recordPause.
  ///
  /// In en, this message translates to:
  /// **'Pause'**
  String get recordPause;

  /// No description provided for @recordFinish.
  ///
  /// In en, this message translates to:
  /// **'Finish'**
  String get recordFinish;

  /// No description provided for @recordingDuration.
  ///
  /// In en, this message translates to:
  /// **'Duration'**
  String get recordingDuration;

  /// No description provided for @currentSpeed.
  ///
  /// In en, this message translates to:
  /// **'Speed'**
  String get currentSpeed;

  /// No description provided for @speedKmh.
  ///
  /// In en, this message translates to:
  /// **'{speed} km/h'**
  String speedKmh(String speed);

  /// No description provided for @speedUnavailable.
  ///
  /// In en, this message translates to:
  /// **'—'**
  String get speedUnavailable;

  /// No description provided for @locationPermissionDenied.
  ///
  /// In en, this message translates to:
  /// **'Location permission is required to record a workout track'**
  String get locationPermissionDenied;

  /// No description provided for @notificationPermissionDenied.
  ///
  /// In en, this message translates to:
  /// **'Notification permission is required to show the recording status'**
  String get notificationPermissionDenied;

  /// No description provided for @locationServicesDisabled.
  ///
  /// In en, this message translates to:
  /// **'Enable location services to record a workout track'**
  String get locationServicesDisabled;

  /// No description provided for @openSettings.
  ///
  /// In en, this message translates to:
  /// **'Open settings'**
  String get openSettings;

  /// No description provided for @discardRecordingTitle.
  ///
  /// In en, this message translates to:
  /// **'Discard recording?'**
  String get discardRecordingTitle;

  /// No description provided for @discardRecordingMessage.
  ///
  /// In en, this message translates to:
  /// **'The current track recording will be lost.'**
  String get discardRecordingMessage;

  /// No description provided for @discardRecordingConfirm.
  ///
  /// In en, this message translates to:
  /// **'Discard'**
  String get discardRecordingConfirm;

  /// No description provided for @recordingNotificationTitle.
  ///
  /// In en, this message translates to:
  /// **'Recording workout'**
  String get recordingNotificationTitle;

  /// No description provided for @recordingNotificationText.
  ///
  /// In en, this message translates to:
  /// **'Tap to return to Grom'**
  String get recordingNotificationText;

  /// No description provided for @recordingNotificationChannelName.
  ///
  /// In en, this message translates to:
  /// **'Workout recording'**
  String get recordingNotificationChannelName;

  /// No description provided for @recordingPausedNotificationText.
  ///
  /// In en, this message translates to:
  /// **'Workout recording paused'**
  String get recordingPausedNotificationText;

  /// No description provided for @recordingAutoPausedNotificationText.
  ///
  /// In en, this message translates to:
  /// **'Workout recording auto-paused'**
  String get recordingAutoPausedNotificationText;

  /// No description provided for @autoPauseEnabled.
  ///
  /// In en, this message translates to:
  /// **'Auto-pause on'**
  String get autoPauseEnabled;

  /// No description provided for @autoPauseDisabled.
  ///
  /// In en, this message translates to:
  /// **'Auto-pause off'**
  String get autoPauseDisabled;

  /// No description provided for @backgroundLocationRationale.
  ///
  /// In en, this message translates to:
  /// **'Background location lets Grom keep recording your workout when you switch apps.'**
  String get backgroundLocationRationale;

  /// No description provided for @doNotDismissNotification.
  ///
  /// In en, this message translates to:
  /// **'Do not dismiss the notification while recording'**
  String get doNotDismissNotification;

  /// No description provided for @recordingInProgress.
  ///
  /// In en, this message translates to:
  /// **'Recording in progress'**
  String get recordingInProgress;

  /// No description provided for @restoreRecordingTitle.
  ///
  /// In en, this message translates to:
  /// **'Resume recording?'**
  String get restoreRecordingTitle;

  /// No description provided for @restoreRecordingMessage.
  ///
  /// In en, this message translates to:
  /// **'An unfinished workout recording was found. Resume it or discard the saved track.'**
  String get restoreRecordingMessage;

  /// No description provided for @restoreRecordingConfirm.
  ///
  /// In en, this message translates to:
  /// **'Resume'**
  String get restoreRecordingConfirm;

  /// No description provided for @restoreRecordingDiscard.
  ///
  /// In en, this message translates to:
  /// **'Discard'**
  String get restoreRecordingDiscard;

  /// No description provided for @editWorkout.
  ///
  /// In en, this message translates to:
  /// **'Edit'**
  String get editWorkout;

  /// No description provided for @editWorkoutTitle.
  ///
  /// In en, this message translates to:
  /// **'Edit workout'**
  String get editWorkoutTitle;

  /// No description provided for @deleteWorkout.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get deleteWorkout;

  /// No description provided for @deleteWorkoutConfirm.
  ///
  /// In en, this message translates to:
  /// **'The workout will be permanently deleted and cannot be restored.'**
  String get deleteWorkoutConfirm;

  /// No description provided for @workoutDeleted.
  ///
  /// In en, this message translates to:
  /// **'Workout deleted'**
  String get workoutDeleted;

  /// No description provided for @failedToDeleteWorkout.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete workout'**
  String get failedToDeleteWorkout;

  /// No description provided for @workoutActions.
  ///
  /// In en, this message translates to:
  /// **'Workout actions'**
  String get workoutActions;

  /// No description provided for @downloadTrackAsGpx.
  ///
  /// In en, this message translates to:
  /// **'Download track as GPX'**
  String get downloadTrackAsGpx;

  /// No description provided for @downloadTrackOriginal.
  ///
  /// In en, this message translates to:
  /// **'Download track (original)'**
  String get downloadTrackOriginal;

  /// No description provided for @downloadingTrack.
  ///
  /// In en, this message translates to:
  /// **'Downloading track…'**
  String get downloadingTrack;

  /// No description provided for @failedToDownloadTrack.
  ///
  /// In en, this message translates to:
  /// **'Failed to download track'**
  String get failedToDownloadTrack;

  /// No description provided for @trackSaved.
  ///
  /// In en, this message translates to:
  /// **'Track saved'**
  String get trackSaved;

  /// No description provided for @failedToLoadWorkoutTrack.
  ///
  /// In en, this message translates to:
  /// **'Failed to load workout track'**
  String get failedToLoadWorkoutTrack;

  /// No description provided for @userSearch.
  ///
  /// In en, this message translates to:
  /// **'User search'**
  String get userSearch;

  /// No description provided for @profile.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profile;

  /// No description provided for @searchUsersHint.
  ///
  /// In en, this message translates to:
  /// **'Nickname or @user@server'**
  String get searchUsersHint;

  /// No description provided for @search.
  ///
  /// In en, this message translates to:
  /// **'Search'**
  String get search;

  /// No description provided for @follow.
  ///
  /// In en, this message translates to:
  /// **'Follow'**
  String get follow;

  /// No description provided for @unfollow.
  ///
  /// In en, this message translates to:
  /// **'Unfollow'**
  String get unfollow;

  /// No description provided for @following.
  ///
  /// In en, this message translates to:
  /// **'Following'**
  String get following;

  /// No description provided for @followers.
  ///
  /// In en, this message translates to:
  /// **'Followers'**
  String get followers;

  /// No description provided for @followPending.
  ///
  /// In en, this message translates to:
  /// **'Pending'**
  String get followPending;

  /// No description provided for @noUsersFound.
  ///
  /// In en, this message translates to:
  /// **'No users found'**
  String get noUsersFound;

  /// No description provided for @noFollowingYet.
  ///
  /// In en, this message translates to:
  /// **'You are not following anyone yet'**
  String get noFollowingYet;

  /// No description provided for @noFollowersYet.
  ///
  /// In en, this message translates to:
  /// **'No one is following you yet'**
  String get noFollowersYet;

  /// No description provided for @searchByNicknameOrHandle.
  ///
  /// In en, this message translates to:
  /// **'Search by nickname or federated handle (@user@server)'**
  String get searchByNicknameOrHandle;

  /// No description provided for @workoutByAuthor.
  ///
  /// In en, this message translates to:
  /// **'By {author}'**
  String workoutByAuthor(String author);

  /// No description provided for @failedToSearchUsers.
  ///
  /// In en, this message translates to:
  /// **'Failed to search users'**
  String get failedToSearchUsers;

  /// No description provided for @failedToLoadProfile.
  ///
  /// In en, this message translates to:
  /// **'Failed to load profile'**
  String get failedToLoadProfile;

  /// No description provided for @editProfile.
  ///
  /// In en, this message translates to:
  /// **'Edit profile'**
  String get editProfile;

  /// No description provided for @profileActions.
  ///
  /// In en, this message translates to:
  /// **'Profile actions'**
  String get profileActions;

  /// No description provided for @deleteAccount.
  ///
  /// In en, this message translates to:
  /// **'Delete account'**
  String get deleteAccount;

  /// No description provided for @deleteAccountWarning.
  ///
  /// In en, this message translates to:
  /// **'All account data, including server login credentials, workouts, equipment, and related data, will be permanently deleted from the server.'**
  String get deleteAccountWarning;

  /// No description provided for @deleteAccountPasswordLabel.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get deleteAccountPasswordLabel;

  /// No description provided for @deleteAccountConfirm.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get deleteAccountConfirm;

  /// No description provided for @deleteAccountGoodbye.
  ///
  /// In en, this message translates to:
  /// **'Goodbye'**
  String get deleteAccountGoodbye;

  /// No description provided for @deleteAccountFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete account'**
  String get deleteAccountFailed;

  /// No description provided for @deleteAccountInvalidPassword.
  ///
  /// In en, this message translates to:
  /// **'Invalid password'**
  String get deleteAccountInvalidPassword;

  /// No description provided for @profileSaved.
  ///
  /// In en, this message translates to:
  /// **'Profile saved'**
  String get profileSaved;

  /// No description provided for @avatarUpdated.
  ///
  /// In en, this message translates to:
  /// **'Avatar updated'**
  String get avatarUpdated;

  /// No description provided for @failedToUploadAvatar.
  ///
  /// In en, this message translates to:
  /// **'Failed to upload avatar'**
  String get failedToUploadAvatar;

  /// No description provided for @cropAvatarTitle.
  ///
  /// In en, this message translates to:
  /// **'Crop avatar'**
  String get cropAvatarTitle;

  /// No description provided for @cropAvatarDone.
  ///
  /// In en, this message translates to:
  /// **'Done'**
  String get cropAvatarDone;

  /// No description provided for @failedToSaveProfile.
  ///
  /// In en, this message translates to:
  /// **'Failed to save profile'**
  String get failedToSaveProfile;

  /// No description provided for @equipment.
  ///
  /// In en, this message translates to:
  /// **'Equipment'**
  String get equipment;

  /// No description provided for @addEquipment.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get addEquipment;

  /// No description provided for @selectEquipment.
  ///
  /// In en, this message translates to:
  /// **'Select equipment'**
  String get selectEquipment;

  /// No description provided for @equipmentType.
  ///
  /// In en, this message translates to:
  /// **'Type'**
  String get equipmentType;

  /// No description provided for @equipmentName.
  ///
  /// In en, this message translates to:
  /// **'Name'**
  String get equipmentName;

  /// No description provided for @equipmentBrand.
  ///
  /// In en, this message translates to:
  /// **'Brand'**
  String get equipmentBrand;

  /// No description provided for @equipmentModel.
  ///
  /// In en, this message translates to:
  /// **'Model'**
  String get equipmentModel;

  /// No description provided for @equipmentWeight.
  ///
  /// In en, this message translates to:
  /// **'Weight (kg)'**
  String get equipmentWeight;

  /// No description provided for @equipmentNotes.
  ///
  /// In en, this message translates to:
  /// **'Notes'**
  String get equipmentNotes;

  /// No description provided for @workoutEquipment.
  ///
  /// In en, this message translates to:
  /// **'Equipment'**
  String get workoutEquipment;

  /// No description provided for @workoutDevice.
  ///
  /// In en, this message translates to:
  /// **'Device'**
  String get workoutDevice;

  /// No description provided for @bikeType.
  ///
  /// In en, this message translates to:
  /// **'Bike type'**
  String get bikeType;

  /// No description provided for @waterEquipmentType.
  ///
  /// In en, this message translates to:
  /// **'Water equipment type'**
  String get waterEquipmentType;

  /// No description provided for @deleteEquipment.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get deleteEquipment;

  /// No description provided for @deleteEquipmentConfirm.
  ///
  /// In en, this message translates to:
  /// **'Delete this equipment? It will be removed from all workouts.'**
  String get deleteEquipmentConfirm;

  /// No description provided for @noEquipmentYet.
  ///
  /// In en, this message translates to:
  /// **'You have no equipment yet'**
  String get noEquipmentYet;

  /// No description provided for @equipmentSaved.
  ///
  /// In en, this message translates to:
  /// **'Equipment saved'**
  String get equipmentSaved;

  /// No description provided for @equipmentDeleted.
  ///
  /// In en, this message translates to:
  /// **'Equipment deleted'**
  String get equipmentDeleted;

  /// No description provided for @failedToLoadEquipment.
  ///
  /// In en, this message translates to:
  /// **'Failed to load equipment'**
  String get failedToLoadEquipment;

  /// No description provided for @failedToSaveEquipment.
  ///
  /// In en, this message translates to:
  /// **'Failed to save equipment'**
  String get failedToSaveEquipment;

  /// No description provided for @enterEquipmentName.
  ///
  /// In en, this message translates to:
  /// **'Enter equipment name'**
  String get enterEquipmentName;

  /// No description provided for @equipmentTypeBike.
  ///
  /// In en, this message translates to:
  /// **'Bike'**
  String get equipmentTypeBike;

  /// No description provided for @equipmentTypeShoes.
  ///
  /// In en, this message translates to:
  /// **'Shoes'**
  String get equipmentTypeShoes;

  /// No description provided for @equipmentTypeWater.
  ///
  /// In en, this message translates to:
  /// **'Water equipment'**
  String get equipmentTypeWater;

  /// No description provided for @equipmentTypeOther.
  ///
  /// In en, this message translates to:
  /// **'Other'**
  String get equipmentTypeOther;

  /// No description provided for @equipmentSubtypeEmpty.
  ///
  /// In en, this message translates to:
  /// **'Not selected'**
  String get equipmentSubtypeEmpty;

  /// No description provided for @bikeTypeMountain.
  ///
  /// In en, this message translates to:
  /// **'Mountain'**
  String get bikeTypeMountain;

  /// No description provided for @bikeTypeGravel.
  ///
  /// In en, this message translates to:
  /// **'Gravel'**
  String get bikeTypeGravel;

  /// No description provided for @bikeTypeRoad.
  ///
  /// In en, this message translates to:
  /// **'Road'**
  String get bikeTypeRoad;

  /// No description provided for @bikeTypeTouring.
  ///
  /// In en, this message translates to:
  /// **'Touring'**
  String get bikeTypeTouring;

  /// No description provided for @bikeTypeTriathlon.
  ///
  /// In en, this message translates to:
  /// **'Triathlon'**
  String get bikeTypeTriathlon;

  /// No description provided for @bikeTypeCyclocross.
  ///
  /// In en, this message translates to:
  /// **'Cyclocross'**
  String get bikeTypeCyclocross;

  /// No description provided for @bikeTypeFixie.
  ///
  /// In en, this message translates to:
  /// **'Fixie'**
  String get bikeTypeFixie;

  /// No description provided for @bikeTypeFolding.
  ///
  /// In en, this message translates to:
  /// **'Folding'**
  String get bikeTypeFolding;

  /// No description provided for @bikeTypeBmx.
  ///
  /// In en, this message translates to:
  /// **'BMX'**
  String get bikeTypeBmx;

  /// No description provided for @waterTypeSup.
  ///
  /// In en, this message translates to:
  /// **'SUP'**
  String get waterTypeSup;

  /// No description provided for @waterTypeKayak.
  ///
  /// In en, this message translates to:
  /// **'Kayak'**
  String get waterTypeKayak;

  /// No description provided for @waterTypeCanoe.
  ///
  /// In en, this message translates to:
  /// **'Canoe'**
  String get waterTypeCanoe;

  /// No description provided for @waterTypeCanoeDouble.
  ///
  /// In en, this message translates to:
  /// **'Double canoe'**
  String get waterTypeCanoeDouble;

  /// No description provided for @waterTypePackraft.
  ///
  /// In en, this message translates to:
  /// **'Packraft'**
  String get waterTypePackraft;

  /// No description provided for @waterTypeSurf.
  ///
  /// In en, this message translates to:
  /// **'Surf'**
  String get waterTypeSurf;

  /// No description provided for @about.
  ///
  /// In en, this message translates to:
  /// **'About'**
  String get about;

  /// No description provided for @aboutAuthorLabel.
  ///
  /// In en, this message translates to:
  /// **'Author'**
  String get aboutAuthorLabel;

  /// No description provided for @aboutSourceCodeLabel.
  ///
  /// In en, this message translates to:
  /// **'Source code'**
  String get aboutSourceCodeLabel;

  /// No description provided for @aboutPrivacyPolicyLabel.
  ///
  /// In en, this message translates to:
  /// **'Privacy Policy'**
  String get aboutPrivacyPolicyLabel;

  /// No description provided for @aboutLicenseLabel.
  ///
  /// In en, this message translates to:
  /// **'License'**
  String get aboutLicenseLabel;

  /// No description provided for @mapDataAttributionTitle.
  ///
  /// In en, this message translates to:
  /// **'Map data'**
  String get mapDataAttributionTitle;

  /// No description provided for @openStreetMapAttribution.
  ///
  /// In en, this message translates to:
  /// **'© OpenStreetMap contributors'**
  String get openStreetMapAttribution;

  /// No description provided for @openStreetMapLicense.
  ///
  /// In en, this message translates to:
  /// **'Map previews and interactive maps use data from OpenStreetMap, available under the Open Database License (ODbL).'**
  String get openStreetMapLicense;

  /// No description provided for @openStreetMapCopyrightLink.
  ///
  /// In en, this message translates to:
  /// **'OpenStreetMap copyright and license'**
  String get openStreetMapCopyrightLink;

  /// No description provided for @integration.
  ///
  /// In en, this message translates to:
  /// **'Integration'**
  String get integration;

  /// No description provided for @strava.
  ///
  /// In en, this message translates to:
  /// **'Strava'**
  String get strava;

  /// No description provided for @stravaApiImportToggle.
  ///
  /// In en, this message translates to:
  /// **'Import workouts from Strava'**
  String get stravaApiImportToggle;

  /// No description provided for @stravaApiImportHelp.
  ///
  /// In en, this message translates to:
  /// **'Create your own Strava API application at strava.com/settings/api (a Strava subscription is required). Set Authorization Callback Domain to localhost. Enter your Client ID and Client Secret, then connect. Sync on Home imports up to the {limit} most recent activities and stops at the first workout already imported (same external_id as the archive). For a full history, use the Strava archive import below. Only activities visible to Everyone or Followers are imported.'**
  String stravaApiImportHelp(int limit);

  /// No description provided for @stravaApiClientIdLabel.
  ///
  /// In en, this message translates to:
  /// **'Strava client id'**
  String get stravaApiClientIdLabel;

  /// No description provided for @stravaApiClientSecretLabel.
  ///
  /// In en, this message translates to:
  /// **'Strava client secret'**
  String get stravaApiClientSecretLabel;

  /// No description provided for @stravaApiConnectStatusDisconnected.
  ///
  /// In en, this message translates to:
  /// **'Not connected'**
  String get stravaApiConnectStatusDisconnected;

  /// No description provided for @stravaApiConnectStatusConnected.
  ///
  /// In en, this message translates to:
  /// **'Connected'**
  String get stravaApiConnectStatusConnected;

  /// No description provided for @stravaApiConnectStatusFailed.
  ///
  /// In en, this message translates to:
  /// **'Connection failed'**
  String get stravaApiConnectStatusFailed;

  /// No description provided for @stravaApiConnectMissingCredentials.
  ///
  /// In en, this message translates to:
  /// **'Enter Client ID and Client Secret'**
  String get stravaApiConnectMissingCredentials;

  /// No description provided for @stravaApiConnectCancelled.
  ///
  /// In en, this message translates to:
  /// **'Strava authorization was cancelled'**
  String get stravaApiConnectCancelled;

  /// No description provided for @stravaApiConnectDenied.
  ///
  /// In en, this message translates to:
  /// **'Strava authorization was denied'**
  String get stravaApiConnectDenied;

  /// No description provided for @stravaApiConnectMissingScope.
  ///
  /// In en, this message translates to:
  /// **'Grant the activity:read permission in Strava'**
  String get stravaApiConnectMissingScope;

  /// No description provided for @stravaApiConnectError.
  ///
  /// In en, this message translates to:
  /// **'Strava connect failed: {message}'**
  String stravaApiConnectError(String message);

  /// No description provided for @stravaApiSyncing.
  ///
  /// In en, this message translates to:
  /// **'Synchronizing…'**
  String get stravaApiSyncing;

  /// No description provided for @stravaApiImported.
  ///
  /// In en, this message translates to:
  /// **'Imported {count} workouts'**
  String stravaApiImported(int count);

  /// No description provided for @stravaApiNoNewWorkouts.
  ///
  /// In en, this message translates to:
  /// **'No new workouts found'**
  String get stravaApiNoNewWorkouts;

  /// No description provided for @stravaApiNotConnected.
  ///
  /// In en, this message translates to:
  /// **'Connect with Strava on the Integration screen first'**
  String get stravaApiNotConnected;

  /// No description provided for @stravaApiNotEnabled.
  ///
  /// In en, this message translates to:
  /// **'Enable Strava API import on the Integration screen'**
  String get stravaApiNotEnabled;

  /// No description provided for @stravaApiAuthFailed.
  ///
  /// In en, this message translates to:
  /// **'Strava authentication failed'**
  String get stravaApiAuthFailed;

  /// No description provided for @stravaApiSyncCancelled.
  ///
  /// In en, this message translates to:
  /// **'Strava sync was cancelled'**
  String get stravaApiSyncCancelled;

  /// No description provided for @stravaApiSyncError.
  ///
  /// In en, this message translates to:
  /// **'Strava sync failed: {message}'**
  String stravaApiSyncError(String message);

  /// No description provided for @stravaImportDescriptionBefore.
  ///
  /// In en, this message translates to:
  /// **'You can download an archive of your workouts from the '**
  String get stravaImportDescriptionBefore;

  /// No description provided for @stravaImportDescriptionLink.
  ///
  /// In en, this message translates to:
  /// **'Strava website'**
  String get stravaImportDescriptionLink;

  /// No description provided for @stravaDownloadArchiveUrl.
  ///
  /// In en, this message translates to:
  /// **'https://www.strava.com/athlete/download_my_account'**
  String get stravaDownloadArchiveUrl;

  /// No description provided for @stravaImportDescriptionAfter.
  ///
  /// In en, this message translates to:
  /// **'. Upload the resulting ZIP archive to Grom. All workouts will be imported with tracks, equipment, and photos.'**
  String get stravaImportDescriptionAfter;

  /// No description provided for @importStravaArchive.
  ///
  /// In en, this message translates to:
  /// **'Import Strava archive'**
  String get importStravaArchive;

  /// No description provided for @uploading.
  ///
  /// In en, this message translates to:
  /// **'Uploading'**
  String get uploading;

  /// No description provided for @importing.
  ///
  /// In en, this message translates to:
  /// **'Importing'**
  String get importing;

  /// No description provided for @stravaImportCompleted.
  ///
  /// In en, this message translates to:
  /// **'Strava import completed: {imported} imported, {skipped} skipped, {parseSkipped} CSV parse skipped, {mediaMissing} media files missing from archive, {errors} errors'**
  String stravaImportCompleted(int imported, int skipped, int parseSkipped,
      int mediaMissing, int errors);

  /// No description provided for @stravaImportFailed.
  ///
  /// In en, this message translates to:
  /// **'Strava import failed: {message}'**
  String stravaImportFailed(String message);

  /// No description provided for @stravaImportInProgress.
  ///
  /// In en, this message translates to:
  /// **'Another import is already in progress'**
  String get stravaImportInProgress;

  /// No description provided for @importTracksTitle.
  ///
  /// In en, this message translates to:
  /// **'Import tracks'**
  String get importTracksTitle;

  /// No description provided for @importTracksDescription.
  ///
  /// In en, this message translates to:
  /// **'Choose one or more GPX or FIT files from your device. You can also pick files from Google Drive or other providers in the system file picker when available. Each file becomes a workout; duplicates are skipped.'**
  String get importTracksDescription;

  /// No description provided for @importTracksButton.
  ///
  /// In en, this message translates to:
  /// **'Import tracks'**
  String get importTracksButton;

  /// No description provided for @importTracksResult.
  ///
  /// In en, this message translates to:
  /// **'Import finished: {created} created, {skipped} skipped, {invalid} invalid, {failed} failed'**
  String importTracksResult(int created, int skipped, int invalid, int failed);

  /// No description provided for @integrationTabGrom.
  ///
  /// In en, this message translates to:
  /// **'Grom'**
  String get integrationTabGrom;

  /// No description provided for @integrationTabExternal.
  ///
  /// In en, this message translates to:
  /// **'External services'**
  String get integrationTabExternal;

  /// No description provided for @gromApiTitle.
  ///
  /// In en, this message translates to:
  /// **'Grom API'**
  String get gromApiTitle;

  /// No description provided for @gromApiDescription.
  ///
  /// In en, this message translates to:
  /// **'Create personal access tokens to connect external apps and scripts to your workouts and equipment.'**
  String get gromApiDescription;

  /// No description provided for @patCreateToken.
  ///
  /// In en, this message translates to:
  /// **'Create token'**
  String get patCreateToken;

  /// No description provided for @patNoTokens.
  ///
  /// In en, this message translates to:
  /// **'No personal access tokens yet'**
  String get patNoTokens;

  /// No description provided for @patNameLabel.
  ///
  /// In en, this message translates to:
  /// **'Token name'**
  String get patNameLabel;

  /// No description provided for @patScopesLabel.
  ///
  /// In en, this message translates to:
  /// **'Scopes'**
  String get patScopesLabel;

  /// No description provided for @patScopeWorkoutsRead.
  ///
  /// In en, this message translates to:
  /// **'Read workouts'**
  String get patScopeWorkoutsRead;

  /// No description provided for @patScopeWorkoutsWrite.
  ///
  /// In en, this message translates to:
  /// **'Write workouts'**
  String get patScopeWorkoutsWrite;

  /// No description provided for @patScopeEquipmentRead.
  ///
  /// In en, this message translates to:
  /// **'Read equipment'**
  String get patScopeEquipmentRead;

  /// No description provided for @patScopeEquipmentWrite.
  ///
  /// In en, this message translates to:
  /// **'Write equipment'**
  String get patScopeEquipmentWrite;

  /// No description provided for @patExpiryLabel.
  ///
  /// In en, this message translates to:
  /// **'Expiration'**
  String get patExpiryLabel;

  /// No description provided for @patExpiry90Days.
  ///
  /// In en, this message translates to:
  /// **'90 days'**
  String get patExpiry90Days;

  /// No description provided for @patExpiry180Days.
  ///
  /// In en, this message translates to:
  /// **'180 days'**
  String get patExpiry180Days;

  /// No description provided for @patExpiryCustomDays.
  ///
  /// In en, this message translates to:
  /// **'Custom (days)'**
  String get patExpiryCustomDays;

  /// No description provided for @patExpiryNone.
  ///
  /// In en, this message translates to:
  /// **'No expiration'**
  String get patExpiryNone;

  /// No description provided for @patNoExpiryWarning.
  ///
  /// In en, this message translates to:
  /// **'Tokens without expiration remain valid until you revoke them. Use only if you understand the risk.'**
  String get patNoExpiryWarning;

  /// No description provided for @patSelectScope.
  ///
  /// In en, this message translates to:
  /// **'Select at least one scope'**
  String get patSelectScope;

  /// No description provided for @patTokenCreatedTitle.
  ///
  /// In en, this message translates to:
  /// **'Token created'**
  String get patTokenCreatedTitle;

  /// No description provided for @patTokenCreatedWarning.
  ///
  /// In en, this message translates to:
  /// **'Copy this token now. You will not be able to see it again.'**
  String get patTokenCreatedWarning;

  /// No description provided for @patCopyToken.
  ///
  /// In en, this message translates to:
  /// **'Copy token'**
  String get patCopyToken;

  /// No description provided for @patTokenCopied.
  ///
  /// In en, this message translates to:
  /// **'Token copied'**
  String get patTokenCopied;

  /// No description provided for @patClose.
  ///
  /// In en, this message translates to:
  /// **'Close'**
  String get patClose;

  /// No description provided for @patRevoke.
  ///
  /// In en, this message translates to:
  /// **'Revoke'**
  String get patRevoke;

  /// No description provided for @patRevokeConfirmTitle.
  ///
  /// In en, this message translates to:
  /// **'Revoke token?'**
  String get patRevokeConfirmTitle;

  /// No description provided for @patRevokeConfirmMessage.
  ///
  /// In en, this message translates to:
  /// **'Revoke \"{name}\"? Apps using this token will lose access immediately.'**
  String patRevokeConfirmMessage(String name);

  /// No description provided for @patExpiresNever.
  ///
  /// In en, this message translates to:
  /// **'Never expires'**
  String get patExpiresNever;

  /// No description provided for @patExpiresAt.
  ///
  /// In en, this message translates to:
  /// **'Expires {date}'**
  String patExpiresAt(String date);

  /// No description provided for @patLastUsedAt.
  ///
  /// In en, this message translates to:
  /// **'Last used {date}'**
  String patLastUsedAt(String date);

  /// No description provided for @patLastUsedNever.
  ///
  /// In en, this message translates to:
  /// **'Never used'**
  String get patLastUsedNever;

  /// No description provided for @patCreatedAt.
  ///
  /// In en, this message translates to:
  /// **'Created {date}'**
  String patCreatedAt(String date);

  /// No description provided for @patFailedToLoad.
  ///
  /// In en, this message translates to:
  /// **'Failed to load tokens'**
  String get patFailedToLoad;

  /// No description provided for @patFailedToCreate.
  ///
  /// In en, this message translates to:
  /// **'Failed to create token'**
  String get patFailedToCreate;

  /// No description provided for @patFailedToRevoke.
  ///
  /// In en, this message translates to:
  /// **'Failed to revoke token'**
  String get patFailedToRevoke;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['de', 'en', 'ru'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'de':
      return AppLocalizationsDe();
    case 'en':
      return AppLocalizationsEn();
    case 'ru':
      return AppLocalizationsRu();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
