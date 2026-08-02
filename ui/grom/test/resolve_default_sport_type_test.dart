import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/sport_types.dart';

void main() {
  test('uses last workout sport when known', () {
    expect(resolveDefaultSportTypeId('Ride'), 'Ride');
    expect(resolveDefaultSportTypeId('Swim'), 'Swim');
  });

  test('falls back to Run when missing or unknown', () {
    expect(resolveDefaultSportTypeId(null), defaultSportTypeId);
    expect(resolveDefaultSportTypeId(''), defaultSportTypeId);
    expect(resolveDefaultSportTypeId('NotASport'), defaultSportTypeId);
  });
}
