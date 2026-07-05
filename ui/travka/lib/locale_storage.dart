import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

const localeStorageKey = 'app_locale';

const supportedLocaleCodes = ['en', 'ru', 'de'];

class LocaleStorage {
  static Future<Locale?> getLocale() async {
    final prefs = await SharedPreferences.getInstance();
    final code = prefs.getString(localeStorageKey);
    if (code == null || !supportedLocaleCodes.contains(code)) {
      return null;
    }
    return Locale(code);
  }

  static Future<void> saveLocale(Locale locale) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(localeStorageKey, locale.languageCode);
  }

  static Locale resolveLocale(Locale? deviceLocale) {
    if (deviceLocale != null &&
        supportedLocaleCodes.contains(deviceLocale.languageCode)) {
      return Locale(deviceLocale.languageCode);
    }
    return const Locale('en');
  }
}
