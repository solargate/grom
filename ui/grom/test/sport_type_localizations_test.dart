import 'package:flutter_test/flutter_test.dart';
import 'package:grom/l10n/app_localizations_en.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/sport_types.dart';

void main() {
  final l10n = AppLocalizationsEn();

  test('every catalog sport type has an English label', () {
    for (final info in sportTypeCatalog) {
      final label = sportTypeLocalization(l10n, info.id);
      expect(label, isNotEmpty, reason: 'missing label for ${info.id}');
    }
  });

  test('sportTypeLabel uses catalog info', () {
    expect(sportTypeLabel(l10n, 'Run'), 'Run');
    expect(sportTypeLabel(l10n, 'NotARealSport'), 'NotARealSport');
  });
}
