import 'package:flutter/material.dart';

enum EquipmentType {
  bike,
  shoes,
  water,
  other,
}

const equipmentTypeOrder = [
  EquipmentType.bike,
  EquipmentType.shoes,
  EquipmentType.water,
  EquipmentType.other,
];

String equipmentTypeId(EquipmentType type) {
  switch (type) {
    case EquipmentType.bike:
      return 'bike';
    case EquipmentType.shoes:
      return 'shoes';
    case EquipmentType.water:
      return 'water';
    case EquipmentType.other:
      return 'other';
  }
}

EquipmentType? equipmentTypeFromId(String id) {
  switch (id) {
    case 'bike':
      return EquipmentType.bike;
    case 'shoes':
      return EquipmentType.shoes;
    case 'water':
      return EquipmentType.water;
    case 'other':
      return EquipmentType.other;
    default:
      return null;
  }
}

IconData equipmentTypeIcon(EquipmentType type) {
  switch (type) {
    case EquipmentType.bike:
      return Icons.directions_bike;
    case EquipmentType.shoes:
      return Icons.directions_walk;
    case EquipmentType.water:
      return Icons.kayaking;
    case EquipmentType.other:
      return Icons.category;
  }
}

Color equipmentTypeColor(EquipmentType type) {
  switch (type) {
    case EquipmentType.bike:
      return const Color(0xFF1E88E5);
    case EquipmentType.shoes:
      return const Color(0xFFFC4C02);
    case EquipmentType.water:
      return const Color(0xFF00897B);
    case EquipmentType.other:
      return const Color(0xFF8E24AA);
  }
}

Color equipmentTypeColorById(String id) {
  final type = equipmentTypeFromId(id);
  if (type == null) {
    return Colors.grey;
  }
  return equipmentTypeColor(type);
}

const bikeTypeIds = [
  '',
  'mountain',
  'gravel',
  'road',
  'touring',
  'triathlon',
  'cyclocross',
  'fixie',
  'folding',
  'bmx',
];

const waterTypeIds = [
  '',
  'sup',
  'kayak',
  'canoe',
  'canoe_double',
  'packraft',
  'surf',
];
