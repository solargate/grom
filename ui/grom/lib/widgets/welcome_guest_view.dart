import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:grom/l10n/app_localizations.dart';

const _logoAsset = 'assets/logo.svg';
const _logoSize = 96.0;
const _logoRadius = 20.0;
const _contentMaxWidth = 420.0;
const _buttonsRowBreakpoint = 360.0;

class WelcomeGuestView extends StatelessWidget {
  const WelcomeGuestView({
    super.key,
    required this.onSignIn,
    required this.onRegister,
    this.showMobileServerHint = false,
  });

  final VoidCallback onSignIn;
  final VoidCallback onRegister;
  final bool showMobileServerHint;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return LayoutBuilder(
      builder: (context, constraints) {
        final buttonsInRow = constraints.maxWidth >= _buttonsRowBreakpoint;
        final minHeight = (constraints.maxHeight - 48).clamp(0.0, double.infinity);

        return SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: ConstrainedBox(
            constraints: BoxConstraints(minHeight: minHeight),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: _contentMaxWidth),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    ClipRRect(
                      borderRadius: BorderRadius.circular(_logoRadius),
                      child: SvgPicture.asset(
                        _logoAsset,
                        width: _logoSize,
                        height: _logoSize,
                        semanticsLabel: l10n.appTitle,
                      ),
                    ),
                    const SizedBox(height: 20),
                    Text(
                      l10n.appTitle,
                      style: theme.textTheme.headlineMedium,
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 12),
                    Text(
                      l10n.welcomeDescription,
                      style: theme.textTheme.bodyLarge?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 24),
                    Text(
                      l10n.welcomeInstructions,
                      style: theme.textTheme.bodyMedium,
                      textAlign: TextAlign.center,
                    ),
                    if (showMobileServerHint) ...[
                      const SizedBox(height: 8),
                      Text(
                        l10n.welcomeMobileServerHint,
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ],
                    const SizedBox(height: 28),
                    _WelcomeActions(
                      inRow: buttonsInRow,
                      onSignIn: onSignIn,
                      onRegister: onRegister,
                      signInLabel: l10n.signIn,
                      registerLabel: l10n.register,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

class _WelcomeActions extends StatelessWidget {
  const _WelcomeActions({
    required this.inRow,
    required this.onSignIn,
    required this.onRegister,
    required this.signInLabel,
    required this.registerLabel,
  });

  final bool inRow;
  final VoidCallback onSignIn;
  final VoidCallback onRegister;
  final String signInLabel;
  final String registerLabel;

  @override
  Widget build(BuildContext context) {
    final signIn = FilledButton(
      onPressed: onSignIn,
      child: Text(signInLabel),
    );
    final register = OutlinedButton(
      onPressed: onRegister,
      child: Text(registerLabel),
    );

    if (inRow) {
      return Row(
        children: [
          Expanded(child: signIn),
          const SizedBox(width: 12),
          Expanded(child: register),
        ],
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        signIn,
        const SizedBox(height: 12),
        register,
      ],
    );
  }
}
