// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for German (`de`).
class AppLocalizationsDe extends AppLocalizations {
  AppLocalizationsDe([String locale = 'de']) : super(locale);

  @override
  String get appTitle => 'Travka';

  @override
  String get home => 'Startseite';

  @override
  String get signIn => 'Anmelden';

  @override
  String get register => 'Registrieren';

  @override
  String get signOut => 'Abmelden';

  @override
  String get settings => 'Einstellungen';

  @override
  String get add => 'Hinzufügen';

  @override
  String welcomeUser(String nickname) {
    return 'Willkommen, $nickname!';
  }

  @override
  String get registrationSuccessful =>
      'Registrierung erfolgreich. Bitte melden Sie sich an.';

  @override
  String get signedOut => 'Sie haben sich abgemeldet';

  @override
  String signedInAs(String nickname) {
    return 'Angemeldet als $nickname';
  }

  @override
  String get failedToSignIn => 'Anmeldung fehlgeschlagen';

  @override
  String get failedToRegister => 'Registrierung fehlgeschlagen';

  @override
  String get enterEmail => 'E-Mail eingeben';

  @override
  String get enterValidEmail => 'Gültige E-Mail eingeben';

  @override
  String get emailLabel => 'E-Mail *';

  @override
  String get enterPassword => 'Passwort eingeben';

  @override
  String get passwordLabel => 'Passwort *';

  @override
  String get enterNickname => 'Nickname eingeben';

  @override
  String get nicknameLabel => 'Nickname *';

  @override
  String get nameLabel => 'Vollständiger Name';

  @override
  String get passwordMinLength =>
      'Passwort muss mindestens 8 Zeichen lang sein';

  @override
  String get confirmPasswordLabel => 'Passwort bestätigen *';

  @override
  String get confirmPassword => 'Passwort bestätigen';

  @override
  String get passwordsDoNotMatch => 'Passwörter stimmen nicht überein';

  @override
  String get serverUrlLabel => 'Server-URL *';

  @override
  String get enterServerUrl => 'Server-URL eingeben';

  @override
  String get enterValidServerUrl => 'Gültige URL eingeben (https://...)';

  @override
  String get language => 'Sprache';

  @override
  String get languageEnglish => 'English';

  @override
  String get languageRussian => 'Русский';

  @override
  String get languageGerman => 'Deutsch';

  @override
  String get addWorkout => 'Training hinzufügen';

  @override
  String get workoutName => 'Trainingsname';

  @override
  String get workoutDescription => 'Beschreibung';

  @override
  String get workoutType => 'Trainingstyp';

  @override
  String get workoutDate => 'Datum';

  @override
  String get workoutStartTime => 'Startzeit';

  @override
  String get workoutDuration => 'Dauer';

  @override
  String get workoutDistance => 'Distanz';

  @override
  String get save => 'Speichern';

  @override
  String get cancel => 'Abbrechen';

  @override
  String get selectWorkoutType => 'Trainingstyp auswählen';

  @override
  String get enterWorkoutName => 'Trainingsname eingeben';

  @override
  String get workoutSaved => 'Training gespeichert';

  @override
  String get failedToSaveWorkout => 'Training konnte nicht gespeichert werden';

  @override
  String get failedToLoadWorkouts => 'Trainings konnten nicht geladen werden';

  @override
  String get retry => 'Erneut versuchen';

  @override
  String get expandMap => 'Karte vergrößern';

  @override
  String get collapseMap => 'Karte verkleinern';

  @override
  String get noWorkoutsYet => 'Sie haben noch keine Trainings';

  @override
  String get durationZero => '0 s';

  @override
  String get distanceZero => '0 km';

  @override
  String durationHours(int hours) {
    return '$hours Std';
  }

  @override
  String durationMinutes(int minutes) {
    return '$minutes Min';
  }

  @override
  String durationSeconds(int seconds) {
    return '$seconds s';
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
  String get selectDuration => 'Dauer auswählen';

  @override
  String get selectDistance => 'Distanz auswählen';

  @override
  String get hoursLabel => 'Stunden';

  @override
  String get minutesLabel => 'Minuten';

  @override
  String get secondsLabel => 'Sekunden';

  @override
  String get kilometersLabel => 'Kilometer';

  @override
  String get sportCategoryFoot => 'Laufsport';

  @override
  String get sportCategoryCycle => 'Radsport';

  @override
  String get sportCategoryWater => 'Wassersport';

  @override
  String get sportCategoryWinter => 'Wintersport';

  @override
  String get sportCategoryOther => 'Andere Sportarten';

  @override
  String get sportTypeRun => 'Laufen';

  @override
  String get sportTypeHike => 'Wandern';

  @override
  String get sportTypeTrailRun => 'Trailrun';

  @override
  String get sportTypeWheelchair => 'Rollstuhl';

  @override
  String get sportTypeWalk => 'Gehen';

  @override
  String get sportTypeRide => 'Radfahren';

  @override
  String get sportTypeEBikeRide => 'E-Bike';

  @override
  String get sportTypeMountainBikeRide => 'Mountainbike';

  @override
  String get sportTypeEMountainBikeRide => 'E-Mountainbike';

  @override
  String get sportTypeGravelRide => 'Gravel';

  @override
  String get sportTypeVelomobile => 'Velomobil';

  @override
  String get sportTypeHandcycle => 'Handbike';

  @override
  String get sportTypeCanoeing => 'Kanu';

  @override
  String get sportTypeStandUpPaddling => 'Stand-Up-Paddling';

  @override
  String get sportTypeKayaking => 'Kajak';

  @override
  String get sportTypeSurfing => 'Surfen';

  @override
  String get sportTypeKitesurf => 'Kitesurfen';

  @override
  String get sportTypeSwim => 'Schwimmen';

  @override
  String get sportTypeRowing => 'Rudern';

  @override
  String get sportTypeWindsurf => 'Windsurfen';

  @override
  String get sportTypeSail => 'Segeln';

  @override
  String get sportTypeIceSkate => 'Eislaufen';

  @override
  String get sportTypeNordicSki => 'Skilanglauf';

  @override
  String get sportTypeAlpineSki => 'Skifahren';

  @override
  String get sportTypeSnowboard => 'Snowboard';

  @override
  String get sportTypeBackcountrySki => 'Skitour';

  @override
  String get sportTypeSnowshoe => 'Schneeschuhwandern';

  @override
  String get sportTypeWorkout => 'Workout';

  @override
  String get sportTypeGolf => 'Golf';

  @override
  String get sportTypeBadminton => 'Badminton';

  @override
  String get sportTypeElliptical => 'Crosstrainer';

  @override
  String get sportTypeBasketball => 'Basketball';

  @override
  String get sportTypeInlineSkate => 'Inlineskaten';

  @override
  String get sportTypeSkateboard => 'Skateboard';

  @override
  String get sportTypeTennis => 'Tennis';

  @override
  String get sportTypeStairStepper => 'Stepper';

  @override
  String get sportTypePadel => 'Padel';

  @override
  String get sportTypeRockClimbing => 'Klettern';

  @override
  String get sportTypeSoccer => 'Fußball';

  @override
  String get sportTypePickleball => 'Pickleball';

  @override
  String get sportTypeWeightTraining => 'Krafttraining';

  @override
  String get sportTypeVolleyball => 'Volleyball';

  @override
  String get sportTypeRollerSki => 'Rollerski';

  @override
  String get sportTypeSquash => 'Squash';

  @override
  String get sportTypeCrossfit => 'Crossfit';

  @override
  String get sportTypeYoga => 'Yoga';

  @override
  String get sportTypeDance => 'Tanz';

  @override
  String get sportTypeTableTennis => 'Tischtennis';

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
  String get selectTrackFile => 'FIT- oder GPX-Datei auswählen';

  @override
  String trackFileSelected(String filename) {
    return '$filename';
  }

  @override
  String get removeTrack => 'Track entfernen';

  @override
  String get invalidTrackFormat =>
      'Nur FIT- und GPX-Dateien werden unterstützt';

  @override
  String get failedToParseTrack => 'Trackdatei konnte nicht gelesen werden';

  @override
  String get trackMetadataApplied => 'Werte aus dem Track übernommen';

  @override
  String get tabRecord => 'Aufzeichnen';

  @override
  String get tabManual => 'Manuell';

  @override
  String get recordStart => 'Aufzeichnen';

  @override
  String get recordPause => 'Pause';

  @override
  String get recordFinish => 'Beenden';

  @override
  String get recordingDuration => 'Dauer';

  @override
  String get currentSpeed => 'Geschwindigkeit';

  @override
  String speedKmh(String speed) {
    return '$speed km/h';
  }

  @override
  String get speedUnavailable => '—';

  @override
  String get locationPermissionDenied =>
      'Standortberechtigung ist für die Trackaufzeichnung erforderlich';

  @override
  String get notificationPermissionDenied =>
      'Benachrichtigungsberechtigung ist für die Aufzeichnung erforderlich';

  @override
  String get locationServicesDisabled =>
      'Aktivieren Sie die Ortungsdienste für die Trackaufzeichnung';

  @override
  String get openSettings => 'Einstellungen öffnen';

  @override
  String get discardRecordingTitle => 'Aufzeichnung verwerfen?';

  @override
  String get discardRecordingMessage =>
      'Die aktuelle Trackaufzeichnung geht verloren.';

  @override
  String get discardRecordingConfirm => 'Verwerfen';

  @override
  String get recordingNotificationTitle => 'Training wird aufgezeichnet';

  @override
  String get recordingNotificationText => 'Tippen, um zu Travka zurückzukehren';

  @override
  String get recordingNotificationChannelName => 'Trainingsaufzeichnung';

  @override
  String get recordingPausedNotificationText =>
      'Trainingsaufzeichnung pausiert';

  @override
  String get backgroundLocationRationale =>
      'Hintergrundstandort ermöglicht die Aufzeichnung, wenn Sie die App wechseln.';

  @override
  String get doNotDismissNotification =>
      'Benachrichtigung während der Aufzeichnung nicht entfernen';

  @override
  String get recordingInProgress => 'Aufzeichnung läuft';

  @override
  String get restoreRecordingTitle => 'Aufzeichnung fortsetzen?';

  @override
  String get restoreRecordingMessage =>
      'Eine unvollständige Trainingsaufzeichnung wurde gefunden. Fortsetzen oder verwerfen?';

  @override
  String get restoreRecordingConfirm => 'Fortsetzen';

  @override
  String get restoreRecordingDiscard => 'Verwerfen';

  @override
  String get editWorkout => 'Bearbeiten';

  @override
  String get deleteWorkout => 'Löschen';

  @override
  String get workoutActions => 'Trainingsaktionen';

  @override
  String get failedToLoadWorkoutTrack =>
      'Trainingsstrecke konnte nicht geladen werden';

  @override
  String get userSearch => 'Benutzersuche';

  @override
  String get profile => 'Profil';

  @override
  String get searchUsersHint => 'Nickname oder @user@server';

  @override
  String get search => 'Suchen';

  @override
  String get follow => 'Folgen';

  @override
  String get unfollow => 'Entfolgen';

  @override
  String get following => 'Abonniert';

  @override
  String get followers => 'Follower';

  @override
  String get followPending => 'Ausstehend';

  @override
  String get noUsersFound => 'Keine Benutzer gefunden';

  @override
  String get noFollowingYet => 'Sie folgen noch niemandem';

  @override
  String get noFollowersYet => 'Ihnen folgt noch niemand';

  @override
  String get searchByNicknameOrHandle =>
      'Suche nach Nickname oder Federations-Adresse (@user@server)';

  @override
  String workoutByAuthor(String author) {
    return 'Von $author';
  }

  @override
  String get failedToSearchUsers => 'Benutzersuche fehlgeschlagen';

  @override
  String get failedToLoadProfile => 'Profil konnte nicht geladen werden';

  @override
  String get editProfile => 'Profil bearbeiten';

  @override
  String get profileSaved => 'Profil gespeichert';

  @override
  String get failedToSaveProfile => 'Profil konnte nicht gespeichert werden';

  @override
  String get equipment => 'Ausrüstung';

  @override
  String get addEquipment => 'Hinzufügen';

  @override
  String get selectEquipment => 'Ausrüstung auswählen';

  @override
  String get equipmentType => 'Typ';

  @override
  String get equipmentName => 'Bezeichnung';

  @override
  String get equipmentBrand => 'Marke';

  @override
  String get equipmentModel => 'Modell';

  @override
  String get equipmentWeight => 'Gewicht (kg)';

  @override
  String get equipmentNotes => 'Notizen';

  @override
  String get workoutEquipment => 'Ausrüstung';

  @override
  String get bikeType => 'Fahrradtyp';

  @override
  String get waterEquipmentType => 'Wasserausrüstungstyp';

  @override
  String get deleteEquipment => 'Löschen';

  @override
  String get deleteEquipmentConfirm =>
      'Diese Ausrüstung löschen? Sie wird aus allen Trainingseinheiten entfernt.';

  @override
  String get noEquipmentYet => 'Sie haben noch keine Ausrüstung';

  @override
  String get equipmentSaved => 'Ausrüstung gespeichert';

  @override
  String get equipmentDeleted => 'Ausrüstung gelöscht';

  @override
  String get failedToLoadEquipment => 'Ausrüstung konnte nicht geladen werden';

  @override
  String get failedToSaveEquipment =>
      'Ausrüstung konnte nicht gespeichert werden';

  @override
  String get enterEquipmentName => 'Bezeichnung eingeben';

  @override
  String get equipmentTypeBike => 'Fahrrad';

  @override
  String get equipmentTypeShoes => 'Schuhe';

  @override
  String get equipmentTypeWater => 'Wasserausrüstung';

  @override
  String get equipmentTypeOther => 'Sonstiges';

  @override
  String get equipmentSubtypeEmpty => 'Nicht ausgewählt';

  @override
  String get bikeTypeMountain => 'Mountainbike';

  @override
  String get bikeTypeGravel => 'Gravelbike';

  @override
  String get bikeTypeRoad => 'Rennrad';

  @override
  String get bikeTypeTouring => 'Tourenrad';

  @override
  String get bikeTypeTriathlon => 'Triathlonrad';

  @override
  String get bikeTypeCyclocross => 'Cyclocross';

  @override
  String get bikeTypeFixie => 'Fixie';

  @override
  String get bikeTypeBmx => 'BMX';

  @override
  String get waterTypeSup => 'SUP';

  @override
  String get waterTypeKayak => 'Kajak';

  @override
  String get waterTypeCanoe => 'Kanu';

  @override
  String get waterTypeCanoeDouble => 'Zweier-Kanu';

  @override
  String get waterTypePackraft => 'Packraft';

  @override
  String get waterTypeSurf => 'Surfen';
}
