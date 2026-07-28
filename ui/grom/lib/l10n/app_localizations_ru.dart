// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Russian (`ru`).
class AppLocalizationsRu extends AppLocalizations {
  AppLocalizationsRu([String locale = 'ru']) : super(locale);

  @override
  String get appTitle => 'Grom';

  @override
  String get home => 'Главная';

  @override
  String get homeTabFeed => 'Лента';

  @override
  String get homeTabMyWorkouts => 'Мои тренировки';

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
  String get nameLabel => 'Полное имя';

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
  String get workoutDuration => 'Время';

  @override
  String get workoutDistance => 'Расстояние';

  @override
  String get workoutPace => 'Темп';

  @override
  String get workoutElevationGain => 'Набор высоты';

  @override
  String get workoutSpeedAvg => 'Средняя скорость';

  @override
  String get workoutTotalTime => 'Общее время';

  @override
  String get workoutHeartRateAvg => 'Средний пульс';

  @override
  String get workoutSteps => 'Шаги';

  @override
  String get workoutCalories => 'Калории';

  @override
  String elevationMeters(String value) {
    return '$value м';
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
  String get expandMap => 'Расширить карту';

  @override
  String get addPhotos => 'Добавить фотографии';

  @override
  String photosSelected(int count) {
    return 'Выбрано фотографий: $count';
  }

  @override
  String get removePhoto => 'Удалить фото';

  @override
  String get failedToUploadPhotos => 'Не удалось загрузить фотографии';

  @override
  String get closePhotoViewer => 'Закрыть';

  @override
  String get collapseMap => 'Уменьшить карту';

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
  String get sportTypeEBikeRide => 'Электровелосипед';

  @override
  String get sportTypeMountainBikeRide => 'Горный велосипед';

  @override
  String get sportTypeEMountainBikeRide => 'Горный электровелосипед';

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
  String get sportTypePackraft => 'Пакрафт';

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
  String get sportTypeHiit => 'HIIT';

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
  String get shareTrackLoginRequired =>
      'Войдите, чтобы импортировать переданный трек';

  @override
  String get shareTrackReadFailed => 'Не удалось прочитать переданный файл';

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
  String get recordingNotificationText => 'Нажмите, чтобы вернуться в Grom';

  @override
  String get recordingNotificationChannelName => 'Запись тренировки';

  @override
  String get recordingPausedNotificationText => 'Запись тренировки на паузе';

  @override
  String get recordingAutoPausedNotificationText =>
      'Запись тренировки на автопаузе';

  @override
  String get autoPauseEnabled => 'Автопауза включена';

  @override
  String get autoPauseDisabled => 'Автопауза выключена';

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

  @override
  String get editWorkout => 'Изменить';

  @override
  String get editWorkoutTitle => 'Изменить тренировку';

  @override
  String get deleteWorkout => 'Удалить';

  @override
  String get deleteWorkoutConfirm =>
      'Тренировка будет удалена окончательно, без возможности восстановления.';

  @override
  String get workoutDeleted => 'Тренировка удалена';

  @override
  String get failedToDeleteWorkout => 'Не удалось удалить тренировку';

  @override
  String get workoutActions => 'Действия с тренировкой';

  @override
  String get downloadTrackAsGpx => 'Скачать трек как GPX';

  @override
  String get downloadTrackOriginal => 'Скачать трек (оригинал)';

  @override
  String get downloadingTrack => 'Скачивание трека…';

  @override
  String get failedToDownloadTrack => 'Не удалось скачать трек';

  @override
  String get trackSaved => 'Трек сохранён';

  @override
  String get failedToLoadWorkoutTrack => 'Не удалось загрузить трек тренировки';

  @override
  String get userSearch => 'Поиск пользователей';

  @override
  String get profile => 'Профиль';

  @override
  String get searchUsersHint => 'Никнейм или @user@server';

  @override
  String get search => 'Поиск';

  @override
  String get follow => 'Подписаться';

  @override
  String get unfollow => 'Отписаться';

  @override
  String get following => 'Подписки';

  @override
  String get followers => 'Подписчики';

  @override
  String get followPending => 'Ожидание';

  @override
  String get noUsersFound => 'Пользователи не найдены';

  @override
  String get noFollowingYet => 'Вы ни на кого не подписаны';

  @override
  String get noFollowersYet => 'На вас пока никто не подписан';

  @override
  String get searchByNicknameOrHandle =>
      'Поиск по никнейму или федеративному адресу (@user@server)';

  @override
  String workoutByAuthor(String author) {
    return 'Автор: $author';
  }

  @override
  String get failedToSearchUsers => 'Не удалось выполнить поиск пользователей';

  @override
  String get failedToLoadProfile => 'Не удалось загрузить профиль';

  @override
  String get editProfile => 'Редактирование профиля';

  @override
  String get profileSaved => 'Профиль сохранён';

  @override
  String get avatarUpdated => 'Аватар обновлён';

  @override
  String get failedToUploadAvatar => 'Не удалось загрузить аватар';

  @override
  String get cropAvatarTitle => 'Обрезка аватарки';

  @override
  String get cropAvatarDone => 'Готово';

  @override
  String get failedToSaveProfile => 'Не удалось сохранить профиль';

  @override
  String get equipment => 'Снаряжение';

  @override
  String get addEquipment => 'Добавить';

  @override
  String get selectEquipment => 'Выбрать снаряжение';

  @override
  String get equipmentType => 'Тип';

  @override
  String get equipmentName => 'Наименование';

  @override
  String get equipmentBrand => 'Марка';

  @override
  String get equipmentModel => 'Модель';

  @override
  String get equipmentWeight => 'Вес (кг)';

  @override
  String get equipmentNotes => 'Примечания';

  @override
  String get workoutEquipment => 'Снаряжение';

  @override
  String get workoutDevice => 'Устройство';

  @override
  String get bikeType => 'Тип велосипеда';

  @override
  String get waterEquipmentType => 'Тип водного оборудования';

  @override
  String get deleteEquipment => 'Удалить';

  @override
  String get deleteEquipmentConfirm =>
      'Удалить это снаряжение? Оно будет убрано из всех тренировок.';

  @override
  String get noEquipmentYet => 'У вас пока нет снаряжения';

  @override
  String get equipmentSaved => 'Снаряжение сохранено';

  @override
  String get equipmentDeleted => 'Снаряжение удалено';

  @override
  String get failedToLoadEquipment => 'Не удалось загрузить снаряжение';

  @override
  String get failedToSaveEquipment => 'Не удалось сохранить снаряжение';

  @override
  String get enterEquipmentName => 'Введите наименование';

  @override
  String get equipmentTypeBike => 'Велосипед';

  @override
  String get equipmentTypeShoes => 'Обувь';

  @override
  String get equipmentTypeWater => 'Водное оборудование';

  @override
  String get equipmentTypeOther => 'Другое';

  @override
  String get equipmentSubtypeEmpty => 'Не выбрано';

  @override
  String get bikeTypeMountain => 'Горный';

  @override
  String get bikeTypeGravel => 'Гравийный';

  @override
  String get bikeTypeRoad => 'Шоссейный';

  @override
  String get bikeTypeTouring => 'Туристический';

  @override
  String get bikeTypeTriathlon => 'Разделочный';

  @override
  String get bikeTypeCyclocross => 'Кроссовый';

  @override
  String get bikeTypeFixie => 'Фикс';

  @override
  String get bikeTypeFolding => 'Складной';

  @override
  String get bikeTypeBmx => 'BMX';

  @override
  String get waterTypeSup => 'САП';

  @override
  String get waterTypeKayak => 'Каяк';

  @override
  String get waterTypeCanoe => 'Каноэ';

  @override
  String get waterTypeCanoeDouble => 'Байдарка';

  @override
  String get waterTypePackraft => 'Пакрафт';

  @override
  String get waterTypeSurf => 'Сёрф';

  @override
  String get about => 'О приложении';

  @override
  String get aboutAuthorLabel => 'Автор';

  @override
  String get aboutSourceCodeLabel => 'Исходный код';

  @override
  String get aboutLicenseLabel => 'Лицензия';

  @override
  String get mapDataAttributionTitle => 'Картографические данные';

  @override
  String get openStreetMapAttribution => '© Участники OpenStreetMap';

  @override
  String get openStreetMapLicense =>
      'Превью и интерактивные карты используют данные OpenStreetMap, доступные по лицензии Open Database License (ODbL).';

  @override
  String get openStreetMapCopyrightLink =>
      'Авторские права и лицензия OpenStreetMap';

  @override
  String get integration => 'Интеграция';

  @override
  String get strava => 'Strava';

  @override
  String get importStravaArchive => 'Импорт архива Strava';

  @override
  String get uploading => 'Загрузка';

  @override
  String get importing => 'Импорт';

  @override
  String stravaImportCompleted(int imported, int skipped, int parseSkipped,
      int mediaMissing, int errors) {
    return 'Импорт Strava завершён: импортировано $imported, пропущено $skipped, не разобрано в CSV $parseSkipped, отсутствует медиа в архиве $mediaMissing, ошибок $errors';
  }

  @override
  String stravaImportFailed(String message) {
    return 'Импорт Strava не удался: $message';
  }

  @override
  String get stravaImportInProgress => 'Импорт уже выполняется';
}
