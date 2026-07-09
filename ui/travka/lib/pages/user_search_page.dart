import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/social.dart';

class UserSearchPage extends StatefulWidget {
  const UserSearchPage({super.key});

  @override
  State<UserSearchPage> createState() => _UserSearchPageState();
}

class _UserSearchPageState extends State<UserSearchPage> {
  final ApiRequest _api = ApiRequest();
  final _queryController = TextEditingController();

  List<UserSearchResult> _results = [];
  Map<String, FollowInfo> _followingByHandle = {};
  bool _isLoading = false;
  String? _error;
  String? _token;

  @override
  void initState() {
    super.initState();
    _loadFollowing();
  }

  @override
  void dispose() {
    _queryController.dispose();
    super.dispose();
  }

  Future<void> _loadFollowing() async {
    final token = await AuthStorage.getToken();
    if (token == null) {
      return;
    }
    _token = token;
    try {
      final following = await _api.listFollowing(token);
      if (!mounted) return;
      setState(() {
        _followingByHandle = {
          for (final item in following) item.targetHandle: item,
        };
      });
    } catch (_) {
      // Ignore: search still works without follow state.
    }
  }

  Future<void> _search() async {
    final query = _queryController.text.trim();
    if (query.isEmpty) {
      setState(() {
        _results = [];
        _error = null;
      });
      return;
    }

    final token = _token ?? await AuthStorage.getToken();
    if (token == null) {
      return;
    }
    _token = token;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final results = await _api.searchUsers(token: token, query: query);
      if (!mounted) return;
      setState(() {
        _results = results;
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
        _error = l10n.failedToSearchUsers;
        _isLoading = false;
      });
    }
  }

  Future<void> _toggleFollow(UserSearchResult user) async {
    final token = _token;
    if (token == null) {
      return;
    }

    final existing = _followingByHandle[user.handle];
    try {
      if (existing != null) {
        await _api.unfollowUser(token: token, followId: existing.id);
        if (!mounted) return;
        setState(() {
          _followingByHandle.remove(user.handle);
        });
      } else {
        final follow = await _api.followUser(token: token, handle: user.handle);
        if (!mounted) return;
        setState(() {
          _followingByHandle[user.handle] = follow;
        });
      }
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            l10n.searchByNicknameOrHandle,
            style: theme.textTheme.bodyMedium?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _queryController,
                  decoration: InputDecoration(
                    hintText: l10n.searchUsersHint,
                    border: const OutlineInputBorder(),
                    isDense: true,
                  ),
                  textInputAction: TextInputAction.search,
                  onSubmitted: (_) => _search(),
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(
                onPressed: _isLoading ? null : _search,
                child: Text(l10n.search),
              ),
            ],
          ),
          const SizedBox(height: 16),
          if (_isLoading)
            const Expanded(
              child: Center(child: CircularProgressIndicator()),
            )
          else if (_error != null)
            Expanded(
              child: Center(child: Text(_error!, textAlign: TextAlign.center)),
            )
          else if (_results.isEmpty)
            Expanded(
              child: Center(
                child: Text(
                  l10n.noUsersFound,
                  style: theme.textTheme.titleMedium,
                  textAlign: TextAlign.center,
                ),
              ),
            )
          else
            Expanded(
              child: ListView.separated(
                itemCount: _results.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, index) {
                  final user = _results[index];
                  final follow = _followingByHandle[user.handle];
                  final isFollowing = follow != null &&
                      (follow.status == 'active' || follow.status == 'pending');

                  return ListTile(
                    title: Text(user.nickname),
                    subtitle: Text(
                      user.name.isNotEmpty ? '${user.name} · ${user.handle}' : user.handle,
                    ),
                    trailing: FilledButton.tonal(
                      onPressed: () => _toggleFollow(user),
                      child: Text(isFollowing ? l10n.unfollow : l10n.follow),
                    ),
                  );
                },
              ),
            ),
        ],
      ),
    );
  }
}
