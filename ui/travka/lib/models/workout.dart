class Workout {
  Workout({
    required this.id,
    required this.name,
    required this.description,
    required this.sportType,
    required this.startDate,
    required this.durationSeconds,
    required this.distance,
  });

  final String id;
  final String name;
  final String description;
  final String sportType;
  final DateTime startDate;
  final int durationSeconds;
  final double distance;

  double get distanceKm => distance / 1000;

  factory Workout.fromJson(Map<String, dynamic> json) {
    return Workout(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String? ?? '',
      sportType: json['sport_type'] as String,
      startDate: DateTime.parse(json['start_date'] as String),
      durationSeconds: json['duration_seconds'] as int? ?? 0,
      distance: (json['distance'] as num?)?.toDouble() ?? 0,
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
  });

  final String name;
  final String description;
  final String sportType;
  final DateTime startDate;
  final int durationSeconds;
  final double distanceKm;

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      if (description.isNotEmpty) 'description': description,
      'sport_type': sportType,
      'start_date': startDate.toUtc().toIso8601String(),
      'duration_seconds': durationSeconds,
      'distance': distanceKm * 1000,
    };
  }
}
