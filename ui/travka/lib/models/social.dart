class WorkoutAuthor {
  WorkoutAuthor({
    required this.nickname,
    required this.name,
    required this.handle,
    required this.isLocal,
  });

  final String nickname;
  final String name;
  final String handle;
  final bool isLocal;

  factory WorkoutAuthor.fromJson(Map<String, dynamic> json) {
    return WorkoutAuthor(
      nickname: json['nickname'] as String,
      name: json['name'] as String? ?? '',
      handle: json['handle'] as String,
      isLocal: json['is_local'] as bool? ?? true,
    );
  }
}

class FollowInfo {
  FollowInfo({
    required this.id,
    required this.targetHandle,
    required this.targetNickname,
    required this.targetName,
    required this.targetIsLocal,
    required this.status,
  });

  final String id;
  final String targetHandle;
  final String targetNickname;
  final String targetName;
  final bool targetIsLocal;
  final String status;

  factory FollowInfo.fromJson(Map<String, dynamic> json) {
    return FollowInfo(
      id: json['id'] as String,
      targetHandle: json['target_handle'] as String,
      targetNickname: json['target_nickname'] as String,
      targetName: json['target_name'] as String? ?? '',
      targetIsLocal: json['target_is_local'] as bool? ?? true,
      status: json['status'] as String,
    );
  }
}

class FollowerInfo {
  FollowerInfo({
    required this.followerHandle,
    required this.followerNickname,
    required this.followerName,
    required this.followerIsLocal,
  });

  final String followerHandle;
  final String followerNickname;
  final String followerName;
  final bool followerIsLocal;

  factory FollowerInfo.fromJson(Map<String, dynamic> json) {
    return FollowerInfo(
      followerHandle: json['follower_handle'] as String,
      followerNickname: json['follower_nickname'] as String,
      followerName: json['follower_name'] as String? ?? '',
      followerIsLocal: json['follower_is_local'] as bool? ?? true,
    );
  }
}

class UserSearchResult {
  UserSearchResult({
    required this.nickname,
    required this.name,
    required this.handle,
    required this.isLocal,
  });

  final String nickname;
  final String name;
  final String handle;
  final bool isLocal;

  factory UserSearchResult.fromJson(Map<String, dynamic> json) {
    return UserSearchResult(
      nickname: json['nickname'] as String,
      name: json['name'] as String? ?? '',
      handle: json['handle'] as String,
      isLocal: json['is_local'] as bool? ?? true,
    );
  }
}
