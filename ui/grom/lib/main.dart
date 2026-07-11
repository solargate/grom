import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:grom/l10n/app_localizations.dart';

import 'app_theme.dart';
import 'locale_storage.dart';
import 'navigation/grom_shell.dart';
import 'platform/is_mobile_client.dart';
import 'server_storage.dart';
import 'services/track_recording_bootstrap.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
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

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Grom',
      locale: _locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: AppLocalizations.supportedLocales,
      localeResolutionCallback: (deviceLocale, supportedLocales) {
        return LocaleStorage.resolveLocale(deviceLocale);
      },
      theme: buildAppTheme(),
      home: GromShell(
        locale: _locale,
        onLocaleChanged: _setLocale,
      ),
    );
  }
}
