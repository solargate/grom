class ViewerFollow {
  ViewerFollow({
    required this.id,
    required this.status,
  });

  final String id;
  final String status;

  bool get isActive => status == 'active' || status == 'pending';

  factory ViewerFollow.fromJson(Map<String, dynamic> json) {
    return ViewerFollow(
      id: json['id'] as String? ?? '',
      status: json['status'] as String? ?? '',
    );
  }
}

class UserPublicProfile {
  UserPublicProfile({
    required this.nickname,
    required this.name,
    required this.handle,
    required this.isLocal,
    this.hasAvatar = false,
    this.avatarUrl,
    this.viewerFollow,
  });

  final String nickname;
  final String name;
  final String handle;
  final bool isLocal;
  final bool hasAvatar;
  final String? avatarUrl;
  final ViewerFollow? viewerFollow;

  factory UserPublicProfile.fromJson(Map<String, dynamic> json) {
    ViewerFollow? viewerFollow;
    final followJson = json['viewer_follow'];
    if (followJson is Map<String, dynamic>) {
      viewerFollow = ViewerFollow.fromJson(followJson);
    }
    return UserPublicProfile(
      nickname: json['nickname'] as String,
      name: json['name'] as String? ?? '',
      handle: json['handle'] as String,
      isLocal: json['is_local'] as bool? ?? true,
      hasAvatar: json['has_avatar'] as bool? ?? false,
      avatarUrl: json['avatar_url'] as String?,
      viewerFollow: viewerFollow,
    );
  }
}
