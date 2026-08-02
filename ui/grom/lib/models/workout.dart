import 'social.dart';

class WorkoutEquipmentItem {
  WorkoutEquipmentItem({
    required this.id,
    required this.name,
    this.type = '',
  });

  final String id;
  final String name;
  final String type;

  factory WorkoutEquipmentItem.fromJson(Map<String, dynamic> json) {
    return WorkoutEquipmentItem(
      id: json['id'] as String,
      name: json['name'] as String,
      type: json['type'] as String? ?? '',
    );
  }
}

class Workout {
  Workout({
    required this.id,
    required this.name,
    required this.description,
    required this.sportType,
    required this.startDate,
    required this.durationSeconds,
    required this.distance,
    this.durationTotalSeconds,
    this.tempAvgKmm,
    this.speedMaxKmh,
    this.speedAvgKmh,
    this.elevationGain,
    this.heartRateAvg,
    this.heartRateMax,
    this.stepsTotal,
    this.calories,
    this.owner = '',
    this.device = '',
    this.track = '',
    this.hasMapPreview = false,
    this.hasMedia = false,
    this.mediaFiles = const [],
    this.author,
    this.equipment = const [],
  });

  final String id;
  final String owner;
  final String name;
  final String description;
  final String sportType;
  final DateTime startDate;
  final int durationSeconds;
  final double distance;
  final int? durationTotalSeconds;
  final String? tempAvgKmm;
  final double? speedMaxKmh;
  final double? speedAvgKmh;
  final double? elevationGain;
  final double? heartRateAvg;
  final double? heartRateMax;
  final int? stepsTotal;
  final double? calories;
  final String device;
  final String track;
  final bool hasMapPreview;
  final bool hasMedia;
  final List<String> mediaFiles;
  final WorkoutAuthor? author;
  final List<WorkoutEquipmentItem> equipment;

  double get distanceKm => distance / 1000;

  String get ownerNickname => owner.isNotEmpty ? owner : (author?.nickname ?? '');

  factory Workout.fromJson(Map<String, dynamic> json) {
    WorkoutAuthor? author;
    final authorJson = json['author'];
    if (authorJson is Map<String, dynamic>) {
      author = WorkoutAuthor.fromJson(authorJson);
    }
    final equipmentJson = json['equipment'];
    final equipment = <WorkoutEquipmentItem>[];
    if (equipmentJson is List) {
      for (final item in equipmentJson) {
        if (item is Map<String, dynamic>) {
          equipment.add(WorkoutEquipmentItem.fromJson(item));
        }
      }
    }
    return Workout(
      id: json['id'] as String,
      owner: json['owner'] as String? ?? author?.nickname ?? '',
      name: json['name'] as String,
      description: json['description'] as String? ?? '',
      sportType: json['sport_type'] as String,
      startDate: DateTime.parse(json['start_date'] as String),
      durationSeconds: json['duration_seconds'] as int? ?? 0,
      distance: (json['distance'] as num?)?.toDouble() ?? 0,
      durationTotalSeconds: json['duration_total_seconds'] as int?,
      tempAvgKmm: json['temp_avg_kmm'] as String?,
      speedMaxKmh: (json['speed_max_kmh'] as num?)?.toDouble(),
      speedAvgKmh: (json['speed_avg_kmh'] as num?)?.toDouble(),
      elevationGain: (json['elevation_gain'] as num?)?.toDouble(),
      heartRateAvg: (json['heart_rate_avg'] as num?)?.toDouble(),
      heartRateMax: (json['heart_rate_max'] as num?)?.toDouble(),
      stepsTotal: json['steps_total'] as int?,
      calories: (json['calories'] as num?)?.toDouble(),
      device: json['device'] as String? ?? '',
      track: json['track'] as String? ?? '',
      hasMapPreview: json['has_map_preview'] as bool? ?? false,
      hasMedia: json['has_media'] as bool? ?? false,
      mediaFiles: (json['media_files'] as List<dynamic>?)
              ?.map((item) => item.toString())
              .toList() ??
          const [],
      author: author,
      equipment: equipment,
    );
  }

  Map<String, dynamic> toCreateJson() {
    return {
      'name': name,
      if (description.isNotEmpty) 'description': description,
      'sport_type': sportType,
      'start_date': startDate.toUtc().toIso8601String(),
      'duration_seconds': durationSeconds,
      'distance': distance,
    };
  }
}

class CreateWorkoutDraft {
  CreateWorkoutDraft({
    required this.name,
    required this.description,
    required this.sportType,
    required this.startDate,
    required this.durationSeconds,
    this.durationTotalSeconds,
    required this.distanceKm,
    this.speedMaxKmh,
    this.speedAvgKmh,
    this.equipmentIds = const [],
  });

  final String name;
  final String description;
  final String sportType;
  final DateTime startDate;
  final int durationSeconds;
  final int? durationTotalSeconds;
  final double distanceKm;
  final double? speedMaxKmh;
  final double? speedAvgKmh;
  final List<String> equipmentIds;

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      if (description.isNotEmpty) 'description': description,
      'sport_type': sportType,
      'start_date': startDate.toUtc().toIso8601String(),
      'duration_seconds': durationSeconds,
      if (durationTotalSeconds != null)
        'duration_total_seconds': durationTotalSeconds,
      'distance': distanceKm * 1000,
      if (speedMaxKmh != null) 'speed_max_kmh': speedMaxKmh,
      if (speedAvgKmh != null) 'speed_avg_kmh': speedAvgKmh,
      'equipment_ids': equipmentIds,
    };
  }
}

class WorkoutListPage {
  WorkoutListPage({
    required this.items,
    this.nextCursor,
    this.hasMore = false,
  });

  final List<Workout> items;
  final String? nextCursor;
  final bool hasMore;

  factory WorkoutListPage.fromJson(Map<String, dynamic> json) {
    final rawItems = json['items'];
    final items = <Workout>[];
    if (rawItems is List) {
      for (final item in rawItems) {
        if (item is Map<String, dynamic>) {
          items.add(Workout.fromJson(item));
        }
      }
    }
    final cursor = json['next_cursor'];
    return WorkoutListPage(
      items: items,
      nextCursor: cursor is String && cursor.isNotEmpty ? cursor : null,
      hasMore: json['has_more'] as bool? ?? false,
    );
  }
}
