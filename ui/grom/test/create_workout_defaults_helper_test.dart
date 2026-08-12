import 'package:flutter_test/flutter_test.dart';
import 'package:grom/widgets/create_workout_defaults.dart';

void main() {
  group('resolveEquipmentIdsForSport', () {
    test('returns empty when profile has no equipment for sport', () {
      expect(
        resolveEquipmentIdsForSport(
          sportTypeId: 'Ride',
          lastEquipmentBySport: const {'Run': ['shoe-1']},
          existingEquipmentIds: const ['shoe-1'],
        ),
        isEmpty,
      );
    });

    test('returns empty when profile lists empty equipment', () {
      expect(
        resolveEquipmentIdsForSport(
          sportTypeId: 'Ride',
          lastEquipmentBySport: const {'Ride': []},
          existingEquipmentIds: const ['bike-1'],
        ),
        isEmpty,
      );
    });

    test('keeps only ids that still exist in catalog', () {
      expect(
        resolveEquipmentIdsForSport(
          sportTypeId: 'Ride',
          lastEquipmentBySport: const {
            'Ride': ['bike-1', 'removed-bike', 'helmet-1'],
          },
          existingEquipmentIds: const ['bike-1', 'helmet-1'],
        ),
        ['bike-1', 'helmet-1'],
      );
    });

    test('returns empty when none of the profile ids exist anymore', () {
      expect(
        resolveEquipmentIdsForSport(
          sportTypeId: 'Run',
          lastEquipmentBySport: const {'Run': ['old-shoe']},
          existingEquipmentIds: const ['new-shoe'],
        ),
        isEmpty,
      );
    });
  });
}
