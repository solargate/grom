import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/equipment_type_localizations.dart';
import 'package:grom/models/equipment.dart';
import 'package:grom/models/equipment_types.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../widgets/equipment_form_dialog.dart';

class EquipmentPage extends StatefulWidget {
  const EquipmentPage({super.key});

  @override
  State<EquipmentPage> createState() => _EquipmentPageState();
}

class _EquipmentPageState extends State<EquipmentPage> {
  final ApiRequest _api = ApiRequest();

  List<Equipment> _items = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }
      final items = await _api.listEquipment(token);
      if (!mounted) return;
      setState(() {
        _items = items;
        _isLoading = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _error = l10n.failedToLoadEquipment;
        _isLoading = false;
      });
    }
  }

  Future<void> _openForm({Equipment? equipment}) async {
    final saved = await showEquipmentFormDialog(context, equipment: equipment);
    if (saved == true) {
      await _load();
    }
  }

  List<Equipment> _itemsForType(EquipmentType type) {
    final typeId = equipmentTypeId(type);
    return _items.where((item) => item.type == typeId).toList();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_error!, textAlign: TextAlign.center),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _load,
                child: Text(l10n.retry),
              ),
            ],
          ),
        ),
      );
    }

    final hasItems = _items.isNotEmpty;

    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Align(
            alignment: Alignment.centerRight,
            child: FilledButton.icon(
              onPressed: () => _openForm(),
              icon: const Icon(Icons.add),
              label: Text(l10n.addEquipment),
            ),
          ),
          const SizedBox(height: 16),
          if (!hasItems)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 48),
              child: Text(
                l10n.noEquipmentYet,
                style: theme.textTheme.bodyLarge,
                textAlign: TextAlign.center,
              ),
            )
          else
            for (final type in equipmentTypeOrder) ...[
              _EquipmentGroupSection(
                type: type,
                items: _itemsForType(type),
                onItemTap: (item) => _openForm(equipment: item),
              ),
              const SizedBox(height: 16),
            ],
        ],
      ),
    );
  }
}

class _EquipmentGroupSection extends StatelessWidget {
  const _EquipmentGroupSection({
    required this.type,
    required this.items,
    required this.onItemTap,
  });

  final EquipmentType type;
  final List<Equipment> items;
  final ValueChanged<Equipment> onItemTap;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const SizedBox.shrink();
    }

    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final color = equipmentTypeColor(type);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: color.withValues(alpha: 0.5)),
          ),
          child: Row(
            children: [
              Icon(equipmentTypeIcon(type), color: color, size: 20),
              const SizedBox(width: 8),
              Text(
                equipmentTypeLabel(l10n, type),
                style: theme.textTheme.titleSmall?.copyWith(
                  color: color,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 8),
        ...items.map(
          (item) => Card(
            margin: const EdgeInsets.only(bottom: 8),
            child: ListTile(
              title: Text(item.name),
              subtitle: _buildSubtitle(item),
              onTap: () => onItemTap(item),
            ),
          ),
        ),
      ],
    );
  }

  Widget? _buildSubtitle(Equipment item) {
    final parts = <String>[];
    if (item.brand.isNotEmpty) {
      parts.add(item.brand);
    }
    if (item.model.isNotEmpty) {
      parts.add(item.model);
    }
    if (parts.isEmpty) {
      return null;
    }
    return Text(parts.join(' · '));
  }
}
