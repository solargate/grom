import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/social.dart';
import '../widgets/follow_list_dialog.dart';
import '../widgets/profile_form_dialog.dart';
import '../widgets/user_profile_view.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({
    super.key,
    required this.nickname,
    this.api,
  });

  final String nickname;

  /// Optional API client override (tests).
  final ApiRequest? api;

  @override
  State<ProfilePage> createState() => ProfilePageState();
}

class ProfilePageState extends State<ProfilePage> {
  late final ApiRequest _api = widget.api ?? ApiRequest();

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
        _following = following
            .where((f) => f.status == 'active' || f.status == 'pending')
            .toList();
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

  Future<void> openEditProfile() async {
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

  Future<void> _openFollowers() {
    return showFollowersDialog(
      context,
      followers: _followers,
      authToken: _authToken,
      selfNickname: widget.nickname,
    );
  }

  Future<void> _openFollowing() {
    return showFollowingDialog(
      context,
      following: _following,
      authToken: _authToken,
      selfNickname: widget.nickname,
    );
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
          ProfileIdentityCard(
            nickname: widget.nickname,
            name: _name,
            hasAvatar: _hasAvatar,
            avatarUrl: _avatarUrl,
            authToken: _authToken,
            onTap: openEditProfile,
            trailing: Icon(
              Icons.edit_outlined,
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 12),
          ProfileFollowCountCards(
            followingCount: activeFollowingCount(_following),
            followersCount: _followers.length,
            onFollowingTap: _openFollowing,
            onFollowersTap: _openFollowers,
          ),
        ],
      ),
    );
  }
}
