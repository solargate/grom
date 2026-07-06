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
  String get settings => 'Настройки';

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
  String get serverUrlLabel => 'URL сервера *';

  @override
  String get enterServerUrl => 'Введите URL сервера';

  @override
  String get enterValidServerUrl => 'Введите корректный URL (https://...)';

  @override
  String get language => 'Язык';

  @override
  String get languageEnglish => 'English';

  @override
  String get languageRussian => 'Русский';

  @override
  String get languageGerman => 'Deutsch';

  @override
  String get addWorkout => 'Добавить тренировку';

  @override
  String get workoutName => 'Название тренировки';

  @override
  String get workoutDescription => 'Описание';

  @override
  String get workoutType => 'Тип тренировки';

  @override
  String get workoutDate => 'Дата';

  @override
  String get workoutStartTime => 'Время начала';

  @override
  String get workoutDuration => 'Длительность';

  @override
  String get workoutDistance => 'Расстояние';

  @override
  String get save => 'Сохранить';

  @override
  String get cancel => 'Отмена';

  @override
  String get selectWorkoutType => 'Выберите тип тренировки';

  @override
  String get enterWorkoutName => 'Введите название тренировки';

  @override
  String get workoutSaved => 'Тренировка сохранена';

  @override
  String get failedToSaveWorkout => 'Не удалось сохранить тренировку';

  @override
  String get failedToLoadWorkouts => 'Не удалось загрузить тренировки';

  @override
  String get retry => 'Повторить';

  @override
  String get noWorkoutsYet => 'У вас пока нет тренировок';

  @override
  String get durationZero => '0 с';

  @override
  String get distanceZero => '0 км';

  @override
  String durationHours(int hours) {
    return '$hours ч';
  }

  @override
  String durationMinutes(int minutes) {
    return '$minutes мин';
  }

  @override
  String durationSeconds(int seconds) {
    return '$seconds с';
  }

  @override
  String distanceKilometers(String value) {
    return '$value км';
  }

  @override
  String distanceMeters(int value) {
    return '$value м';
  }

  @override
  String get selectDuration => 'Выберите длительность';

  @override
  String get selectDistance => 'Выберите расстояние';

  @override
  String get hoursLabel => 'Часы';

  @override
  String get minutesLabel => 'Минуты';

  @override
  String get secondsLabel => 'Секунды';

  @override
  String get kilometersLabel => 'Километры';

  @override
  String get sportCategoryFoot => 'Бег и ходьба';

  @override
  String get sportCategoryCycle => 'Велоспорт';

  @override
  String get sportCategoryWater => 'Водные виды';

  @override
  String get sportCategoryWinter => 'Зимние виды';

  @override
  String get sportCategoryOther => 'Другие виды';

  @override
  String get sportTypeRun => 'Бег';

  @override
  String get sportTypeHike => 'Поход';

  @override
  String get sportTypeTrailRun => 'Трейл';

  @override
  String get sportTypeWheelchair => 'Инвалидная коляска';

  @override
  String get sportTypeWalk => 'Ходьба';

  @override
  String get sportTypeRide => 'Велосипед';

  @override
  String get sportTypeEBikeRide => 'Э-велосипед';

  @override
  String get sportTypeMountainBikeRide => 'Горный велосипед';

  @override
  String get sportTypeEMountainBikeRide => 'Э-горный велосипед';

  @override
  String get sportTypeGravelRide => 'Гревел';

  @override
  String get sportTypeVelomobile => 'Веломобиль';

  @override
  String get sportTypeHandcycle => 'Хендбайк';

  @override
  String get sportTypeCanoeing => 'Каноэ';

  @override
  String get sportTypeStandUpPaddling => 'САП';

  @override
  String get sportTypeKayaking => 'Каяк';

  @override
  String get sportTypeSurfing => 'Сёрфинг';

  @override
  String get sportTypeKitesurf => 'Кайтсёрфинг';

  @override
  String get sportTypeSwim => 'Плавание';

  @override
  String get sportTypeRowing => 'Гребля';

  @override
  String get sportTypeWindsurf => 'Виндсёрфинг';

  @override
  String get sportTypeSail => 'Парусный спорт';

  @override
  String get sportTypeIceSkate => 'Коньки';

  @override
  String get sportTypeNordicSki => 'Лыжи (классика)';

  @override
  String get sportTypeAlpineSki => 'Горные лыжи';

  @override
  String get sportTypeSnowboard => 'Сноуборд';

  @override
  String get sportTypeBackcountrySki => 'Бэккантри';

  @override
  String get sportTypeSnowshoe => 'Снегоступы';

  @override
  String get sportTypeWorkout => 'Тренировка';

  @override
  String get sportTypeGolf => 'Гольф';

  @override
  String get sportTypeBadminton => 'Бадминтон';

  @override
  String get sportTypeElliptical => 'Эллипс';

  @override
  String get sportTypeBasketball => 'Баскетбол';

  @override
  String get sportTypeInlineSkate => 'Ролики';

  @override
  String get sportTypeSkateboard => 'Скейтборд';

  @override
  String get sportTypeTennis => 'Теннис';

  @override
  String get sportTypeStairStepper => 'Степпер';

  @override
  String get sportTypePadel => 'Падел';

  @override
  String get sportTypeRockClimbing => 'Скалолазание';

  @override
  String get sportTypeSoccer => 'Футбол';

  @override
  String get sportTypePickleball => 'Пиклбол';

  @override
  String get sportTypeWeightTraining => 'Силовая тренировка';

  @override
  String get sportTypeVolleyball => 'Волейбол';

  @override
  String get sportTypeRollerSki => 'Роллеры';

  @override
  String get sportTypeSquash => 'Сквош';

  @override
  String get sportTypeCrossfit => 'Кроссфит';

  @override
  String get sportTypeYoga => 'Йога';

  @override
  String get sportTypeDance => 'Танцы';

  @override
  String get sportTypeTableTennis => 'Настольный теннис';

  @override
  String get sportTypePilates => 'Пилатес';

  @override
  String get sportTypeRacquetball => 'Ракетбол';

  @override
  String get sportTypeHiit => 'ВИИТ';

  @override
  String get sportTypeCricket => 'Крикет';

  @override
  String get workoutTrack => 'Трек';

  @override
  String get selectTrackFile => 'Выберите файл FIT или GPX';

  @override
  String trackFileSelected(String filename) {
    return '$filename';
  }

  @override
  String get removeTrack => 'Удалить трек';

  @override
  String get invalidTrackFormat => 'Поддерживаются только файлы FIT и GPX';

  @override
  String get failedToParseTrack => 'Не удалось прочитать файл трека';

  @override
  String get trackMetadataApplied => 'Значения обновлены из трека';

  @override
  String get tabRecord => 'Запись';

  @override
  String get tabManual => 'Вручную';

  @override
  String get recordStart => 'Запись';

  @override
  String get recordPause => 'Пауза';

  @override
  String get recordFinish => 'Завершить';

  @override
  String get recordingDuration => 'Длительность';

  @override
  String get currentSpeed => 'Скорость';

  @override
  String speedKmh(String speed) {
    return '$speed км/ч';
  }

  @override
  String get speedUnavailable => '—';

  @override
  String get locationPermissionDenied =>
      'Для записи трека нужен доступ к геолокации';

  @override
  String get notificationPermissionDenied =>
      'Для записи трека нужно разрешение на показ уведомлений';

  @override
  String get locationServicesDisabled => 'Включите геолокацию для записи трека';

  @override
  String get openSettings => 'Открыть настройки';

  @override
  String get discardRecordingTitle => 'Прервать запись?';

  @override
  String get discardRecordingMessage => 'Текущая запись трека будет потеряна.';

  @override
  String get discardRecordingConfirm => 'Прервать';

  @override
  String get recordingNotificationTitle => 'Идёт запись тренировки';

  @override
  String get recordingNotificationText => 'Нажмите, чтобы вернуться в Travka';

  @override
  String get recordingNotificationChannelName => 'Запись тренировки';

  @override
  String get recordingPausedNotificationText => 'Запись тренировки на паузе';

  @override
  String get backgroundLocationRationale =>
      'Фоновая геолокация позволяет продолжать запись, когда вы переключаетесь на другие приложения.';

  @override
  String get doNotDismissNotification =>
      'Не убирайте уведомление во время записи';

  @override
  String get recordingInProgress => 'Идёт запись';

  @override
  String get restoreRecordingTitle => 'Восстановить запись?';

  @override
  String get restoreRecordingMessage =>
      'Найдена незавершённая запись тренировки. Восстановить её или удалить сохранённый трек?';

  @override
  String get restoreRecordingConfirm => 'Восстановить';

  @override
  String get restoreRecordingDiscard => 'Удалить';
}
