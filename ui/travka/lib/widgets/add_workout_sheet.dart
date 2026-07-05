import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travka/l10n/app_localizations.dart';
import 'package:travka/l10n/sport_type_localizations.dart';
import 'package:travka/models/sport_types.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/workout.dart';

Future<bool?> showAddWorkoutSheet(BuildContext context) {
  final width = MediaQuery.sizeOf(context).width;
  if (width >= 600) {
    return showDialog<bool>(
      context: context,
      builder: (context) => Dialog(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520, maxHeight: 720),
          child: const AddWorkoutSheet(),
        ),
      ),
    );
  }

  return showModalBottomSheet<bool>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (context) => const AddWorkoutSheet(),
  );
}

class AddWorkoutSheet extends StatefulWidget {
  const AddWorkoutSheet({super.key});

  @override
  State<AddWorkoutSheet> createState() => _AddWorkoutSheetState();
}

class _AddWorkoutSheetState extends State<AddWorkoutSheet> {
  final _formKey = GlobalKey<FormState>();
  final _api = ApiRequest();
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();

  String _sportTypeId = defaultSportTypeId;
  late DateTime _date;
  late TimeOfDay _time;
  int _durationSeconds = 0;
  double _distanceKm = 0;
  bool _isSubmitting = false;

  @override
  void initState() {
    super.initState();
    final now = DateTime.now();
    _date = DateTime(now.year, now.month, now.day);
    _time = TimeOfDay.fromDateTime(now);
  }

  @override
  void dispose() {
    _nameController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  DateTime get _startDateTime {
    return DateTime(
      _date.year,
      _date.month,
      _date.day,
      _time.hour,
      _time.minute,
    );
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _date,
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );
    if (picked != null) {
      setState(() => _date = picked);
    }
  }

  Future<void> _pickTime() async {
    final picked = await showTimePicker(
      context: context,
      initialTime: _time,
    );
    if (picked != null) {
      setState(() => _time = picked);
    }
  }

  Future<void> _pickDuration() async {
    final l10n = AppLocalizations.of(context)!;
    final hoursController = TextEditingController(
      text: (_durationSeconds ~/ 3600).toString(),
    );
    final minutesController = TextEditingController(
      text: ((_durationSeconds % 3600) ~/ 60).toString(),
    );
    final secondsController = TextEditingController(
      text: (_durationSeconds % 60).toString(),
    );

    final result = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text(l10n.selectDuration),
          content: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: hoursController,
                  keyboardType: TextInputType.number,
                  decoration: InputDecoration(
                    labelText: l10n.hoursLabel,
                    border: const OutlineInputBorder(),
                  ),
                  autofocus: true,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: TextField(
                  controller: minutesController,
                  keyboardType: TextInputType.number,
                  decoration: InputDecoration(
                    labelText: l10n.minutesLabel,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: TextField(
                  controller: secondsController,
                  keyboardType: TextInputType.number,
                  decoration: InputDecoration(
                    labelText: l10n.secondsLabel,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: Text(l10n.cancel),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: Text(l10n.save),
            ),
          ],
        );
      },
    );

    if (result == true) {
      final hours = _clampDurationPart(hoursController.text, max: 99);
      final minutes = _clampDurationPart(minutesController.text, max: 59);
      final seconds = _clampDurationPart(secondsController.text, max: 59);
      setState(() {
        _durationSeconds = hours * 3600 + minutes * 60 + seconds;
      });
    }

    hoursController.dispose();
    minutesController.dispose();
    secondsController.dispose();
  }

  int _clampDurationPart(String value, {required int max}) {
    final parsed = int.tryParse(value.trim()) ?? 0;
    if (parsed < 0) {
      return 0;
    }
    if (parsed > max) {
      return max;
    }
    return parsed;
  }

  Future<void> _pickDistance() async {
    final l10n = AppLocalizations.of(context)!;
    final controller = TextEditingController(
      text: _distanceKm == 0 ? '0' : _distanceKm.toString(),
    );

    final result = await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text(l10n.selectDistance),
          content: TextField(
            controller: controller,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              labelText: l10n.kilometersLabel,
              border: const OutlineInputBorder(),
              suffixText: 'km',
            ),
            autofocus: true,
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: Text(l10n.cancel),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: Text(l10n.save),
            ),
          ],
        );
      },
    );

    if (result == true) {
      final parsed = double.tryParse(controller.text.replaceAll(',', '.')) ?? 0;
      setState(() => _distanceKm = parsed < 0 ? 0 : parsed);
    }
    controller.dispose();
  }

  Future<void> _pickSportType() async {
    final l10n = AppLocalizations.of(context)!;
    final selected = await showModalBottomSheet<String>(
      context: context,
      isScrollControlled: true,
      builder: (context) {
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
                  l10n.selectWorkoutType,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                Expanded(
                  child: ListView(
                    controller: scrollController,
                    children: [
                      for (final category in SportCategory.values)
                        ..._buildCategorySection(l10n, category),
                    ],
                  ),
                ),
              ],
            );
          },
        );
      },
    );

    if (selected != null) {
      setState(() => _sportTypeId = selected);
    }
  }

  List<Widget> _buildCategorySection(
    AppLocalizations l10n,
    SportCategory category,
  ) {
    final items = sportTypeCatalog
        .where((sportType) => sportType.category == category)
        .toList();

    return [
      Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
        child: Text(
          sportCategoryLabel(l10n, category),
          style: Theme.of(context).textTheme.titleSmall?.copyWith(
                color: Theme.of(context).colorScheme.primary,
                fontWeight: FontWeight.w600,
              ),
        ),
      ),
      for (final sportType in items)
        ListTile(
          leading: CircleAvatar(
            backgroundColor: sportTypeColor(sportType.id).withValues(alpha: 0.15),
            child: Icon(
              sportType.icon,
              color: sportTypeColor(sportType.id),
              size: 20,
            ),
          ),
          title: Text(sportTypeLabel(l10n, sportType.id)),
          selected: _sportTypeId == sportType.id,
          onTap: () => Navigator.pop(context, sportType.id),
        ),
    ];
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    final token = await AuthStorage.getToken();
    if (token == null) {
      return;
    }

    setState(() => _isSubmitting = true);

    try {
      final draft = CreateWorkoutDraft(
        name: _nameController.text.trim(),
        description: _descriptionController.text.trim(),
        sportType: _sportTypeId,
        startDate: _startDateTime,
        durationSeconds: _durationSeconds,
        distanceKm: _distanceKm,
      );

      await _api.createWorkout(token: token, body: draft.toJson());

      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.workoutSaved)),
      );
      Navigator.pop(context, true);
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToSaveWorkout)),
      );
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = l10n.localeName;
    final dateText = DateFormat.yMMMd(locale).format(_date);
    final timeText = _time.format(context);
    final durationText = formatDuration(l10n, _durationSeconds);
    final distanceText = _distanceKm == 0
        ? l10n.distanceZero
        : l10n.distanceKilometers(
            _distanceKm >= 10
                ? _distanceKm.toStringAsFixed(1)
                : _distanceKm.toStringAsFixed(2),
          );
    final selectedSport = sportTypeById(_sportTypeId);

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
                l10n.addWorkout,
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 20),
              TextFormField(
                controller: _nameController,
                decoration: InputDecoration(
                  labelText: l10n.workoutName,
                  border: const OutlineInputBorder(),
                ),
                textInputAction: TextInputAction.next,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return l10n.enterWorkoutName;
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _descriptionController,
                decoration: InputDecoration(
                  labelText: l10n.workoutDescription,
                  border: const OutlineInputBorder(),
                ),
                minLines: 3,
                maxLines: 5,
              ),
              const SizedBox(height: 16),
              InkWell(
                onTap: _pickSportType,
                borderRadius: BorderRadius.circular(4),
                child: InputDecorator(
                  decoration: InputDecoration(
                    labelText: l10n.workoutType,
                    border: const OutlineInputBorder(),
                  ),
                  child: Row(
                    children: [
                      if (selectedSport != null)
                        CircleAvatar(
                          radius: 14,
                          backgroundColor:
                              sportTypeColor(selectedSport.id).withValues(alpha: 0.15),
                          child: Icon(
                            selectedSport.icon,
                            size: 16,
                            color: sportTypeColor(selectedSport.id),
                          ),
                        ),
                      if (selectedSport != null) const SizedBox(width: 12),
                      Expanded(
                        child: Text(sportTypeLabel(l10n, _sportTypeId)),
                      ),
                      const Icon(Icons.arrow_drop_down),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              InkWell(
                onTap: _pickDate,
                child: InputDecorator(
                  decoration: InputDecoration(
                    labelText: l10n.workoutDate,
                    border: const OutlineInputBorder(),
                  ),
                  child: Text(dateText),
                ),
              ),
              const SizedBox(height: 16),
              InkWell(
                onTap: _pickTime,
                child: InputDecorator(
                  decoration: InputDecoration(
                    labelText: l10n.workoutStartTime,
                    border: const OutlineInputBorder(),
                  ),
                  child: Text(timeText),
                ),
              ),
              const SizedBox(height: 16),
              InkWell(
                onTap: _pickDuration,
                child: InputDecorator(
                  decoration: InputDecoration(
                    labelText: l10n.workoutDuration,
                    border: const OutlineInputBorder(),
                  ),
                  child: Text(durationText),
                ),
              ),
              const SizedBox(height: 16),
              InkWell(
                onTap: _pickDistance,
                child: InputDecorator(
                  decoration: InputDecoration(
                    labelText: l10n.workoutDistance,
                    border: const OutlineInputBorder(),
                  ),
                  child: Text(distanceText),
                ),
              ),
              const SizedBox(height: 24),
              Row(
                children: [
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
                      onPressed: _isSubmitting ? null : _submit,
                      child: _isSubmitting
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child: CircularProgressIndicator(strokeWidth: 2),
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
