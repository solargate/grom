import 'package:travka/l10n/app_localizations.dart';

String sportTypeLocalization(AppLocalizations l10n, String id) {
  switch (id) {
    case 'Run':
      return l10n.sportTypeRun;
    case 'Hike':
      return l10n.sportTypeHike;
    case 'TrailRun':
      return l10n.sportTypeTrailRun;
    case 'Wheelchair':
      return l10n.sportTypeWheelchair;
    case 'Walk':
      return l10n.sportTypeWalk;
    case 'Ride':
      return l10n.sportTypeRide;
    case 'EBikeRide':
      return l10n.sportTypeEBikeRide;
    case 'MountainBikeRide':
      return l10n.sportTypeMountainBikeRide;
    case 'EMountainBikeRide':
      return l10n.sportTypeEMountainBikeRide;
    case 'GravelRide':
      return l10n.sportTypeGravelRide;
    case 'Velomobile':
      return l10n.sportTypeVelomobile;
    case 'Handcycle':
      return l10n.sportTypeHandcycle;
    case 'Canoeing':
      return l10n.sportTypeCanoeing;
    case 'StandUpPaddling':
      return l10n.sportTypeStandUpPaddling;
    case 'Kayaking':
      return l10n.sportTypeKayaking;
    case 'Surfing':
      return l10n.sportTypeSurfing;
    case 'Kitesurf':
      return l10n.sportTypeKitesurf;
    case 'Swim':
      return l10n.sportTypeSwim;
    case 'Rowing':
      return l10n.sportTypeRowing;
    case 'Windsurf':
      return l10n.sportTypeWindsurf;
    case 'Sail':
      return l10n.sportTypeSail;
    case 'IceSkate':
      return l10n.sportTypeIceSkate;
    case 'NordicSki':
      return l10n.sportTypeNordicSki;
    case 'AlpineSki':
      return l10n.sportTypeAlpineSki;
    case 'Snowboard':
      return l10n.sportTypeSnowboard;
    case 'BackcountrySki':
      return l10n.sportTypeBackcountrySki;
    case 'Snowshoe':
      return l10n.sportTypeSnowshoe;
    case 'Workout':
      return l10n.sportTypeWorkout;
    case 'Golf':
      return l10n.sportTypeGolf;
    case 'Badminton':
      return l10n.sportTypeBadminton;
    case 'Elliptical':
      return l10n.sportTypeElliptical;
    case 'Basketball':
      return l10n.sportTypeBasketball;
    case 'InlineSkate':
      return l10n.sportTypeInlineSkate;
    case 'Skateboard':
      return l10n.sportTypeSkateboard;
    case 'Tennis':
      return l10n.sportTypeTennis;
    case 'StairStepper':
      return l10n.sportTypeStairStepper;
    case 'Padel':
      return l10n.sportTypePadel;
    case 'RockClimbing':
      return l10n.sportTypeRockClimbing;
    case 'Soccer':
      return l10n.sportTypeSoccer;
    case 'Pickleball':
      return l10n.sportTypePickleball;
    case 'WeightTraining':
      return l10n.sportTypeWeightTraining;
    case 'Volleyball':
      return l10n.sportTypeVolleyball;
    case 'RollerSki':
      return l10n.sportTypeRollerSki;
    case 'Squash':
      return l10n.sportTypeSquash;
    case 'Crossfit':
      return l10n.sportTypeCrossfit;
    case 'Yoga':
      return l10n.sportTypeYoga;
    case 'Dance':
      return l10n.sportTypeDance;
    case 'TableTennis':
      return l10n.sportTypeTableTennis;
    case 'Pilates':
      return l10n.sportTypePilates;
    case 'Racquetball':
      return l10n.sportTypeRacquetball;
    case 'HighIntensityIntervalTraining':
      return l10n.sportTypeHiit;
    case 'Cricket':
      return l10n.sportTypeCricket;
    default:
      return id;
  }
}

String formatDuration(AppLocalizations l10n, int totalSeconds) {
  if (totalSeconds <= 0) {
    return l10n.durationZero;
  }
  final hours = totalSeconds ~/ 3600;
  final minutes = (totalSeconds % 3600) ~/ 60;
  final seconds = totalSeconds % 60;

  final parts = <String>[];
  if (hours > 0) {
    parts.add(l10n.durationHours(hours));
  }
  if (minutes > 0 || hours > 0) {
    parts.add(l10n.durationMinutes(minutes));
  }
  if (hours == 0 && seconds > 0 && minutes == 0) {
    parts.add(l10n.durationSeconds(seconds));
  }
  return parts.join(' ');
}

String formatDistance(AppLocalizations l10n, double meters) {
  if (meters <= 0) {
    return l10n.distanceZero;
  }
  final km = meters / 1000;
  if (km >= 10) {
    return l10n.distanceKilometers(km.toStringAsFixed(1));
  }
  if (km >= 1) {
    return l10n.distanceKilometers(km.toStringAsFixed(2));
  }
  return l10n.distanceMeters(meters.round());
}
