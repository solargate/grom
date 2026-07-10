import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../services/avatar_cache.dart';
import '../services/avatar_image_service.dart';
import '../widgets/user_avatar.dart';

class ProfileFormResult {
  const ProfileFormResult({
    required this.saved,
    required this.avatarChanged,
    this.hasAvatar = false,
    this.avatarUrl,
  });

  final bool saved;
  final bool avatarChanged;
  final bool hasAvatar;
  final String? avatarUrl;
}

Future<ProfileFormResult?> showProfileFormDialog(
  BuildContext context, {
  required String initialName,
  required String nickname,
  bool initialHasAvatar = false,
  String? initialAvatarUrl,
  String? initialAuthToken,
}) {
  final width = MediaQuery.sizeOf(context).width;
  if (width >= 600) {
    return showDialog<ProfileFormResult>(
      context: context,
      builder: (dialogContext) => Dialog(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520),
          child: ProfileFormDialog(
            initialName: initialName,
            nickname: nickname,
            initialHasAvatar: initialHasAvatar,
            initialAvatarUrl: initialAvatarUrl,
            initialAuthToken: initialAuthToken,
            navigatorContext: context,
          ),
        ),
      ),
    );
  }

  return showModalBottomSheet<ProfileFormResult>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (sheetContext) => ProfileFormDialog(
      initialName: initialName,
      nickname: nickname,
      initialHasAvatar: initialHasAvatar,
      initialAvatarUrl: initialAvatarUrl,
      initialAuthToken: initialAuthToken,
      navigatorContext: context,
    ),
  );
}

class ProfileFormDialog extends StatefulWidget {
  const ProfileFormDialog({
    super.key,
    required this.initialName,
    required this.nickname,
    required this.navigatorContext,
    this.initialHasAvatar = false,
    this.initialAvatarUrl,
    this.initialAuthToken,
  });

  final String initialName;
  final String nickname;
  final BuildContext navigatorContext;
  final bool initialHasAvatar;
  final String? initialAvatarUrl;
  final String? initialAuthToken;

  @override
  State<ProfileFormDialog> createState() => _ProfileFormDialogState();
}

class _ProfileFormDialogState extends State<ProfileFormDialog> {
  final _formKey = GlobalKey<FormState>();
  final _api = ApiRequest();
  late final TextEditingController _nameController;
  bool _isSubmitting = false;
  bool _isUploadingAvatar = false;
  bool _hasAvatar = false;
  String? _avatarUrl;
  String? _authToken;
  bool _avatarChanged = false;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.initialName);
    _hasAvatar = widget.initialHasAvatar;
    _avatarUrl = widget.initialAvatarUrl;
    _authToken = widget.initialAuthToken;
    if (_authToken == null) {
      _loadToken();
    }
  }

  Future<void> _loadToken() async {
    final token = await AuthStorage.getToken();
    if (!mounted || token == null) {
      return;
    }
    setState(() => _authToken = token);
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  void _close({required bool saved}) {
    Navigator.pop(
      context,
      ProfileFormResult(
        saved: saved,
        avatarChanged: _avatarChanged,
        hasAvatar: _hasAvatar,
        avatarUrl: _avatarUrl,
      ),
    );
  }

  Future<void> _changeAvatar() async {
    if (_isUploadingAvatar || _isSubmitting) {
      return;
    }

    final l10n = AppLocalizations.of(context)!;

    try {
      final bytes = await AvatarImageService.pickCropAndEncode(
        context,
        navigatorContext: widget.navigatorContext,
      );
      if (bytes == null || !mounted) {
        return;
      }

      final token = _authToken ?? await AuthStorage.getToken();
      if (token == null || !mounted) {
        return;
      }
      _authToken = token;

      setState(() => _isUploadingAvatar = true);

      final user = await _api.uploadAvatar(token: token, bytes: bytes);
      if (!mounted) return;
      AvatarCache.instance.bump(widget.nickname);
      setState(() {
        _hasAvatar = user.hasAvatar;
        _avatarUrl = user.avatarUrl;
        _avatarChanged = true;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.avatarUpdated)),
      );
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (e) {
      if (!mounted) return;
      debugPrint('Avatar upload failed: $e');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToUploadAvatar)),
      );
    } finally {
      if (mounted) {
        setState(() => _isUploadingAvatar = false);
      }
    }
  }

  Future<void> _save() async {
    final token = _authToken ?? await AuthStorage.getToken();
    if (token == null || !mounted) {
      return;
    }

    setState(() => _isSubmitting = true);
    final l10n = AppLocalizations.of(context)!;

    try {
      await _api.updateMe(
        token: token,
        name: _nameController.text.trim(),
      );

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.profileSaved)),
      );
      _close(saved: true);
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToSaveProfile)),
      );
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isBusy = _isSubmitting || _isUploadingAvatar;

    return Padding(
      padding: EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: 24 + MediaQuery.viewInsetsOf(context).bottom,
      ),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              l10n.editProfile,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 20),
            Center(
              child: Stack(
                alignment: Alignment.center,
                children: [
                  UserAvatar(
                    nickname: widget.nickname,
                    hasAvatar: _hasAvatar,
                    avatarUrl: _avatarUrl,
                    authToken: _authToken,
                    radius: 48,
                    onTap: isBusy ? null : _changeAvatar,
                    showEditBadge: true,
                  ),
                  if (_isUploadingAvatar)
                    const Positioned.fill(
                      child: ColoredBox(
                        color: Color(0x66000000),
                        child: Center(
                          child: CircularProgressIndicator(),
                        ),
                      ),
                    ),
                ],
              ),
            ),
            const SizedBox(height: 20),
            TextFormField(
              controller: _nameController,
              decoration: InputDecoration(
                labelText: l10n.nameLabel,
                border: const OutlineInputBorder(),
              ),
              textInputAction: TextInputAction.done,
              onFieldSubmitted: isBusy ? null : (_) => _save(),
            ),
            const SizedBox(height: 24),
            Row(
              children: [
                Expanded(
                  child: OutlinedButton(
                    onPressed: isBusy ? null : () => _close(saved: false),
                    child: Text(l10n.cancel),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: FilledButton(
                    onPressed: isBusy ? null : _save,
                    child: _isSubmitting
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(l10n.save),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
