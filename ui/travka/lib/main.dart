import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:travka/l10n/app_localizations.dart';

import 'locale_storage.dart';
import 'navigation/travka_shell.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final savedLocale = await LocaleStorage.getLocale();
  final initialLocale = savedLocale ??
      LocaleStorage.resolveLocale(
        WidgetsBinding.instance.platformDispatcher.locale,
      );

  runApp(TravkaApp(initialLocale: initialLocale));
}

class TravkaApp extends StatefulWidget {
  const TravkaApp({super.key, required this.initialLocale});

  final Locale initialLocale;

  @override
  State<TravkaApp> createState() => _TravkaAppState();
}

class _TravkaAppState extends State<TravkaApp> {
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
      title: 'Travka',
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
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
            seedColor: const Color.fromARGB(255, 45, 148, 49)),
        useMaterial3: true,
      ),
      home: TravkaShell(
        locale: _locale,
        onLocaleChanged: _setLocale,
      ),
    );
  }
}
