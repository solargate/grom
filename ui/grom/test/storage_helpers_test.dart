import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/auth_storage.dart';
import 'package:grom/locale_storage.dart';
import 'package:grom/server_storage.dart';
import 'package:grom/session.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ServerStorage.clear();
    await AuthStorage.clear();
  });

  group('AuthStorage', () {
    test('saves, reads, and clears token', () async {
      expect(await AuthStorage.getToken(), isNull);
      await AuthStorage.saveToken('abc');
      expect(await AuthStorage.getToken(), 'abc');
      await AuthStorage.clear();
      expect(await AuthStorage.getToken(), isNull);
    });
  });

  group('ServerStorage', () {
    test('normalizeBaseUrl adds scheme and strips trailing slash', () {
      expect(
        ServerStorage.normalizeBaseUrl('grom.example:8443/'),
        'https://grom.example:8443',
      );
      expect(
        ServerStorage.normalizeBaseUrl('http://localhost:8080/api/'),
        'http://localhost:8080/api',
      );
    });

    test('isValidBaseUrl accepts http(s) hosts only', () {
      expect(ServerStorage.isValidBaseUrl('https://grom.example'), isTrue);
      expect(ServerStorage.isValidBaseUrl('grom.example'), isTrue);
      expect(ServerStorage.isValidBaseUrl(''), isFalse);
      expect(ServerStorage.isValidBaseUrl('ftp://grom.example'), isFalse);
    });

    test('saveBaseUrl persists and clears auth when URL changes', () async {
      await AuthStorage.saveToken('old-token');
      await ServerStorage.saveBaseUrl('https://one.example');
      expect(await ServerStorage.getBaseUrl(), 'https://one.example');
      expect(ServerStorage.cachedBaseUrl, 'https://one.example');
      expect(await AuthStorage.getToken(), 'old-token');

      await ServerStorage.saveBaseUrl('https://two.example');
      expect(await ServerStorage.getBaseUrl(), 'https://two.example');
      expect(await AuthStorage.getToken(), isNull);
    });

    test('load restores cached base URL', () async {
      SharedPreferences.setMockInitialValues({
        serverBaseUrlStorageKey: 'https://cached.example',
      });
      await ServerStorage.load();
      expect(ServerStorage.cachedBaseUrl, 'https://cached.example');
    });
  });

  group('LocaleStorage', () {
    test('save and get locale round-trip', () async {
      expect(await LocaleStorage.getLocale(), isNull);
      await LocaleStorage.saveLocale(const Locale('ru'));
      expect(await LocaleStorage.getLocale(), const Locale('ru'));
    });

    test('resolveLocale prefers supported device locale', () {
      expect(
          LocaleStorage.resolveLocale(const Locale('de')), const Locale('de'));
      expect(
          LocaleStorage.resolveLocale(const Locale('fr')), const Locale('en'));
      expect(LocaleStorage.resolveLocale(null), const Locale('en'));
    });

    test('getLocale ignores unsupported stored codes', () async {
      SharedPreferences.setMockInitialValues({localeStorageKey: 'fr'});
      expect(await LocaleStorage.getLocale(), isNull);
    });
  });

  group('clearLocalSession', () {
    test('clears auth token', () async {
      await AuthStorage.saveToken('abc');
      await clearLocalSession();
      expect(await AuthStorage.getToken(), isNull);
    });
  });
}
