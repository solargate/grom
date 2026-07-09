import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';
import 'package:travka/l10n/equipment_type_localizations.dart';
import 'package:travka/models/equipment.dart';
import 'package:travka/models/equipment_types.dart';

import '../api_request.dart';
import '../auth_storage.dart';

Future<bool?> showEquipmentFormDialog(
  BuildContext context, {
  Equipment? equipment,
}) {
  final width = MediaQuery.sizeOf(context).width;
  if (width >= 600) {
    return showDialog<bool>(
      context: context,
      builder: (context) => Dialog(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520, maxHeight: 720),
          child: EquipmentFormDialog(equipment: equipment),
        ),
      ),
    );
  }

  return showModalBottomSheet<bool>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (context) => EquipmentFormDialog(equipment: equipment),
  );
}

class EquipmentFormDialog extends StatefulWidget {
  const EquipmentFormDialog({
    super.key,
    this.equipment,
  });

  final Equipment? equipment;

  @override
  State<EquipmentFormDialog> createState() => _EquipmentFormDialogState();
}

class _EquipmentFormDialogState extends State<EquipmentFormDialog> {
  final _formKey = GlobalKey<FormState>();
  final _api = ApiRequest();

  late EquipmentType _type;
  late final TextEditingController _nameController;
  late final TextEditingController _brandController;
  late final TextEditingController _modelController;
  late final TextEditingController _weightController;
  late final TextEditingController _notesController;
  late String _bikeType;
  late String _waterType;
  bool _isSubmitting = false;

  bool get _isEditing => widget.equipment != null;

  @override
  void initState() {
    super.initState();
    final equipment = widget.equipment;
    _type = equipmentTypeFromId(equipment?.type ?? '') ?? EquipmentType.bike;
    _nameController = TextEditingController(text: equipment?.name ?? '');
    _brandController = TextEditingController(text: equipment?.brand ?? '');
    _modelController = TextEditingController(text: equipment?.model ?? '');
    _weightController = TextEditingController(
      text: equipment?.weightKg?.toString() ?? '',
    );
    _notesController = TextEditingController(text: equipment?.notes ?? '');
    _bikeType = equipment?.bikeType ?? '';
    _waterType = equipment?.waterType ?? '';
  }

  @override
  void dispose() {
    _nameController.dispose();
    _brandController.dispose();
    _modelController.dispose();
    _weightController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  EquipmentDraft _buildDraft() {
    final weightText = _weightController.text.trim().replaceAll(',', '.');
    final weight = weightText.isEmpty ? null : double.tryParse(weightText);
    return EquipmentDraft(
      type: equipmentTypeId(_type),
      name: _nameController.text.trim(),
      bikeType: _type == EquipmentType.bike ? _bikeType : '',
      waterType: _type == EquipmentType.water ? _waterType : '',
      brand: _brandController.text.trim(),
      model: _modelController.text.trim(),
      weightKg: weight,
      notes: _notesController.text.trim(),
    );
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    final token = await AuthStorage.getToken();
    if (token == null || !mounted) {
      return;
    }

    setState(() => _isSubmitting = true);
    final l10n = AppLocalizations.of(context)!;

    try {
      final draft = _buildDraft();
      if (_isEditing) {
        await _api.updateEquipment(
          token: token,
          id: widget.equipment!.id,
          body: draft,
        );
      } else {
        await _api.createEquipment(token: token, body: draft);
      }

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.equipmentSaved)),
      );
      Navigator.pop(context, true);
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToSaveEquipment)),
      );
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  Future<void> _confirmDelete() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.deleteEquipment),
        content: Text(l10n.deleteEquipmentConfirm),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.deleteEquipment),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) {
      return;
    }

    final token = await AuthStorage.getToken();
    if (token == null) {
      return;
    }

    setState(() => _isSubmitting = true);
    try {
      await _api.deleteEquipment(
        token: token,
        id: widget.equipment!.id,
      );
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.equipmentDeleted)),
      );
      Navigator.pop(context, true);
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  Future<void> _pickType() async {
    final l10n = AppLocalizations.of(context)!;
    final selected = await showModalBottomSheet<EquipmentType>(
      context: context,
      builder: (context) {
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const SizedBox(height: 12),
              Text(
                l10n.equipmentType,
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 8),
              for (final type in equipmentTypeOrder)
                ListTile(
                  leading: CircleAvatar(
                    backgroundColor:
                        equipmentTypeColor(type).withValues(alpha: 0.15),
                    child: Icon(
                      equipmentTypeIcon(type),
                      color: equipmentTypeColor(type),
                      size: 20,
                    ),
                  ),
                  title: Text(equipmentTypeLabel(l10n, type)),
                  selected: _type == type,
                  onTap: () => Navigator.pop(context, type),
                ),
            ],
          ),
        );
      },
    );

    if (selected != null) {
      setState(() => _type = selected);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final title = _isEditing ? l10n.equipment : l10n.addEquipment;

    return Padding(
      padding: EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: 24 + MediaQuery.viewInsetsOf(context).bottom,
      ),
      child: Form(
        key: _formKey,
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                title,
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 20),
              InkWell(
                onTap: _isSubmitting ? null : _pickType,
                borderRadius: BorderRadius.circular(4),
                child: InputDecorator(
                  decoration: InputDecoration(
                    labelText: '${l10n.equipmentType} *',
                    border: const OutlineInputBorder(),
                  ),
                  child: Row(
                    children: [
                      CircleAvatar(
                        radius: 14,
                        backgroundColor:
                            equipmentTypeColor(_type).withValues(alpha: 0.15),
                        child: Icon(
                          equipmentTypeIcon(_type),
                          size: 16,
                          color: equipmentTypeColor(_type),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(equipmentTypeLabel(l10n, _type)),
                      ),
                      const Icon(Icons.arrow_drop_down),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _nameController,
                decoration: InputDecoration(
                  labelText: '${l10n.equipmentName} *',
                  border: const OutlineInputBorder(),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return l10n.enterEquipmentName;
                  }
                  return null;
                },
              ),
              if (_type == EquipmentType.bike) ...[
                const SizedBox(height: 16),
                DropdownButtonFormField<String>(
                  key: ValueKey(_bikeType),
                  initialValue: _bikeType,
                  decoration: InputDecoration(
                    labelText: l10n.bikeType,
                    border: const OutlineInputBorder(),
                  ),
                  items: [
                    for (final id in bikeTypeIds)
                      DropdownMenuItem(
                        value: id,
                        child: Text(bikeTypeLabel(l10n, id)),
                      ),
                  ],
                  onChanged: _isSubmitting
                      ? null
                      : (value) => setState(() => _bikeType = value ?? ''),
                ),
              ],
              if (_type == EquipmentType.water) ...[
                const SizedBox(height: 16),
                DropdownButtonFormField<String>(
                  key: ValueKey(_waterType),
                  initialValue: _waterType,
                  decoration: InputDecoration(
                    labelText: l10n.waterEquipmentType,
                    border: const OutlineInputBorder(),
                  ),
                  items: [
                    for (final id in waterTypeIds)
                      DropdownMenuItem(
                        value: id,
                        child: Text(waterTypeLabel(l10n, id)),
                      ),
                  ],
                  onChanged: _isSubmitting
                      ? null
                      : (value) => setState(() => _waterType = value ?? ''),
                ),
              ],
              const SizedBox(height: 16),
              TextFormField(
                controller: _brandController,
                decoration: InputDecoration(
                  labelText: l10n.equipmentBrand,
                  border: const OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _modelController,
                decoration: InputDecoration(
                  labelText: l10n.equipmentModel,
                  border: const OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _weightController,
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                decoration: InputDecoration(
                  labelText: l10n.equipmentWeight,
                  border: const OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _notesController,
                decoration: InputDecoration(
                  labelText: l10n.equipmentNotes,
                  border: const OutlineInputBorder(),
                ),
                minLines: 3,
                maxLines: 5,
              ),
              const SizedBox(height: 24),
              Row(
                children: [
                  if (_isEditing) ...[
                    Expanded(
                      child: OutlinedButton(
                        onPressed: _isSubmitting ? null : _confirmDelete,
                        child: Text(l10n.deleteEquipment),
                      ),
                    ),
                    const SizedBox(width: 12),
                  ],
                  Expanded(
                    child: OutlinedButton(
                      onPressed: _isSubmitting
                          ? null
                          : () => Navigator.pop(context, false),
                      child: Text(l10n.cancel),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: FilledButton(
                      onPressed: _isSubmitting ? null : _save,
                      child: _isSubmitting
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child:
                                  CircularProgressIndicator(strokeWidth: 2),
                            )
                          : Text(l10n.save),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
