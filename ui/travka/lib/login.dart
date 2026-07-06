import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import 'api_request.dart';
import 'auth_storage.dart';
import 'platform/is_mobile_client.dart';
import 'server_storage.dart';
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

  @override
  void initState() {
    super.initState();
    if (isMobileClient) {
      _loadSavedServerUrl();
    }
  }

  Future<void> _loadSavedServerUrl() async {
    final url = await ServerStorage.getBaseUrl();
    if (url != null && mounted) {
      _serverUrlController.text = url;
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

    try {
      if (isMobileClient) {
        await ServerStorage.saveBaseUrl(_serverUrlController.text);
      }

      final result = await _api.login(
        email: _emailController.text.trim(),
        password: _passwordController.text,
      );

      await AuthStorage.saveToken(result.token);

      if (!mounted) return;

      widget.onLoggedIn?.call(result.user);
      if (widget.popOnSuccess) {
        Navigator.pop(context, result.user);
      }
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
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
          const SizedBox(height: 24),
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
