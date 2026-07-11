import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/equipment_type_localizations.dart';
import 'package:grom/models/equipment.dart';
import 'package:grom/models/equipment_types.dart';

class EquipmentPickerField extends StatelessWidget {
  const EquipmentPickerField({
    super.key,
    required this.equipment,
    required this.selectedIds,
    required this.isSubmitting,
    required this.onPick,
    required this.onRemove,
  });

  final List<Equipment> equipment;
  final List<String> selectedIds;
  final bool isSubmitting;
  final VoidCallback onPick;
  final ValueChanged<String> onRemove;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final selected = selectedIds
        .map((id) => equipment.where((item) => item.id == id).firstOrNull)
        .whereType<Equipment>()
        .toList();

    return InkWell(
      onTap: isSubmitting ? null : onPick,
      borderRadius: BorderRadius.circular(4),
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: l10n.workoutEquipment,
          border: const OutlineInputBorder(),
        ),
        child: selected.isEmpty
            ? Text(
                l10n.selectEquipment,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                    ),
              )
            : Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  for (final item in selected)
                    _EquipmentChip(
                      item: item,
                      enabled: !isSubmitting,
                      onRemove: () => onRemove(item.id),
                    ),
                ],
              ),
      ),
    );
  }
}

class _EquipmentChip extends StatelessWidget {
  const _EquipmentChip({
    required this.item,
    required this.enabled,
    required this.onRemove,
  });

  final Equipment item;
  final bool enabled;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    final type = equipmentTypeFromId(item.type);
    final color = type == null
        ? Colors.grey
        : equipmentTypeColor(type);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(item.name),
          const SizedBox(width: 4),
          InkWell(
            onTap: enabled ? onRemove : null,
            child: Icon(
              Icons.close,
              size: 16,
              color: enabled ? color : color.withValues(alpha: 0.4),
            ),
          ),
        ],
      ),
    );
  }
}

Future<List<String>?> showEquipmentPickerSheet(
  BuildContext context, {
  required List<Equipment> equipment,
  required List<String> selectedIds,
}) {
  return showModalBottomSheet<List<String>>(
    context: context,
    isScrollControlled: true,
    builder: (context) => _EquipmentPickerSheet(
      equipment: equipment,
      initialSelectedIds: selectedIds,
    ),
  );
}

class _EquipmentPickerSheet extends StatefulWidget {
  const _EquipmentPickerSheet({
    required this.equipment,
    required this.initialSelectedIds,
  });

  final List<Equipment> equipment;
  final List<String> initialSelectedIds;

  @override
  State<_EquipmentPickerSheet> createState() => _EquipmentPickerSheetState();
}

class _EquipmentPickerSheetState extends State<_EquipmentPickerSheet> {
  late final Set<String> _selected;

  @override
  void initState() {
    super.initState();
    _selected = widget.initialSelectedIds.toSet();
  }

  List<Equipment> _itemsForType(EquipmentType type) {
    final typeId = equipmentTypeId(type);
    return widget.equipment.where((item) => item.type == typeId).toList();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return DraggableScrollableSheet(
      expand: false,
      initialChildSize: 0.75,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      builder: (context, scrollController) {
        return Column(
          children: [
            const SizedBox(height: 12),
            Text(
              l10n.selectEquipment,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Expanded(
              child: ListView(
                controller: scrollController,
                children: [
                  for (final type in equipmentTypeOrder) ...[
                    if (_itemsForType(type).isNotEmpty) ...[
                      Padding(
                        padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
                        child: Row(
                          children: [
                            Icon(
                              equipmentTypeIcon(type),
                              color: equipmentTypeColor(type),
                              size: 20,
                            ),
                            const SizedBox(width: 8),
                            Text(
                              equipmentTypeLabel(l10n, type),
                              style: Theme.of(context)
                                  .textTheme
                                  .titleSmall
                                  ?.copyWith(
                                    color: equipmentTypeColor(type),
                                    fontWeight: FontWeight.w600,
                                  ),
                            ),
                          ],
                        ),
                      ),
                      for (final item in _itemsForType(type))
                        CheckboxListTile(
                          value: _selected.contains(item.id),
                          title: Text(item.name),
                          subtitle: item.brand.isNotEmpty || item.model.isNotEmpty
                              ? Text(
                                  [item.brand, item.model]
                                      .where((part) => part.isNotEmpty)
                                      .join(' · '),
                                )
                              : null,
                          onChanged: (checked) {
                            setState(() {
                              if (checked == true) {
                                _selected.add(item.id);
                              } else {
                                _selected.remove(item.id);
                              }
                            });
                          },
                        ),
                    ],
                  ],
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: FilledButton(
                onPressed: () => Navigator.pop(context, _selected.toList()),
                child: Text(l10n.save),
              ),
            ),
          ],
        );
      },
    );
  }
}
