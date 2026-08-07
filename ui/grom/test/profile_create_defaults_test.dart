import 'package:flutter_test/flutter_test.dart';
import 'package:grom/api_request.dart';
import 'package:grom/models/sport_types.dart';

void main() {
  test('profile last sport and equipment drive create-workout defaults', () {
    final profile = UserProfile.fromJson({
      'last_sport_type': 'Ride',
      'last_equipment_by_sport': {
        'Ride': ['bike-1', 'helmet-1'],
        'Run': ['shoe-1'],
      },
    });

    final sportId = resolveDefaultSportTypeId(profile.lastSportType);
    expect(sportId, 'Ride');
    expect(profile.lastEquipmentBySport[sportId], ['bike-1', 'helmet-1']);

    final unknown = resolveDefaultSportTypeId(null);
    expect(unknown, defaultSportTypeId);
    expect(UserProfile.fromJson({}).lastEquipmentBySport[unknown], isNull);
  });
}
