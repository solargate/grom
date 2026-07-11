import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/social.dart';
import '../widgets/profile_form_dialog.dart';
import '../widgets/user_avatar.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({
    super.key,
    required this.nickname,
  });

  final String nickname;

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  final ApiRequest _api = ApiRequest();

  String _name = '';
  bool _hasAvatar = false;
  String? _avatarUrl;
  String? _authToken;
  List<FollowInfo> _following = [];
  List<FollowerInfo> _followers = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }

      final me = await _api.getMe(token);
      final following = await _api.listFollowing(token);
      final followers = await _api.listFollowers(token);
      if (!mounted) return;
      setState(() {
        _authToken = token;
        _name = me.name;
        _hasAvatar = me.hasAvatar;
        _avatarUrl = me.avatarUrl;
        _following = following.where((f) => f.status == 'active' || f.status == 'pending').toList();
        _followers = followers;
        _isLoading = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _error = l10n.failedToLoadProfile;
        _isLoading = false;
      });
    }
  }

  Future<void> _openEditProfile() async {
    final result = await showProfileFormDialog(
      context,
      initialName: _name,
      nickname: widget.nickname,
      initialHasAvatar: _hasAvatar,
      initialAvatarUrl: _avatarUrl,
      initialAuthToken: _authToken,
    );
    if (result == null) {
      return;
    }
    if (result.avatarChanged) {
      setState(() {
        _hasAvatar = result.hasAvatar;
        _avatarUrl = result.avatarUrl;
      });
    }
    if (result.saved) {
      await _load();
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_error!, textAlign: TextAlign.center),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _load,
                child: Text(l10n.retry),
              ),
            ],
          ),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            clipBehavior: Clip.antiAlias,
            child: InkWell(
              onTap: _openEditProfile,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    UserAvatar(
                      nickname: widget.nickname,
                      hasAvatar: _hasAvatar,
                      avatarUrl: _avatarUrl,
                      authToken: _authToken,
                      radius: 28,
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            widget.nickname,
                            style: theme.textTheme.titleLarge,
                          ),
                          if (_name.isNotEmpty) ...[
                            const SizedBox(height: 4),
                            Text(
                              _name,
                              style: theme.textTheme.bodyLarge?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                    Icon(
                      Icons.edit_outlined,
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 24),
          Text(
            l10n.followers,
            style: theme.textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          if (_followers.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 24),
              child: Text(
                l10n.noFollowersYet,
                style: theme.textTheme.bodyLarge,
                textAlign: TextAlign.center,
              ),
            )
          else
            ..._followers.map(
              (follower) => Card(
                margin: const EdgeInsets.only(bottom: 8),
                child: ListTile(
                  leading: UserAvatar(
                    nickname: follower.followerNickname,
                    hasAvatar: follower.followerHasAvatar,
                    avatarUrl: follower.followerAvatarUrl,
                    authToken: _authToken,
                    radius: 20,
                  ),
                  title: Text(follower.followerNickname),
                  subtitle: Text(
                    follower.followerName.isNotEmpty
                        ? '${follower.followerName} · ${follower.followerHandle}'
                        : follower.followerHandle,
                  ),
                ),
              ),
            ),
          const SizedBox(height: 24),
          Text(
            l10n.following,
            style: theme.textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          if (_following.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 24),
              child: Text(
                l10n.noFollowingYet,
                style: theme.textTheme.bodyLarge,
                textAlign: TextAlign.center,
              ),
            )
          else
            ..._following.map(
              (follow) => Card(
                margin: const EdgeInsets.only(bottom: 8),
                child: ListTile(
                  leading: UserAvatar(
                    nickname: follow.targetNickname,
                    hasAvatar: follow.targetHasAvatar,
                    avatarUrl: follow.targetAvatarUrl,
                    authToken: _authToken,
                    radius: 20,
                  ),
                  title: Text(follow.targetNickname),
                  subtitle: Text(
                    follow.targetName.isNotEmpty
                        ? '${follow.targetName} · ${follow.targetHandle}'
                        : follow.targetHandle,
                  ),
                  trailing: follow.status == 'pending'
                      ? Text(
                          l10n.followPending,
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        )
                      : null,
                ),
              ),
            ),
        ],
      ),
    );
  }
}
