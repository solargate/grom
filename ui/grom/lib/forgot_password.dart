import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import 'api_request.dart';
import 'platform/is_mobile_client.dart';
import 'server_storage.dart';
import 'server_url_resolver.dart';
import 'widgets/altcha_field.dart';
import 'widgets/server_url_field.dart';

class ForgotPasswordPage extends StatelessWidget {
  const ForgotPasswordPage({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.forgotPasswordTitle)),
      body: const SingleChildScrollView(
        padding: EdgeInsets.all(24),
        child: ForgotPasswordForm(),
      ),
    );
  }
}

class ForgotPasswordForm extends StatefulWidget {
  const ForgotPasswordForm({super.key});

  @override
  State<ForgotPasswordForm> createState() => _ForgotPasswordFormState();
}

class _ForgotPasswordFormState extends State<ForgotPasswordForm> {
  final _formKey = GlobalKey<FormState>();
  final _api = ApiRequest();
  final _emailController = TextEditingController();
  final _serverUrlController = TextEditingController();
  bool _isSubmitting = false;
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
    _emailController.dispose();
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

      await _api.forgotPassword(
        email: _emailController.text.trim(),
        altcha: _altchaPayload,
      );

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.forgotPasswordCheckEmail)),
      );
      Navigator.pop(context);
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
        SnackBar(content: Text(l10n.forgotPasswordFailed)),
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
          Text(l10n.forgotPasswordHint),
          const SizedBox(height: 16),
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
            textInputAction: TextInputAction.done,
            onFieldSubmitted: (_) => _submit(),
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
                : Text(l10n.forgotPasswordSubmit),
          ),
        ],
      ),
    );
  }
}
