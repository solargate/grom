import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/equipment_types.dart';

String equipmentTypeLabel(AppLocalizations l10n, EquipmentType type) {
  switch (type) {
    case EquipmentType.bike:
      return l10n.equipmentTypeBike;
    case EquipmentType.shoes:
      return l10n.equipmentTypeShoes;
    case EquipmentType.water:
      return l10n.equipmentTypeWater;
    case EquipmentType.other:
      return l10n.equipmentTypeOther;
  }
}

String equipmentTypeLabelById(AppLocalizations l10n, String id) {
  final type = equipmentTypeFromId(id);
  if (type == null) {
    return id;
  }
  return equipmentTypeLabel(l10n, type);
}

String bikeTypeLabel(AppLocalizations l10n, String id) {
  switch (id) {
    case '':
      return l10n.equipmentSubtypeEmpty;
    case 'mountain':
      return l10n.bikeTypeMountain;
    case 'gravel':
      return l10n.bikeTypeGravel;
    case 'road':
      return l10n.bikeTypeRoad;
    case 'touring':
      return l10n.bikeTypeTouring;
    case 'triathlon':
      return l10n.bikeTypeTriathlon;
    case 'cyclocross':
      return l10n.bikeTypeCyclocross;
    case 'fixie':
      return l10n.bikeTypeFixie;
    case 'bmx':
      return l10n.bikeTypeBmx;
    default:
      return id;
  }
}

String waterTypeLabel(AppLocalizations l10n, String id) {
  switch (id) {
    case '':
      return l10n.equipmentSubtypeEmpty;
    case 'sup':
      return l10n.waterTypeSup;
    case 'kayak':
      return l10n.waterTypeKayak;
    case 'canoe':
      return l10n.waterTypeCanoe;
    case 'canoe_double':
      return l10n.waterTypeCanoeDouble;
    case 'packraft':
      return l10n.waterTypePackraft;
    case 'surf':
      return l10n.waterTypeSurf;
    default:
      return id;
  }
}
