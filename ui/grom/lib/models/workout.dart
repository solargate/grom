import 'social.dart';

class WorkoutEquipmentItem {
  WorkoutEquipmentItem({
    required this.id,
    required this.name,
  });

  final String id;
  final String name;

  factory WorkoutEquipmentItem.fromJson(Map<String, dynamic> json) {
    return WorkoutEquipmentItem(
      id: json['id'] as String,
      name: json['name'] as String,
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
    this.owner = '',
    this.device = '',
    this.track = '',
    this.hasMapPreview = false,
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
  final String device;
  final String track;
  final bool hasMapPreview;
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
      device: json['device'] as String? ?? '',
      track: json['track'] as String? ?? '',
      hasMapPreview: json['has_map_preview'] as bool? ?? false,
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
    required this.distanceKm,
    this.equipmentIds = const [],
  });

  final String name;
  final String description;
  final String sportType;
  final DateTime startDate;
  final int durationSeconds;
  final double distanceKm;
  final List<String> equipmentIds;

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      if (description.isNotEmpty) 'description': description,
      'sport_type': sportType,
      'start_date': startDate.toUtc().toIso8601String(),
      'duration_seconds': durationSeconds,
      'distance': distanceKm * 1000,
      if (equipmentIds.isNotEmpty) 'equipment_ids': equipmentIds,
    };
  }
}
