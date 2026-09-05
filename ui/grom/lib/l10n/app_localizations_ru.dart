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
  String get filterWorkouts => 'Фильтр';

  @override
  String get myWorkoutsLayoutList => 'Список';

  @override
  String get myWorkoutsLayoutCards => 'Карточки';

  @override
  String get noWorkoutsMatchSportFilter =>
      'Нет тренировок по выбранным видам спорта';

  @override
  String get welcomeDescription =>
      'Тренировки, снаряжение и лента друзей на своём сервере.';

  @override
  String get welcomeInstructions =>
      'Чтобы начать, войдите или зарегистрируйтесь.';

  @override
  String get welcomeMobileServerHint =>
      'На мобильном телефоне укажите адрес сервера Grom.';

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
  String get forgotPasswordLink => 'Забыли пароль?';

  @override
  String get forgotPasswordTitle => 'Сброс пароля';

  @override
  String get forgotPasswordHint =>
      'Введите email аккаунта. Если он зарегистрирован, мы отправим ссылку для сброса. Откройте её в браузере, чтобы задать новый пароль.';

  @override
  String get forgotPasswordSubmit => 'Отправить ссылку';

  @override
  String get forgotPasswordCheckEmail =>
      'Если аккаунт с таким email есть, ссылка для сброса отправлена. Откройте её в браузере, затем войдите здесь.';

  @override
  String get forgotPasswordFailed => 'Не удалось запросить сброс пароля';

  @override
  String get captchaRequired => 'Пройдите проверку «я не робот»';

  @override
  String get captchaNotRobot => 'Я не робот';

  @override
  String get resetPasswordTitle => 'Новый пароль';

  @override
  String get resetPasswordHint => 'Введите новый пароль для аккаунта.';

  @override
  String get resetPasswordSubmit => 'Сохранить пароль';

  @override
  String get resetPasswordSuccess => 'Пароль обновлён. Войдите в аккаунт.';

  @override
  String get resetPasswordFailed => 'Не удалось сбросить пароль';

  @override
  String get resetPasswordInvalidToken =>
      'Ссылка для сброса отсутствует или недействительна.';

  @override
  String get serverUrlLabel => 'URL сервера *';

  @override
  String get enterServerUrl => 'Введите URL сервера';

  @override
  String get enterValidServerUrl => 'Введите корректный хост или URL сервера';

  @override
  String get serverUrlHint => 'example.com';

  @override
  String get chooseServerTooltip => 'Выбрать сервер';

  @override
  String get chooseServerTitle => 'Выбор сервера';

  @override
  String get approvedServersSection => 'Одобренные серверы';

  @override
  String get recentServersSection => 'Недавние серверы';

  @override
  String get serverPickerEmpty =>
      'Пока нет серверов. Введите URL или войдите, чтобы запомнить свой.';

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
  String get workoutSpeedMax => 'Максимальная скорость';

  @override
  String get workoutSpeedChartTitle => 'Скорость';

  @override
  String get workoutHeartRateChartTitle => 'Пульс';

  @override
  String get workoutTotalTime => 'Общее время';

  @override
  String get workoutHeartRateAvg => 'Средний пульс';

  @override
  String get workoutHeartRateMax => 'Максимальный пульс';

  @override
  String chartMinutes(String value) {
    return '$value мин';
  }

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
  String get ok => 'OK';

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
  String get failedToLoadWorkoutLikes =>
      'Не удалось загрузить лайки тренировки';

  @override
  String get failedToUpdateWorkoutLike => 'Не удалось обновить лайк тренировки';

  @override
  String get workoutLikeAction => 'Лайкнуть тренировку';

  @override
  String get workoutNoLikesYet => 'Пока нет лайков';

  @override
  String workoutLikesTitle(String count) {
    return 'Лайки ($count)';
  }

  @override
  String get failedToLoadWorkoutComments => 'Не удалось загрузить комментарии';

  @override
  String get failedToAddWorkoutComment => 'Не удалось добавить комментарий';

  @override
  String get failedToDeleteWorkoutComment => 'Не удалось удалить комментарий';

  @override
  String get workoutCommentAction => 'Комментарии';

  @override
  String get workoutNoCommentsYet => 'Пока нет комментариев';

  @override
  String workoutCommentsTitle(String count) {
    return 'Комментарии ($count)';
  }

  @override
  String get workoutCommentHint => 'Напишите комментарий';

  @override
  String get addWorkoutCommentAction => 'Добавить комментарий';

  @override
  String get deleteWorkoutCommentAction => 'Удалить комментарий';

  @override
  String get deleteWorkoutCommentTitle => 'Удалить комментарий?';

  @override
  String get deleteWorkoutCommentConfirm => 'Удалить этот комментарий?';

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
  String get sportCategoryStrength => 'Силовые виды';

  @override
  String get sportCategoryWater => 'Водные виды';

  @override
  String get sportCategoryWinter => 'Зимние виды';

  @override
  String get sportCategoryTeam => 'Командные виды';

  @override
  String get sportCategoryRacket => 'Ракеточные виды';

  @override
  String get sportCategoryOther => 'Другие виды';

  @override
  String get sportTypeRun => 'Бег';

  @override
  String get sportTypeHike => 'Хайкинг';

  @override
  String get sportTypeTrailRun => 'Трейлраннинг';

  @override
  String get sportTypeWheelchair => 'Кресло-коляска';

  @override
  String get sportTypeWalk => 'Ходьба';

  @override
  String get sportTypeNordicWalk => 'Скандинавская ходьба';

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
  String get sportTypeHandcycle => 'Ручной велосипед';

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
  String get sportTypeIceHockey => 'Хоккей';

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
  String get sportTypeRollerSki => 'Роликовые лыжи';

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
  String get profileActions => 'Действия с профилем';

  @override
  String get deleteAccount => 'Удалить аккаунт';

  @override
  String get deleteAccountWarning =>
      'Все данные аккаунта, включая данные для входа на сервер, данные о тренировках, снаряжении и т.д., будут безвозвратно удалены с сервера.';

  @override
  String get deleteAccountPasswordLabel => 'Пароль';

  @override
  String get deleteAccountConfirm => 'Удалить';

  @override
  String get deleteAccountGoodbye => 'До свидания';

  @override
  String get deleteAccountFailed => 'Не удалось удалить аккаунт';

  @override
  String get deleteAccountInvalidPassword => 'Неверный пароль';

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
  String get aboutPrivacyPolicyLabel => 'Политика конфиденциальности';

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
  String get stravaApiImportToggle => 'Импорт тренировок из Strava';

  @override
  String stravaApiImportHelp(int limit) {
    return 'Создайте своё API-приложение Strava на strava.com/settings/api (нужна подписка Strava). Укажите Authorization Callback Domain = localhost. Введите Client ID и Client Secret, затем подключитесь. Синхронизация на Главной импортирует до $limit последних активностей и останавливается на первой уже импортированной тренировке (тот же external_id, что у архива). Для полной истории используйте импорт архива ниже. Импортируются только активности, видимые Everyone или Followers.';
  }

  @override
  String get stravaApiClientIdLabel => 'Strava client id';

  @override
  String get stravaApiClientSecretLabel => 'Strava client secret';

  @override
  String get stravaApiConnectStatusDisconnected => 'Не подключено';

  @override
  String get stravaApiConnectStatusConnected => 'Подключено';

  @override
  String get stravaApiConnectStatusFailed => 'Ошибка подключения';

  @override
  String get stravaApiConnectMissingCredentials =>
      'Введите Client ID и Client Secret';

  @override
  String get stravaApiConnectCancelled => 'Авторизация Strava отменена';

  @override
  String get stravaApiConnectDenied => 'Авторизация Strava отклонена';

  @override
  String get stravaApiConnectMissingScope =>
      'Выдайте разрешение activity:read в Strava';

  @override
  String stravaApiConnectError(String message) {
    return 'Не удалось подключить Strava: $message';
  }

  @override
  String get stravaApiSyncing => 'Синхронизация…';

  @override
  String stravaApiImported(int count) {
    return 'Импортировано тренировок: $count';
  }

  @override
  String get stravaApiNoNewWorkouts => 'Новых тренировок не найдено';

  @override
  String get stravaApiNotConnected =>
      'Сначала подключите Strava на экране «Интеграция»';

  @override
  String get stravaApiNotEnabled =>
      'Включите импорт через Strava API на экране «Интеграция»';

  @override
  String get stravaApiAuthFailed => 'Ошибка аутентификации Strava';

  @override
  String get stravaApiSyncCancelled => 'Синхронизация Strava отменена';

  @override
  String stravaApiSyncError(String message) {
    return 'Ошибка синхронизации Strava: $message';
  }

  @override
  String get stravaImportDescriptionBefore =>
      'Вы можете выгрузить архив своих тренировок на ';

  @override
  String get stravaImportDescriptionLink => 'сайте Strava';

  @override
  String get stravaDownloadArchiveUrl =>
      'https://www.strava.com/athlete/download_my_account';

  @override
  String get stravaImportDescriptionAfter =>
      '. Полученный zip-архив можно загрузить в Grom. Все тренировки будут импортированы с треками, снаряжением и фотографиями.';

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

  @override
  String get importTracksTitle => 'Импорт треков';

  @override
  String get importTracksDescription =>
      'Выберите один или несколько файлов GPX или FIT на устройстве. В системном выборе файлов при необходимости можно открыть Google Drive или другой источник. Каждый файл становится тренировкой; дубликаты пропускаются.';

  @override
  String get importTracksButton => 'Импортировать треки';

  @override
  String importTracksResult(int created, int skipped, int invalid, int failed) {
    return 'Импорт завершён: создано $created, пропущено $skipped, неверных $invalid, ошибок $failed';
  }

  @override
  String get integrationTabGrom => 'Grom';

  @override
  String get integrationTabExternal => 'Внешние сервисы';

  @override
  String get gromApiTitle => 'Grom API';

  @override
  String get gromApiDescription =>
      'Создавайте персональные токены доступа для подключения внешних приложений и скриптов к вашим тренировкам и снаряжению.';

  @override
  String get patCreateToken => 'Создать токен';

  @override
  String get patNoTokens => 'Персональных токенов пока нет';

  @override
  String get patNameLabel => 'Название токена';

  @override
  String get patScopesLabel => 'Права доступа';

  @override
  String get patScopeWorkoutsRead => 'Чтение тренировок';

  @override
  String get patScopeWorkoutsWrite => 'Запись тренировок';

  @override
  String get patScopeEquipmentRead => 'Чтение снаряжения';

  @override
  String get patScopeEquipmentWrite => 'Запись снаряжения';

  @override
  String get patExpiryLabel => 'Срок действия';

  @override
  String get patExpiry90Days => '90 дней';

  @override
  String get patExpiry180Days => '180 дней';

  @override
  String get patExpiryCustomDays => 'Свой срок (дни)';

  @override
  String get patExpiryNone => 'Без срока';

  @override
  String get patNoExpiryWarning =>
      'Токены без срока действуют до отзыва. Используйте только если понимаете риски.';

  @override
  String get patSelectScope => 'Выберите хотя бы одно право доступа';

  @override
  String get patTokenCreatedTitle => 'Токен создан';

  @override
  String get patTokenCreatedWarning =>
      'Скопируйте токен сейчас. Повторно он не будет показан.';

  @override
  String get patCopyToken => 'Скопировать токен';

  @override
  String get patTokenCopied => 'Токен скопирован';

  @override
  String get patClose => 'Закрыть';

  @override
  String get patRevoke => 'Отозвать';

  @override
  String get patRevokeConfirmTitle => 'Отозвать токен?';

  @override
  String patRevokeConfirmMessage(String name) {
    return 'Отозвать «$name»? Приложения с этим токеном сразу потеряют доступ.';
  }

  @override
  String get patExpiresNever => 'Без срока';

  @override
  String patExpiresAt(String date) {
    return 'Истекает $date';
  }

  @override
  String patLastUsedAt(String date) {
    return 'Использовался $date';
  }

  @override
  String get patLastUsedNever => 'Не использовался';

  @override
  String patCreatedAt(String date) {
    return 'Создан $date';
  }

  @override
  String get patFailedToLoad => 'Не удалось загрузить токены';

  @override
  String get patFailedToCreate => 'Не удалось создать токен';

  @override
  String get patFailedToRevoke => 'Не удалось отозвать токен';
}
