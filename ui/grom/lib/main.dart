import 'package:altcha_widget/altcha_widget.dart';
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_web_plugins/url_strategy.dart';
import 'package:grom/l10n/app_localizations.dart';

import 'app_theme.dart';
import 'locale_storage.dart';
import 'navigation/grom_shell.dart';
import 'platform/is_mobile_client.dart';
import 'reset_password.dart';
import 'server_storage.dart';
import 'services/track_recording_bootstrap.dart';
import 'widgets/altcha_field.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  usePathUrlStrategy();
  await bootstrapTrackRecording();

  if (isMobileClient) {
    await ServerStorage.load();
  }

  final savedLocale = await LocaleStorage.getLocale();
  final initialLocale = savedLocale ??
      LocaleStorage.resolveLocale(
        WidgetsBinding.instance.platformDispatcher.locale,
      );

  runApp(GromApp(initialLocale: initialLocale));
}

class GromApp extends StatefulWidget {
  const GromApp({super.key, required this.initialLocale});

  final Locale initialLocale;

  @override
  State<GromApp> createState() => _GromAppState();
}

class _GromAppState extends State<GromApp> {
  late Locale _locale;

  @override
  void initState() {
    super.initState();
    _locale = widget.initialLocale;
  }

  Future<void> _setLocale(Locale locale) async {
    await LocaleStorage.saveLocale(locale);
    if (!mounted) return;
    setState(() => _locale = locale);
  }

  Route<dynamic> _onGenerateRoute(RouteSettings settings) {
    final name = settings.name ?? '/';
    final uri = Uri.parse(name);
    if (uri.path == '/reset-password') {
      return MaterialPageRoute<void>(
        settings: settings,
        builder: (_) => ResetPasswordPage(
          token: uri.queryParameters['token'] ?? '',
        ),
      );
    }
    return MaterialPageRoute<void>(
      settings: settings,
      builder: (_) => GromShell(
        locale: _locale,
        onLocaleChanged: _setLocale,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Grom',
      locale: _locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        AltchaLocalizationsDelegate(
          customTranslations: altchaCustomTranslations,
        ),
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: AppLocalizations.supportedLocales,
      localeResolutionCallback: (deviceLocale, supportedLocales) {
        return LocaleStorage.resolveLocale(deviceLocale);
      },
      theme: buildAppTheme(),
      onGenerateRoute: _onGenerateRoute,
      initialRoute: '/',
    );
  }
}
