import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import 'api_request.dart';
import 'auth_storage.dart';
import 'forgot_password.dart';
import 'platform/is_mobile_client.dart';
import 'server_storage.dart';
import 'server_url_resolver.dart';
import 'widgets/altcha_field.dart';
import 'widgets/server_url_field.dart';

class LoginForm extends StatefulWidget {
  const LoginForm({
    super.key,
    this.onLoggedIn,
    this.popOnSuccess = false,
  });

  final void Function(UserInfo user)? onLoggedIn;
  final bool popOnSuccess;

  @override
  State<LoginForm> createState() => _LoginFormState();
}

class _LoginFormState extends State<LoginForm> {
  final _formKey = GlobalKey<FormState>();
  final _api = ApiRequest();

  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _serverUrlController = TextEditingController();

  bool _isSubmitting = false;
  bool _obscurePassword = true;
  bool _passwordResetEnabled = false;
  bool _captchaEnabled = false;
  String? _altchaPayload;

  @override
  void initState() {
    super.initState();
    if (isMobileClient) {
      _loadSavedServerUrl();
    } else {
      _loadServerFlags();
    }
  }

  Future<void> _loadSavedServerUrl() async {
    final url = await ServerStorage.getBaseUrl();
    if (url != null && mounted) {
      _serverUrlController.text = url;
    }
    await _loadServerFlags();
  }

  Future<void> _loadServerFlags() async {
    try {
      final info = await _api.getServerInfo();
      if (mounted) {
        setState(() {
          _passwordResetEnabled = info.passwordResetEnabled;
          _captchaEnabled = info.captchaEnabled;
          if (!_captchaEnabled) {
            _altchaPayload = null;
          }
        });
      }
    } catch (_) {
      // Keep optional features hidden when server-info is unavailable.
    }
  }

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
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
        await _loadServerFlags();
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

      final result = await _api.login(
        email: _emailController.text.trim(),
        password: _passwordController.text,
        altcha: _altchaPayload,
      );

      await AuthStorage.saveToken(result.token);

      if (!mounted) return;

      widget.onLoggedIn?.call(result.user);
      if (widget.popOnSuccess) {
        Navigator.pop(context, result.user);
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
        SnackBar(content: Text(l10n.failedToSignIn)),
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
            textInputAction: TextInputAction.done,
            onFieldSubmitted: (_) => _submit(),
            validator: (value) {
              if (value == null || value.isEmpty) {
                return l10n.enterPassword;
              }
              return null;
            },
          ),
          if (_passwordResetEnabled)
            Align(
              alignment: Alignment.centerRight,
              child: TextButton(
                onPressed: _isSubmitting
                    ? null
                    : () {
                        Navigator.push<void>(
                          context,
                          MaterialPageRoute<void>(
                            builder: (_) => const ForgotPasswordPage(),
                          ),
                        );
                      },
                child: Text(l10n.forgotPasswordLink),
              ),
            )
          else
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
                : Text(l10n.signIn),
          ),
        ],
      ),
    );
  }
}

class LoginPage extends StatelessWidget {
  const LoginPage({super.key, this.onLoggedIn});

  final void Function(UserInfo user)? onLoggedIn;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.signIn),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: LoginForm(
          onLoggedIn: onLoggedIn,
          popOnSuccess: true,
        ),
      ),
    );
  }
}
