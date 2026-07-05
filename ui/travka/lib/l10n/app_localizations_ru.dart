// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Russian (`ru`).
class AppLocalizationsRu extends AppLocalizations {
  AppLocalizationsRu([String locale = 'ru']) : super(locale);

  @override
  String get appTitle => 'Travka';

  @override
  String get home => 'Главная';

  @override
  String get signIn => 'Вход';

  @override
  String get register => 'Регистрация';

  @override
  String get signOut => 'Выйти';

  @override
  String get add => 'Добавить';

  @override
  String welcomeUser(String nickname) {
    return 'Добро пожаловать, $nickname!';
  }

  @override
  String get registrationSuccessful =>
      'Регистрация успешна. Войдите в аккаунт.';

  @override
  String get signedOut => 'Вы вышли из аккаунта';

  @override
  String signedInAs(String nickname) {
    return 'Вы вошли как $nickname';
  }

  @override
  String get failedToSignIn => 'Не удалось выполнить вход';

  @override
  String get failedToRegister => 'Не удалось выполнить регистрацию';

  @override
  String get enterEmail => 'Введите email';

  @override
  String get enterValidEmail => 'Введите корректный email';

  @override
  String get emailLabel => 'Email *';

  @override
  String get enterPassword => 'Введите пароль';

  @override
  String get passwordLabel => 'Пароль *';

  @override
  String get enterNickname => 'Введите nickname';

  @override
  String get nicknameLabel => 'Nickname *';

  @override
  String get nameLabel => 'Имя';

  @override
  String get passwordMinLength => 'Пароль должен быть не короче 8 символов';

  @override
  String get confirmPasswordLabel => 'Подтверждение пароля *';

  @override
  String get confirmPassword => 'Подтвердите пароль';

  @override
  String get passwordsDoNotMatch => 'Пароли не совпадают';

  @override
  String get language => 'Язык';

  @override
  String get languageEnglish => 'English';

  @override
  String get languageRussian => 'Русский';

  @override
  String get languageGerman => 'Deutsch';
}
