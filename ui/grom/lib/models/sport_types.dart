import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:material_symbols_icons/symbols.dart';

enum SportCategory {
  foot,
  cycle,
  strength,
  water,
  winter,
  team,
  racket,
  other,
}

class SportTypeInfo {
  const SportTypeInfo({
    required this.id,
    required this.category,
    required this.icon,
  });

  final String id;
  final SportCategory category;
  final IconData icon;
}

const defaultSportTypeId = 'Run';

const sportTypeCatalog = <SportTypeInfo>[
  // foot
  SportTypeInfo(id: 'Run', category: SportCategory.foot, icon: Icons.directions_run),
  SportTypeInfo(id: 'Walk', category: SportCategory.foot, icon: Icons.directions_walk),
  SportTypeInfo(id: 'NordicWalk', category: SportCategory.foot, icon: Icons.nordic_walking),
  SportTypeInfo(id: 'Hike', category: SportCategory.foot, icon: Icons.hiking),
  SportTypeInfo(id: 'TrailRun', category: SportCategory.foot, icon: Icons.directions_run),
  SportTypeInfo(id: 'Wheelchair', category: SportCategory.foot, icon: Icons.accessible),
  // cycle
  SportTypeInfo(id: 'Ride', category: SportCategory.cycle, icon: Icons.directions_bike),
  SportTypeInfo(id: 'EBikeRide', category: SportCategory.cycle, icon: Icons.electric_bike),
  SportTypeInfo(id: 'MountainBikeRide', category: SportCategory.cycle, icon: Icons.pedal_bike),
  SportTypeInfo(id: 'EMountainBikeRide', category: SportCategory.cycle, icon: Icons.electric_bolt),
  SportTypeInfo(id: 'GravelRide', category: SportCategory.cycle, icon: Icons.directions_bike),
  SportTypeInfo(id: 'Velomobile', category: SportCategory.cycle, icon: Icons.sports_motorsports),
  SportTypeInfo(id: 'Handcycle', category: SportCategory.cycle, icon: Icons.two_wheeler),
  // strength
  SportTypeInfo(id: 'Workout', category: SportCategory.strength, icon: Symbols.physical_therapy),
  SportTypeInfo(id: 'WeightTraining', category: SportCategory.strength, icon: Symbols.exercise),
  SportTypeInfo(id: 'HIIT', category: SportCategory.strength, icon: Icons.local_fire_department),
  SportTypeInfo(id: 'Crossfit', category: SportCategory.strength, icon: Icons.fitness_center),
  // water
  SportTypeInfo(id: 'Canoeing', category: SportCategory.water, icon: Icons.kayaking),
  SportTypeInfo(id: 'SUP', category: SportCategory.water, icon: Icons.surfing),
  SportTypeInfo(id: 'Kayaking', category: SportCategory.water, icon: Icons.kayaking),
  SportTypeInfo(id: 'Packraft', category: SportCategory.water, icon: Icons.kayaking),
  SportTypeInfo(id: 'Surfing', category: SportCategory.water, icon: Icons.surfing),
  SportTypeInfo(id: 'Kitesurf', category: SportCategory.water, icon: Icons.kitesurfing),
  SportTypeInfo(id: 'Swim', category: SportCategory.water, icon: Icons.pool),
  SportTypeInfo(id: 'Rowing', category: SportCategory.water, icon: Icons.rowing),
  SportTypeInfo(id: 'Windsurf', category: SportCategory.water, icon: Icons.sailing),
  SportTypeInfo(id: 'Sail', category: SportCategory.water, icon: Icons.sailing),
  // winter
  SportTypeInfo(id: 'IceSkate', category: SportCategory.winter, icon: Icons.ice_skating),
  SportTypeInfo(id: 'NordicSki', category: SportCategory.winter, icon: Icons.downhill_skiing),
  SportTypeInfo(id: 'AlpineSki', category: SportCategory.winter, icon: Icons.downhill_skiing),
  SportTypeInfo(id: 'Snowboard', category: SportCategory.winter, icon: Icons.snowboarding),
  SportTypeInfo(id: 'BackcountrySki', category: SportCategory.winter, icon: Icons.downhill_skiing),
  SportTypeInfo(id: 'IceHockey', category: SportCategory.winter, icon: Icons.sports_hockey),
  SportTypeInfo(id: 'Snowshoe', category: SportCategory.winter, icon: Icons.snowshoeing),
  // team
  SportTypeInfo(id: 'Soccer', category: SportCategory.team, icon: Icons.sports_soccer),
  SportTypeInfo(id: 'Basketball', category: SportCategory.team, icon: Icons.sports_basketball),
  SportTypeInfo(id: 'Volleyball', category: SportCategory.team, icon: Icons.sports_volleyball),
  SportTypeInfo(id: 'Cricket', category: SportCategory.team, icon: Icons.sports_cricket),
  // racket
  SportTypeInfo(id: 'Tennis', category: SportCategory.racket, icon: Icons.sports_tennis),
  SportTypeInfo(id: 'Padel', category: SportCategory.racket, icon: Icons.sports_tennis),
  SportTypeInfo(id: 'Pickleball', category: SportCategory.racket, icon: Symbols.pickleball),
  SportTypeInfo(id: 'Racquetball', category: SportCategory.racket, icon: Icons.sports_tennis),
  SportTypeInfo(id: 'Squash', category: SportCategory.racket, icon: Icons.sports_tennis),
  SportTypeInfo(id: 'Badminton', category: SportCategory.racket, icon: Symbols.badminton),
  SportTypeInfo(id: 'TableTennis', category: SportCategory.racket, icon: Icons.sports_tennis),
  // other
  SportTypeInfo(id: 'Golf', category: SportCategory.other, icon: Icons.golf_course),
  SportTypeInfo(id: 'Elliptical', category: SportCategory.other, icon: Icons.monitor_heart),
  SportTypeInfo(id: 'InlineSkate', category: SportCategory.other, icon: Icons.roller_skating),
  SportTypeInfo(id: 'Skateboard', category: SportCategory.other, icon: Icons.skateboarding),
  SportTypeInfo(id: 'StairStepper', category: SportCategory.other, icon: Icons.stairs),
  SportTypeInfo(id: 'RockClimbing', category: SportCategory.other, icon: Icons.landscape),
  SportTypeInfo(id: 'RollerSki', category: SportCategory.other, icon: Icons.nordic_walking),
  SportTypeInfo(id: 'Yoga', category: SportCategory.other, icon: Icons.self_improvement),
  SportTypeInfo(id: 'Dance', category: SportCategory.other, icon: Icons.music_note),
  SportTypeInfo(id: 'Pilates', category: SportCategory.other, icon: Icons.self_improvement),
];

SportTypeInfo? sportTypeById(String id) {
  for (final sportType in sportTypeCatalog) {
    if (sportType.id == id) {
      return sportType;
    }
  }
  return null;
}

String sportCategoryLabel(AppLocalizations l10n, SportCategory category) {
  switch (category) {
    case SportCategory.foot:
      return l10n.sportCategoryFoot;
    case SportCategory.cycle:
      return l10n.sportCategoryCycle;
    case SportCategory.strength:
      return l10n.sportCategoryStrength;
    case SportCategory.water:
      return l10n.sportCategoryWater;
    case SportCategory.winter:
      return l10n.sportCategoryWinter;
    case SportCategory.team:
      return l10n.sportCategoryTeam;
    case SportCategory.racket:
      return l10n.sportCategoryRacket;
    case SportCategory.other:
      return l10n.sportCategoryOther;
  }
}

String sportTypeLabel(AppLocalizations l10n, String id) {
  return sportTypeLocalization(l10n, id);
}

Color sportTypeColor(String id) {
  final info = sportTypeById(id);
  if (info == null) {
    return Colors.grey;
  }
  switch (info.category) {
    case SportCategory.foot:
      return const Color(0xFFFC4C02);
    case SportCategory.cycle:
      return const Color(0xFF1E88E5);
    case SportCategory.strength:
      return const Color(0xFF8A9440);
    case SportCategory.water:
      return const Color(0xFF00897B);
    case SportCategory.winter:
      return const Color(0xFF5C6BC0);
    case SportCategory.team:
      return const Color(0xFFC45C5C);
    case SportCategory.racket:
      return const Color(0xFF5E8A5E);
    case SportCategory.other:
      return const Color(0xFF8E24AA);
  }
}
