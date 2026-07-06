import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travka/l10n/app_localizations.dart';
import 'package:travka/l10n/sport_type_localizations.dart';
import 'package:travka/models/sport_types.dart';
import 'package:travka/platform/track_file_picker.dart';

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
  bool _isPickingFile = false;
  String? _trackFilename;
  List<int>? _trackBytes;
  bool _isParsingTrack = false;

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
    final result = await showDialog<int>(
      context: context,
      builder: (context) => _DurationPickerDialog(
        initialSeconds: _durationSeconds,
      ),
    );

    if (result != null) {
      setState(() => _durationSeconds = result);
    }
  }

  Future<void> _pickDistance() async {
    final result = await showDialog<double>(
      context: context,
      builder: (context) => _DistancePickerDialog(
        initialDistanceKm: _distanceKm,
      ),
    );

    if (result != null) {
      setState(() => _distanceKm = result);
    }
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

  Future<void> _pickTrack() async {
    if (_isPickingFile || _isSubmitting) {
      return;
    }
    _isPickingFile = true;

    final l10n = AppLocalizations.of(context)!;

    try {
      final picked = await pickTrackFile();
      if (picked == null) {
        return;
      }

      final filename = picked.filename;
      final bytes = picked.bytes;

      final lower = filename.toLowerCase();
      if (!lower.endsWith('.gpx') && !lower.endsWith('.fit')) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.invalidTrackFormat)),
        );
        return;
      }

      setState(() {
        _trackFilename = filename;
        _trackBytes = bytes;
        _isParsingTrack = true;
      });

      final token = await AuthStorage.getToken();
      if (token == null) {
        return;
      }

      final metadata = await _api.parseTrack(
        token: token,
        trackBytes: bytes,
        trackFilename: filename,
      );

      if (!mounted) return;
      setState(() {
        if (metadata.startDate != null) {
          final local = metadata.startDate!.toLocal();
          _date = DateTime(local.year, local.month, local.day);
          _time = TimeOfDay(hour: local.hour, minute: local.minute);
        }
        if (metadata.durationSeconds != null) {
          _durationSeconds = metadata.durationSeconds!;
        }
        if (metadata.distanceMeters != null) {
          _distanceKm = metadata.distanceMeters! / 1000;
        }
      });

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.trackMetadataApplied)),
      );
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (e) {
      if (!mounted) return;
      debugPrint('Track pick failed: $e');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToParseTrack)),
      );
    } finally {
      _isPickingFile = false;
      if (mounted) {
        setState(() => _isParsingTrack = false);
      }
    }
  }

  void _removeTrack() {
    setState(() {
      _trackFilename = null;
      _trackBytes = null;
    });
  }

  Widget _buildTrackPickerField(AppLocalizations l10n) {
    if (_trackFilename != null) {
      return InputDecorator(
        decoration: InputDecoration(
          labelText: l10n.workoutTrack,
          border: const OutlineInputBorder(),
        ),
        child: Row(
          children: [
            Expanded(
              child: Text(
                l10n.trackFileSelected(_trackFilename!),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            if (_isParsingTrack)
              const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            else
              IconButton(
                tooltip: l10n.removeTrack,
                onPressed: _isSubmitting ? null : _removeTrack,
                icon: const Icon(Icons.close),
              ),
          ],
        ),
      );
    }

    return FilledButton.tonalIcon(
      onPressed: (_isSubmitting || _isPickingFile || _isParsingTrack)
          ? null
          : _pickTrack,
      icon: const Icon(Icons.upload_file),
      label: Text(l10n.selectTrackFile),
    );
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

      if (_trackBytes != null && _trackFilename != null) {
        await _api.createWorkoutMultipart(
          token: token,
          fields: {
            'name': draft.name,
            if (draft.description.isNotEmpty) 'description': draft.description,
            'sport_type': draft.sportType,
            'start_date': draft.startDate.toUtc().toIso8601String(),
            'duration_seconds': draft.durationSeconds.toString(),
            'distance': (draft.distanceKm * 1000).toString(),
          },
          trackBytes: _trackBytes!,
          trackFilename: _trackFilename!,
        );
      } else {
        await _api.createWorkout(token: token, body: draft.toJson());
      }

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
      padding: const EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: 24,
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
              const SizedBox(height: 16),
              _buildTrackPickerField(l10n),
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

class _DurationPickerDialog extends StatefulWidget {
  const _DurationPickerDialog({required this.initialSeconds});

  final int initialSeconds;

  @override
  State<_DurationPickerDialog> createState() => _DurationPickerDialogState();
}

class _DurationPickerDialogState extends State<_DurationPickerDialog> {
  late final TextEditingController _hoursController;
  late final TextEditingController _minutesController;
  late final TextEditingController _secondsController;

  @override
  void initState() {
    super.initState();
    final seconds = widget.initialSeconds;
    _hoursController = TextEditingController(text: '${seconds ~/ 3600}');
    _minutesController =
        TextEditingController(text: '${(seconds % 3600) ~/ 60}');
    _secondsController = TextEditingController(text: '${seconds % 60}');
  }

  @override
  void dispose() {
    _hoursController.dispose();
    _minutesController.dispose();
    _secondsController.dispose();
    super.dispose();
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

  void _save() {
    final hours = _clampDurationPart(_hoursController.text, max: 99);
    final minutes = _clampDurationPart(_minutesController.text, max: 59);
    final seconds = _clampDurationPart(_secondsController.text, max: 59);
    Navigator.pop(context, hours * 3600 + minutes * 60 + seconds);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return AlertDialog(
      title: Text(l10n.selectDuration),
      content: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _hoursController,
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
              controller: _minutesController,
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
              controller: _secondsController,
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
          onPressed: () => Navigator.pop(context),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: _save,
          child: Text(l10n.save),
        ),
      ],
    );
  }
}

class _DistancePickerDialog extends StatefulWidget {
  const _DistancePickerDialog({required this.initialDistanceKm});

  final double initialDistanceKm;

  @override
  State<_DistancePickerDialog> createState() => _DistancePickerDialogState();
}

class _DistancePickerDialogState extends State<_DistancePickerDialog> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(
      text: widget.initialDistanceKm == 0
          ? '0'
          : widget.initialDistanceKm.toString(),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _save() {
    final parsed =
        double.tryParse(_controller.text.replaceAll(',', '.')) ?? 0;
    Navigator.pop(context, parsed < 0 ? 0.0 : parsed);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return AlertDialog(
      title: Text(l10n.selectDistance),
      content: TextField(
        controller: _controller,
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
          onPressed: () => Navigator.pop(context),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: _save,
          child: Text(l10n.save),
        ),
      ],
    );
  }
}
