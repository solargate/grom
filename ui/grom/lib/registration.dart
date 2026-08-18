import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import 'api_request.dart';
import 'login.dart';
import 'platform/is_mobile_client.dart';
import 'server_history.dart';
import 'server_storage.dart';
import 'server_url_resolver.dart';
import 'widgets/altcha_field.dart';
import 'widgets/server_url_field.dart';

class RegistrationForm extends StatefulWidget {
  const RegistrationForm({
    super.key,
    this.onRegistered,
    this.popOnSuccess = false,
  });

  final VoidCallback? onRegistered;
  final bool popOnSuccess;

  @override
  State<RegistrationForm> createState() => _RegistrationFormState();
}

class _RegistrationFormState extends State<RegistrationForm> {
  final _formKey = GlobalKey<FormState>();
  final _api = ApiRequest();

  final _nicknameController = TextEditingController();
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final _serverUrlController = TextEditingController();

  bool _isSubmitting = false;
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;
  bool _captchaEnabled = false;
  String? _altchaPayload;

  @override
  void initState() {
    super.initState();
    if (isMobileClient) {
      _loadSavedServerUrl();
    } else {
      _loadCaptchaFlag();
    }
  }

  Future<void> _loadSavedServerUrl() async {
    final url = await ServerStorage.getBaseUrl();
    if (url != null && mounted) {
      _serverUrlController.text = url;
    }
    await _loadCaptchaFlag();
  }

  Future<void> _loadCaptchaFlag() async {
    try {
      final info = await _api.getServerInfo();
      if (mounted) {
        setState(() {
          _captchaEnabled = info.captchaEnabled;
          if (!_captchaEnabled) {
            _altchaPayload = null;
          }
        });
      }
    } catch (_) {
      // Keep captcha hidden when server-info is unavailable.
    }
  }

  @override
  void dispose() {
    _nicknameController.dispose();
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    _serverUrlController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() => _isSubmitting = true);
    final l10n = AppLocalizations.of(context)!;

    try {
      if (isMobileClient) {
        final resolved = await resolveServerBaseUrl(_serverUrlController.text);
        if (mounted) {
          _serverUrlController.text = resolved;
        }
        await ServerStorage.saveBaseUrl(resolved);
        await _loadCaptchaFlag();
      }

      if (_captchaEnabled &&
          (_altchaPayload == null || _altchaPayload!.isEmpty)) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(l10n.captchaRequired)),
          );
        }
        return;
      }

      await _api.register(
        nickname: _nicknameController.text.trim(),
        name: _nameController.text.trim(),
        email: _emailController.text.trim(),
        password: _passwordController.text,
        altcha: _altchaPayload,
      );
      if (isMobileClient) {
        await ServerHistory.remember(_serverUrlController.text);
      }

      if (!mounted) return;

      widget.onRegistered?.call();
      if (widget.popOnSuccess) {
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (context) => const LoginPage()),
        );
      }
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() => _altchaPayload = null);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) return;
      setState(() => _altchaPayload = null);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToRegister)),
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

    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (isMobileClient) ...[
            ServerUrlField(controller: _serverUrlController),
            const SizedBox(height: 16),
          ],
          TextFormField(
            controller: _nicknameController,
            decoration: InputDecoration(
              labelText: l10n.nicknameLabel,
              border: const OutlineInputBorder(),
            ),
            textInputAction: TextInputAction.next,
            validator: (value) {
              if (value == null || value.trim().isEmpty) {
                return l10n.enterNickname;
              }
              return null;
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: _nameController,
            decoration: InputDecoration(
              labelText: l10n.nameLabel,
              border: const OutlineInputBorder(),
            ),
            textInputAction: TextInputAction.next,
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: _emailController,
            decoration: InputDecoration(
              labelText: l10n.emailLabel,
              border: const OutlineInputBorder(),
            ),
            keyboardType: TextInputType.emailAddress,
            textInputAction: TextInputAction.next,
            validator: (value) {
              if (value == null || value.trim().isEmpty) {
                return l10n.enterEmail;
              }
              final email = value.trim();
              if (!email.contains('@') || !email.contains('.')) {
                return l10n.enterValidEmail;
              }
              return null;
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: _passwordController,
            decoration: InputDecoration(
              labelText: l10n.passwordLabel,
              border: const OutlineInputBorder(),
              suffixIcon: IconButton(
                icon: Icon(
                  _obscurePassword ? Icons.visibility : Icons.visibility_off,
                ),
                onPressed: () {
                  setState(() => _obscurePassword = !_obscurePassword);
                },
              ),
            ),
            obscureText: _obscurePassword,
            textInputAction: TextInputAction.next,
            validator: (value) {
              if (value == null || value.isEmpty) {
                return l10n.enterPassword;
              }
              if (value.length < 8) {
                return l10n.passwordMinLength;
              }
              return null;
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: _confirmPasswordController,
            decoration: InputDecoration(
              labelText: l10n.confirmPasswordLabel,
              border: const OutlineInputBorder(),
              suffixIcon: IconButton(
                icon: Icon(
                  _obscureConfirmPassword
                      ? Icons.visibility
                      : Icons.visibility_off,
                ),
                onPressed: () {
                  setState(
                    () => _obscureConfirmPassword = !_obscureConfirmPassword,
                  );
                },
              ),
            ),
            obscureText: _obscureConfirmPassword,
            textInputAction: TextInputAction.done,
            onFieldSubmitted: (_) => _submit(),
            validator: (value) {
              if (value == null || value.isEmpty) {
                return l10n.confirmPassword;
              }
              if (value != _passwordController.text) {
                return l10n.passwordsDoNotMatch;
              }
              return null;
            },
          ),
          const SizedBox(height: 24),
          AltchaField(
            enabled: _captchaEnabled,
            challengeUrl: captchaChallengeUrl(_api),
            onPayloadChanged: (payload) {
              setState(() => _altchaPayload = payload);
            },
          ),
          FilledButton(
            onPressed: _isSubmitting ? null : _submit,
            child: _isSubmitting
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Text(l10n.register),
          ),
        ],
      ),
    );
  }
}

class RegistrationPage extends StatelessWidget {
  const RegistrationPage({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.register),
      ),
      body: const SingleChildScrollView(
        padding: EdgeInsets.all(24),
        child: RegistrationForm(popOnSuccess: true),
      ),
    );
  }
}
