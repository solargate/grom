class Equipment {
  Equipment({
    required this.id,
    required this.type,
    required this.name,
    this.bikeType = '',
    this.waterType = '',
    this.brand = '',
    this.model = '',
    this.weightKg,
    this.notes = '',
  });

  final String id;
  final String type;
  final String name;
  final String bikeType;
  final String waterType;
  final String brand;
  final String model;
  final double? weightKg;
  final String notes;

  factory Equipment.fromJson(Map<String, dynamic> json) {
    return Equipment(
      id: json['id'] as String,
      type: json['type'] as String,
      name: json['name'] as String,
      bikeType: json['bike_type'] as String? ?? '',
      waterType: json['water_type'] as String? ?? '',
      brand: json['brand'] as String? ?? '',
      model: json['model'] as String? ?? '',
      weightKg: (json['weight_kg'] as num?)?.toDouble(),
      notes: json['notes'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'name': name,
      if (bikeType.isNotEmpty) 'bike_type': bikeType,
      if (waterType.isNotEmpty) 'water_type': waterType,
      if (brand.isNotEmpty) 'brand': brand,
      if (model.isNotEmpty) 'model': model,
      if (weightKg != null) 'weight_kg': weightKg,
      if (notes.isNotEmpty) 'notes': notes,
    };
  }
}

class EquipmentDraft {
  EquipmentDraft({
    required this.type,
    required this.name,
    this.bikeType = '',
    this.waterType = '',
    this.brand = '',
    this.model = '',
    this.weightKg,
    this.notes = '',
  });

  final String type;
  final String name;
  final String bikeType;
  final String waterType;
  final String brand;
  final String model;
  final double? weightKg;
  final String notes;

  Map<String, dynamic> toJson() => {
        'type': type,
        'name': name,
        if (bikeType.isNotEmpty) 'bike_type': bikeType,
        if (waterType.isNotEmpty) 'water_type': waterType,
        if (brand.isNotEmpty) 'brand': brand,
        if (model.isNotEmpty) 'model': model,
        if (weightKg != null) 'weight_kg': weightKg,
        if (notes.isNotEmpty) 'notes': notes,
      };
}
