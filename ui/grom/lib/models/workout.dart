import 'social.dart';

class WorkoutLikeUser {
  WorkoutLikeUser({
    required this.handle,
    required this.nickname,
    required this.name,
    required this.isLocal,
    this.hasAvatar = false,
    this.avatarUrl,
  });

  final String handle;
  final String nickname;
  final String name;
  final bool isLocal;
  final bool hasAvatar;
  final String? avatarUrl;

  factory WorkoutLikeUser.fromJson(Map<String, dynamic> json) {
    return WorkoutLikeUser(
      handle: json['handle'] as String,
      nickname: json['nickname'] as String? ?? '',
      name: json['name'] as String? ?? '',
      isLocal: json['is_local'] as bool? ?? false,
      hasAvatar: json['has_avatar'] as bool? ?? false,
      avatarUrl: json['avatar_url'] as String?,
    );
  }
}

class WorkoutLikeState {
  WorkoutLikeState({required this.count, required this.likedByMe});

  final int count;
  final bool likedByMe;

  factory WorkoutLikeState.fromJson(Map<String, dynamic> json) {
    return WorkoutLikeState(
      count: json['count'] as int? ?? 0,
      likedByMe: json['liked_by_me'] as bool? ?? false,
    );
  }
}

class WorkoutLikesResponse {
  WorkoutLikesResponse({required this.count, required this.users});

  final int count;
  final List<WorkoutLikeUser> users;

  factory WorkoutLikesResponse.fromJson(Map<String, dynamic> json) {
    final rawUsers = json['users'];
    final users = <WorkoutLikeUser>[];
    if (rawUsers is List) {
      for (final item in rawUsers) {
        if (item is Map<String, dynamic>) {
          users.add(WorkoutLikeUser.fromJson(item));
        }
      }
    }
    return WorkoutLikesResponse(
      count: json['count'] as int? ?? users.length,
      users: users,
    );
  }
}

class WorkoutComment {
  WorkoutComment({
    required this.id,
    required this.user,
    required this.datetime,
    required this.text,
    this.canDelete = false,
  });

  final String id;
  final WorkoutLikeUser user;
  final DateTime datetime;
  final String text;
  final bool canDelete;

  factory WorkoutComment.fromJson(Map<String, dynamic> json) {
    final userJson = json['user'];
    return WorkoutComment(
      id: json['id'] as String? ?? '',
      user: userJson is Map<String, dynamic>
          ? WorkoutLikeUser.fromJson(userJson)
          : WorkoutLikeUser(
              handle: '',
              nickname: '',
              name: '',
              isLocal: false,
            ),
      datetime: DateTime.tryParse(json['datetime'] as String? ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0, isUtc: true),
      text: json['text'] as String? ?? '',
      canDelete: json['can_delete'] as bool? ?? false,
    );
  }
}

class WorkoutCommentsResponse {
  WorkoutCommentsResponse({required this.count, required this.comments});

  final int count;
  final List<WorkoutComment> comments;

  factory WorkoutCommentsResponse.fromJson(Map<String, dynamic> json) {
    final raw = json['comments'];
    final comments = <WorkoutComment>[];
    if (raw is List) {
      for (final item in raw) {
        if (item is Map<String, dynamic>) {
          comments.add(WorkoutComment.fromJson(item));
        }
      }
    }
    return WorkoutCommentsResponse(
      count: json['count'] as int? ?? comments.length,
      comments: comments,
    );
  }
}

class WorkoutCommentCreateResponse {
  WorkoutCommentCreateResponse({required this.count, required this.comment});

  final int count;
  final WorkoutComment comment;

  factory WorkoutCommentCreateResponse.fromJson(Map<String, dynamic> json) {
    final commentJson = json['comment'];
    return WorkoutCommentCreateResponse(
      count: json['count'] as int? ?? 0,
      comment: commentJson is Map<String, dynamic>
          ? WorkoutComment.fromJson(commentJson)
          : WorkoutComment(
              id: '',
              user: WorkoutLikeUser(
                handle: '',
                nickname: '',
                name: '',
                isLocal: false,
              ),
              datetime: DateTime.fromMillisecondsSinceEpoch(0, isUtc: true),
              text: '',
            ),
    );
  }
}

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
    this.likesCount = 0,
    this.likedByMe = false,
    this.canLike = false,
    this.commentsCount = 0,
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
  final int likesCount;
  final bool likedByMe;
  final bool canLike;
  final int commentsCount;

  double get distanceKm => distance / 1000;

  String get ownerNickname =>
      owner.isNotEmpty ? owner : (author?.nickname ?? '');

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
      likesCount: json['likes_count'] as int? ?? 0,
      likedByMe: json['liked_by_me'] as bool? ?? false,
      canLike: json['can_like'] as bool? ?? false,
      commentsCount: json['comments_count'] as int? ?? 0,
    );
  }

  Workout copyWith({
    int? likesCount,
    bool? likedByMe,
    bool? canLike,
    int? commentsCount,
  }) {
    return Workout(
      id: id,
      name: name,
      description: description,
      sportType: sportType,
      startDate: startDate,
      durationSeconds: durationSeconds,
      distance: distance,
      durationTotalSeconds: durationTotalSeconds,
      tempAvgKmm: tempAvgKmm,
      speedMaxKmh: speedMaxKmh,
      speedAvgKmh: speedAvgKmh,
      elevationGain: elevationGain,
      heartRateAvg: heartRateAvg,
      heartRateMax: heartRateMax,
      stepsTotal: stepsTotal,
      calories: calories,
      owner: owner,
      device: device,
      track: track,
      hasMapPreview: hasMapPreview,
      hasMedia: hasMedia,
      mediaFiles: mediaFiles,
      author: author,
      equipment: equipment,
      likesCount: likesCount ?? this.likesCount,
      likedByMe: likedByMe ?? this.likedByMe,
      canLike: canLike ?? this.canLike,
      commentsCount: commentsCount ?? this.commentsCount,
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
