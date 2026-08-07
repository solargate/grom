import 'package:flutter/material.dart';

import '../api_request.dart';
import '../server_storage.dart';
import '../services/avatar_cache.dart';

class UserAvatar extends StatefulWidget {
  const UserAvatar({
    super.key,
    required this.nickname,
    this.hasAvatar = false,
    this.avatarUrl,
    this.authToken,
    this.radius = 20,
    this.onTap,
    this.showEditBadge = false,
  });

  final String nickname;
  final bool hasAvatar;
  final String? avatarUrl;
  final String? authToken;
  final double radius;
  final VoidCallback? onTap;
  final bool showEditBadge;

  @override
  State<UserAvatar> createState() => _UserAvatarState();
}

class _UserAvatarState extends State<UserAvatar> {
  bool _imageFailed = false;

  @override
  void initState() {
    super.initState();
    AvatarCache.instance.addListener(_onAvatarCacheChanged);
  }

  @override
  void dispose() {
    AvatarCache.instance.removeListener(_onAvatarCacheChanged);
    super.dispose();
  }

  void _onAvatarCacheChanged() {
    if (mounted) {
      setState(() => _imageFailed = false);
    }
  }

  @override
  void didUpdateWidget(covariant UserAvatar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.avatarUrl != widget.avatarUrl ||
        oldWidget.hasAvatar != widget.hasAvatar ||
        oldWidget.nickname != widget.nickname ||
        oldWidget.authToken != widget.authToken) {
      _imageFailed = false;
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final resolvedUrl = ApiRequest.resolveAvatarUrl(
      avatarUrl: widget.avatarUrl,
      hasAvatar: widget.hasAvatar,
      nickname: widget.nickname,
    );
    final cacheVersion = AvatarCache.instance.versionFor(widget.nickname);
    final displayUrl = AvatarCache.withCacheBuster(resolvedUrl, cacheVersion);
    final imageHeaders = _imageHeaders(displayUrl);
    final showPlaceholder = displayUrl.isEmpty || _imageFailed;

    Widget avatar = CircleAvatar(
      key: ValueKey(displayUrl),
      radius: widget.radius,
      backgroundColor: theme.colorScheme.surfaceContainerHighest,
      backgroundImage: !showPlaceholder
          ? NetworkImage(displayUrl, headers: imageHeaders)
          : null,
      onBackgroundImageError: !showPlaceholder
          ? (_, __) {
              if (mounted) {
                setState(() => _imageFailed = true);
              }
            }
          : null,
      child: showPlaceholder
          ? Icon(
              Icons.person,
              size: widget.radius,
              color: theme.colorScheme.onSurfaceVariant,
            )
          : null,
    );

    if (widget.onTap != null) {
      avatar = InkWell(
        onTap: widget.onTap,
        customBorder: const CircleBorder(),
        child: avatar,
      );
    }

    if (!widget.showEditBadge) {
      return avatar;
    }

    return Stack(
      clipBehavior: Clip.none,
      children: [
        avatar,
        Positioned(
          right: -2,
          bottom: -2,
          child: Material(
            color: theme.colorScheme.primaryContainer,
            shape: const CircleBorder(),
            child: Padding(
              padding: const EdgeInsets.all(4),
              child: Icon(
                Icons.edit,
                size: widget.radius * 0.45,
                color: theme.colorScheme.onPrimaryContainer,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Map<String, String>? _imageHeaders(String resolvedUrl) {
    if (widget.authToken == null || resolvedUrl.isEmpty) {
      return null;
    }
    if (resolvedUrl.contains('/api/v1/users/') ||
        resolvedUrl.contains('/api/v1/federation/authors/')) {
      return {'Authorization': 'Bearer ${widget.authToken}'};
    }
    final base = ServerStorage.cachedBaseUrl;
    if (base != null && base.isNotEmpty && resolvedUrl.startsWith(base)) {
      return {'Authorization': 'Bearer ${widget.authToken}'};
    }
    return null;
  }
}
