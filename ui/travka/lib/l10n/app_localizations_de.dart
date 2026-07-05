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
  String get nameLabel => 'Name';

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
  String get language => 'Sprache';

  @override
  String get languageEnglish => 'English';

  @override
  String get languageRussian => 'Русский';

  @override
  String get languageGerman => 'Deutsch';
}
