import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/social.dart';
import '../models/user_public_profile.dart';
import '../models/workout.dart';
import '../pages/workout_detail_page.dart';
import '../widgets/follow_list_dialog.dart';
import '../widgets/user_profile_view.dart';
import '../widgets/workout_feed_list.dart';

class UserProfilePage extends StatefulWidget {
  const UserProfilePage({
    super.key,
    required this.handle,
    this.viewerNickname,
    this.federationEnabled = false,
    this.api,
    this.viewingWorkout,
    this.isMapExpanded = false,
    this.onViewingWorkoutChanged,
    this.onMapExpandedChanged,
    this.photoViewerIndex,
    this.onPhotoViewerIndexChanged,
  });

  final String handle;
  final String? viewerNickname;
  final bool federationEnabled;
  final ApiRequest? api;

  final Workout? viewingWorkout;
  final bool isMapExpanded;
  final ValueChanged<Workout?>? onViewingWorkoutChanged;
  final ValueChanged<bool>? onMapExpandedChanged;
  final int? photoViewerIndex;
  final ValueChanged<int?>? onPhotoViewerIndexChanged;

  @override
  State<UserProfilePage> createState() => _UserProfilePageState();
}

class _UserProfilePageState extends State<UserProfilePage> {
  late final ApiRequest _api = widget.api ?? ApiRequest();
  final _scrollController = ScrollController();
  final _feedKey = GlobalKey<WorkoutFeedListState>();

  UserPublicProfile? _profile;
  List<FollowInfo> _following = [];
  List<FollowerInfo> _followers = [];
  String? _authToken;
  bool _isLoading = true;
  String? _error;
  bool _followBusy = false;

  ViewerFollow? get _viewerFollow => _profile?.viewerFollow;

  bool get _isFollowing =>
      _viewerFollow != null && _viewerFollow!.isActive;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void didUpdateWidget(covariant UserProfilePage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.handle != oldWidget.handle) {
      _profile = null;
      _following = [];
      _followers = [];
      _load();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
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

      final results = await Future.wait([
        _api.getUserProfile(token: token, handle: widget.handle),
        _api.listUserFollowing(token: token, handle: widget.handle),
        _api.listUserFollowers(token: token, handle: widget.handle),
      ]);
      if (!mounted) return;
      setState(() {
        _authToken = token;
        _profile = results[0] as UserPublicProfile;
        _following = (results[1] as List<FollowInfo>)
            .where((f) => f.status == 'active' || f.status == 'pending')
            .toList();
        _followers = results[2] as List<FollowerInfo>;
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
        _error = l10n.failedToLoadUserProfile;
        _isLoading = false;
      });
    }
  }

  Future<void> _toggleFollow() async {
    final token = _authToken;
    final profile = _profile;
    if (token == null || profile == null || _followBusy) {
      return;
    }

    setState(() => _followBusy = true);
    try {
      if (_isFollowing && _viewerFollow != null && _viewerFollow!.id.isNotEmpty) {
        await _api.unfollowUser(token: token, followId: _viewerFollow!.id);
        if (!mounted) return;
        setState(() {
          _profile = UserPublicProfile(
            nickname: profile.nickname,
            name: profile.name,
            handle: profile.handle,
            isLocal: profile.isLocal,
            hasAvatar: profile.hasAvatar,
            avatarUrl: profile.avatarUrl,
            viewerFollow: null,
          );
          _followBusy = false;
        });
      } else {
        final follow = await _api.followUser(token: token, handle: profile.handle);
        if (!mounted) return;
        setState(() {
          _profile = UserPublicProfile(
            nickname: profile.nickname,
            name: profile.name,
            handle: profile.handle,
            isLocal: profile.isLocal,
            hasAvatar: profile.hasAvatar,
            avatarUrl: profile.avatarUrl,
            viewerFollow: ViewerFollow(id: follow.id, status: follow.status),
          );
          _followBusy = false;
        });
      }
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() => _followBusy = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) return;
      setState(() => _followBusy = false);
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToLoadUserProfile)),
      );
    }
  }

  Future<void> _openFollowers() {
    return showFollowersDialog(
      context,
      followers: _followers,
      authToken: _authToken,
      selfNickname: widget.viewerNickname,
      federationEnabled: widget.federationEnabled,
    );
  }

  Future<void> _openFollowing() {
    return showFollowingDialog(
      context,
      following: _following,
      authToken: _authToken,
      selfNickname: widget.viewerNickname,
      federationEnabled: widget.federationEnabled,
    );
  }

  void _openWorkout(Workout workout) {
    widget.onViewingWorkoutChanged?.call(workout);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final viewingWorkout = widget.viewingWorkout;
    final showWorkoutDetail = viewingWorkout != null && _authToken != null;

    return Stack(
      children: [
        Positioned.fill(
          child: Offstage(
            offstage: showWorkoutDetail,
            child: TickerMode(
              enabled: !showWorkoutDetail,
              child: _buildBody(l10n),
            ),
          ),
        ),
        if (showWorkoutDetail)
          Positioned.fill(
            child: WorkoutDetailView(
              workout: viewingWorkout,
              authToken: _authToken!,
              federationEnabled: widget.federationEnabled,
              selfNickname: widget.viewerNickname,
              isMapExpanded: widget.isMapExpanded,
              onMapExpandedChanged: widget.onMapExpandedChanged,
              photoViewerIndex: widget.photoViewerIndex,
              onPhotoViewerIndexChanged: widget.onPhotoViewerIndexChanged,
            ),
          ),
      ],
    );
  }

  Widget _buildBody(AppLocalizations l10n) {
    if (_isLoading && _profile == null) {
      return const Center(child: CircularProgressIndicator());
    }

    final profile = _profile;
    if (_error != null && profile == null) {
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

    if (profile == null) {
      return const SizedBox.shrink();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              ProfileIdentityCard(
                nickname: profile.nickname,
                name: profile.name,
                hasAvatar: profile.hasAvatar,
                avatarUrl: profile.avatarUrl,
                authToken: _authToken,
                trailing: _followBusy
                    ? const SizedBox(
                        width: 24,
                        height: 24,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : IconButton.filledTonal(
                        onPressed: _toggleFollow,
                        icon: Icon(
                          _isFollowing ? Icons.person_remove : Icons.person_add,
                        ),
                        tooltip: _isFollowing ? l10n.unfollow : l10n.follow,
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
        ),
        const SizedBox(height: 8),
        Expanded(
          child: WorkoutFeedList(
            key: _feedKey,
            nickname: widget.viewerNickname ?? '',
            selfNickname: widget.viewerNickname,
            userHandle: widget.handle,
            scope: 'feed',
            scrollController: _scrollController,
            refreshToken: 0,
            federationEnabled: widget.federationEnabled,
            onWorkoutTap: _openWorkout,
            onPhotoTap: (workout, photoIndex) {
              _openWorkout(workout);
            },
            onAuthTokenLoaded: (token) {
              if (_authToken != token) {
                setState(() => _authToken = token);
              }
            },
            emptyMessage: l10n.noWorkoutsFromUser,
          ),
        ),
      ],
    );
  }
}
