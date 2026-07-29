import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/equipment.dart';

void main() {
  test('Equipment.fromJson reads distance and optional fields', () {
    final item = Equipment.fromJson({
      'id': 'eq-1',
      'type': 'bike',
      'name': 'Gravel',
      'bike_type': 'gravel',
      'brand': 'Canyon',
      'distance': 12500.5,
    });
    expect(item.id, 'eq-1');
    expect(item.bikeType, 'gravel');
    expect(item.distance, 12500.5);
  });

  test('Equipment.fromJson defaults distance to 0', () {
    final item = Equipment.fromJson({
      'id': 'eq-2',
      'type': 'shoes',
      'name': 'Road',
    });
    expect(item.distance, 0);
  });
}
