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
  String get nameLabel => 'Name';

  @override
  String get passwordMinLength => 'Password must be at least 8 characters';

  @override
  String get confirmPasswordLabel => 'Confirm password *';

  @override
  String get confirmPassword => 'Confirm password';

  @override
  String get passwordsDoNotMatch => 'Passwords do not match';

  @override
  String get language => 'Language';

  @override
  String get languageEnglish => 'English';

  @override
  String get languageRussian => 'Russian';

  @override
  String get languageGerman => 'German';
}
