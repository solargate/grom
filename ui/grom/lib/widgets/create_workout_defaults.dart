/// Resolves default equipment ids for a sport from profile history,
/// keeping only ids that still exist in the user's catalog.
List<String> resolveEquipmentIdsForSport({
  required String sportTypeId,
  required Map<String, List<String>> lastEquipmentBySport,
  required Iterable<String> existingEquipmentIds,
}) {
  final lastIds = lastEquipmentBySport[sportTypeId];
  if (lastIds == null || lastIds.isEmpty) {
    return [];
  }
  final existing = existingEquipmentIds.toSet();
  return lastIds.where(existing.contains).toList();
}
