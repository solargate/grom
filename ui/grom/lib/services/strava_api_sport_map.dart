import '../models/sport_types.dart';

/// Maps Strava `sport_type` / `type` strings to Grom sport type ids.
///
/// API responses use English identifiers (e.g. `Run`, `TrailRun`, `EBikeRide`).
String mapStravaSportType(String? raw, {String? activityName}) {
  final key = _normalize(raw);
  if (key.isEmpty) {
    return _inferFromName(activityName) ?? defaultSportTypeId;
  }

  final mapped = _aliases[key];
  if (mapped != null) {
    if (mapped == 'Workout') {
      final inferred = _inferFromName(activityName);
      if (inferred != null) {
        return inferred;
      }
    }
    return mapped;
  }

  final compact = key.replaceAll(' ', '');
  final rawLower = (raw ?? '').trim().toLowerCase();
  for (final sport in sportTypeCatalog) {
    final idLower = sport.id.toLowerCase();
    if (idLower == key || idLower == compact || idLower == rawLower) {
      return sport.id;
    }
  }

  return _inferFromName(activityName) ?? 'Workout';
}

String _normalize(String? raw) {
  if (raw == null) {
    return '';
  }
  var value = raw.toLowerCase().trim();
  value = value.replaceAll('\u00a0', ' ').replaceAll('\u202f', ' ');
  return value.split(RegExp(r'\s+')).where((p) => p.isNotEmpty).join(' ');
}

String? _inferFromName(String? name) {
  final key = _normalize(name);
  if (key.isEmpty) {
    return null;
  }
  for (final item in _nameKeywords) {
    if (key.contains(item.$1)) {
      return item.$2;
    }
  }
  return null;
}

const _nameKeywords = <(String, String)>[
  ('packraft', 'Packraft'),
  ('pilates', 'Pilates'),
  ('yoga', 'Yoga'),
  ('crossfit', 'Crossfit'),
  ('hiit', 'HIIT'),
  ('stretching', 'Yoga'),
  ('dance', 'Dance'),
  ('elliptical', 'Elliptical'),
  ('stair stepper', 'StairStepper'),
];

const _aliases = <String, String>{
  'run': 'Run',
  'ride': 'Ride',
  'walk': 'Walk',
  'nordic walk': 'NordicWalk',
  'nordic walking': 'NordicWalk',
  'hike': 'Hike',
  'swim': 'Swim',
  'workout': 'Workout',
  'weighttraining': 'WeightTraining',
  'weight training': 'WeightTraining',
  'virtualride': 'Ride',
  'virtual ride': 'Ride',
  'virtualrun': 'Run',
  'virtual run': 'Run',
  'ebikeride': 'EBikeRide',
  'e-bike ride': 'EBikeRide',
  'emountainbikeride': 'EMountainBikeRide',
  'mountainbikeride': 'MountainBikeRide',
  'mountain bike ride': 'MountainBikeRide',
  'gravelride': 'GravelRide',
  'gravel ride': 'GravelRide',
  'trailrun': 'TrailRun',
  'trail run': 'TrailRun',
  'canoeing': 'Canoeing',
  'kayaking': 'Kayaking',
  'packraft': 'Packraft',
  'standuppaddling': 'SUP',
  'stand up paddling': 'SUP',
  'stand-up paddling': 'SUP',
  'surfing': 'Surfing',
  'kitesurf': 'Kitesurf',
  'rowing': 'Rowing',
  'windsurf': 'Windsurf',
  'sail': 'Sail',
  'iceskate': 'IceSkate',
  'ice skate': 'IceSkate',
  'nordicski': 'NordicSki',
  'nordic ski': 'NordicSki',
  'alpineski': 'AlpineSki',
  'alpine ski': 'AlpineSki',
  'snowboard': 'Snowboard',
  'backcountryski': 'BackcountrySki',
  'backcountry ski': 'BackcountrySki',
  'icehockey': 'IceHockey',
  'ice hockey': 'IceHockey',
  'hockey': 'IceHockey',
  'snowshoe': 'Snowshoe',
  'golf': 'Golf',
  'badminton': 'Badminton',
  'elliptical': 'Elliptical',
  'basketball': 'Basketball',
  'inlineskate': 'InlineSkate',
  'inline skate': 'InlineSkate',
  'skateboard': 'Skateboard',
  'tennis': 'Tennis',
  'stairstepper': 'StairStepper',
  'stair stepper': 'StairStepper',
  'padel': 'Padel',
  'rockclimbing': 'RockClimbing',
  'rock climbing': 'RockClimbing',
  'soccer': 'Soccer',
  'pickleball': 'Pickleball',
  'volleyball': 'Volleyball',
  'rollerski': 'RollerSki',
  'roller ski': 'RollerSki',
  'squash': 'Squash',
  'crossfit': 'Crossfit',
  'yoga': 'Yoga',
  'dance': 'Dance',
  'tabletennis': 'TableTennis',
  'table tennis': 'TableTennis',
  'pilates': 'Pilates',
  'racquetball': 'Racquetball',
  'hiit': 'HIIT',
  'highintensityintervaltraining': 'HIIT',
  'high intensity interval training': 'HIIT',
  'cricket': 'Cricket',
  'wheelchair': 'Wheelchair',
  'handcycle': 'Handcycle',
  'velomobile': 'Velomobile',
};
