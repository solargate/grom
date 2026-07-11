class WorkoutAuthor {
  WorkoutAuthor({
    required this.nickname,
    required this.name,
    required this.handle,
    required this.isLocal,
    this.hasAvatar = false,
    this.avatarUrl,
  });

  final String nickname;
  final String name;
  final String handle;
  final bool isLocal;
  final bool hasAvatar;
  final String? avatarUrl;

  factory WorkoutAuthor.fromJson(Map<String, dynamic> json) {
    return WorkoutAuthor(
      nickname: json['nickname'] as String,
      name: json['name'] as String? ?? '',
      handle: json['handle'] as String,
      isLocal: json['is_local'] as bool? ?? true,
      hasAvatar: json['has_avatar'] as bool? ?? false,
      avatarUrl: json['avatar_url'] as String?,
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
    this.targetHasAvatar = false,
    this.targetAvatarUrl,
  });

  final String id;
  final String targetHandle;
  final String targetNickname;
  final String targetName;
  final bool targetIsLocal;
  final String status;
  final bool targetHasAvatar;
  final String? targetAvatarUrl;

  factory FollowInfo.fromJson(Map<String, dynamic> json) {
    return FollowInfo(
      id: json['id'] as String,
      targetHandle: json['target_handle'] as String,
      targetNickname: json['target_nickname'] as String,
      targetName: json['target_name'] as String? ?? '',
      targetIsLocal: json['target_is_local'] as bool? ?? true,
      status: json['status'] as String,
      targetHasAvatar: json['target_has_avatar'] as bool? ?? false,
      targetAvatarUrl: json['target_avatar_url'] as String?,
    );
  }
}

class FollowerInfo {
  FollowerInfo({
    required this.followerHandle,
    required this.followerNickname,
    required this.followerName,
    required this.followerIsLocal,
    this.followerHasAvatar = false,
    this.followerAvatarUrl,
  });

  final String followerHandle;
  final String followerNickname;
  final String followerName;
  final bool followerIsLocal;
  final bool followerHasAvatar;
  final String? followerAvatarUrl;

  factory FollowerInfo.fromJson(Map<String, dynamic> json) {
    return FollowerInfo(
      followerHandle: json['follower_handle'] as String,
      followerNickname: json['follower_nickname'] as String,
      followerName: json['follower_name'] as String? ?? '',
      followerIsLocal: json['follower_is_local'] as bool? ?? true,
      followerHasAvatar: json['follower_has_avatar'] as bool? ?? false,
      followerAvatarUrl: json['follower_avatar_url'] as String?,
    );
  }
}

class UserSearchResult {
  UserSearchResult({
    required this.nickname,
    required this.name,
    required this.handle,
    required this.isLocal,
    this.hasAvatar = false,
    this.avatarUrl,
  });

  final String nickname;
  final String name;
  final String handle;
  final bool isLocal;
  final bool hasAvatar;
  final String? avatarUrl;

  factory UserSearchResult.fromJson(Map<String, dynamic> json) {
    return UserSearchResult(
      nickname: json['nickname'] as String,
      name: json['name'] as String? ?? '',
      handle: json['handle'] as String,
      isLocal: json['is_local'] as bool? ?? true,
      hasAvatar: json['has_avatar'] as bool? ?? false,
      avatarUrl: json['avatar_url'] as String?,
    );
  }
}
