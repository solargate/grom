import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/platform/is_mobile_client.dart';
import 'package:grom/platform/track_file_picker.dart';
import 'package:grom/platform/workout_photo_picker.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/recorded_track.dart';
import '../models/equipment.dart';
import '../models/workout.dart';
import '../services/track_recording_service.dart';
import '../platform/shared_track_intent.dart';
import 'create_workout_name_sync.dart';
import 'manual_workout_form.dart';
import 'record_workout_tab.dart';
import 'equipment_picker_field.dart';

Future<Workout?> showAddWorkoutSheet(
  BuildContext context, {
  SharedTrackPayload? initialTrack,
  Workout? workout,
}) {
  final width = MediaQuery.sizeOf(context).width;
  if (width >= 600) {
    return showDialog<Workout>(
      context: context,
      builder: (context) => Dialog(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520, maxHeight: 720),
          child: AddWorkoutSheet(
            initialTrack: initialTrack,
            workout: workout,
          ),
        ),
      ),
    );
  }

  return showModalBottomSheet<Workout>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    isDismissible: false,
    enableDrag: false,
    builder: (context) => AddWorkoutSheet(
      initialTrack: initialTrack,
      workout: workout,
    ),
  );
}

class AddWorkoutSheet extends StatefulWidget {
  const AddWorkoutSheet({super.key, this.initialTrack, this.workout});

  final SharedTrackPayload? initialTrack;
  final Workout? workout;

  @override
  State<AddWorkoutSheet> createState() => _AddWorkoutSheetState();
}

class _AddWorkoutSheetState extends State<AddWorkoutSheet>
    with SingleTickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  final _api = ApiRequest();
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _recorder = TrackRecordingService.instance;
  final _nameSync = CreateWorkoutNameSync();

  TabController? _tabController;
  bool _didInitAutoName = false;
  bool _updatingNameProgrammatically = false;

  String _sportTypeId = defaultSportTypeId;
  late DateTime _date;
  late TimeOfDay _time;
  int _startSeconds = 0;
  int _durationSeconds = 0;
  int? _durationTotalSeconds;
  double _distanceKm = 0;
  double? _speedMaxKmh;
  double? _speedAvgKmh;
  bool _isSubmitting = false;
  bool _isPickingFile = false;
  String? _trackFilename;
  List<int>? _trackBytes;
  bool _isParsingTrack = false;
  List<Equipment> _userEquipment = [];
  List<String> _selectedEquipmentIds = [];
  Map<String, List<String>> _lastEquipmentBySport = {};
  List<({String filename, Uint8List bytes})> _selectedPhotos = [];
  bool _isPickingPhotos = false;

  bool get _showRecordTab => isMobileClient && !_isEditing;

  bool get _isEditing => widget.workout != null;

  @override
  void initState() {
    super.initState();
    final existing = widget.workout;
    if (existing != null) {
      _nameSync.synced = false;
      _nameController.text = existing.name;
      _descriptionController.text = existing.description;
      _sportTypeId = existing.sportType;
      _applyStartDateTime(existing.startDate.toLocal());
      _durationSeconds = existing.durationSeconds;
      _durationTotalSeconds = existing.durationTotalSeconds;
      _distanceKm = existing.distanceKm;
      _speedAvgKmh = existing.speedAvgKmh;
      _selectedEquipmentIds =
          existing.equipment.map((item) => item.id).toList();
      if (existing.track.isNotEmpty) {
        _trackFilename = existing.track;
      }
    } else {
      _applyStartDateTime(DateTime.now());
      _nameController.addListener(_onNameEdited);
    }
    if (_showRecordTab) {
      _tabController = TabController(length: 2, vsync: this);
      _recorder.addListener(_onRecorderChanged);
    }
    _loadEquipmentData();

    final initialTrack = widget.initialTrack;
    if (initialTrack != null && !_isEditing) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) {
          return;
        }
        _tabController?.index = 0;
        _applyTrackFile(initialTrack.filename, initialTrack.bytes);
      });
    }
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_isEditing || _didInitAutoName) {
      return;
    }
    _didInitAutoName = true;
    final l10n = AppLocalizations.of(context)!;
    _setNameProgrammatically(sportTypeLabel(l10n, _sportTypeId));
  }

  Future<void> _loadEquipmentData() async {
    final token = await AuthStorage.getToken();
    if (token == null) {
      return;
    }

    final meFuture = _api.getMe(token);
    final equipmentFuture = _api.listEquipment(token);
    final lastOwnFuture =
        _isEditing ? null : _api.listWorkouts(token, scope: 'own', limit: 1);

    String? lastSportType;
    if (lastOwnFuture != null) {
      try {
        final page = await lastOwnFuture;
        if (page.items.isNotEmpty) {
          lastSportType = page.items.first.sportType;
        }
      } catch (_) {
        // Fall back to defaultSportTypeId below.
      }
    }

    try {
      final me = await meFuture;
      final equipment = await equipmentFuture;
      if (!mounted) return;
      setState(() {
        _userEquipment = equipment;
        _lastEquipmentBySport = me.lastEquipmentBySport;
        if (!_isEditing) {
          final resolved = resolveDefaultSportTypeId(lastSportType);
          _sportTypeId = resolved;
          _applyLastEquipmentForSport(resolved);
          final l10n = AppLocalizations.of(context)!;
          final autoName =
              _nameSync.nameForSportChange(sportTypeLabel(l10n, resolved));
          if (autoName != null) {
            _setNameProgrammatically(autoName);
          }
        } else {
          final existingIds = equipment.map((item) => item.id).toSet();
          _selectedEquipmentIds = _selectedEquipmentIds
              .where((id) => existingIds.contains(id))
              .toList();
        }
      });
    } catch (_) {
      // Keep form usable without equipment presets; still apply last sport.
      if (!_isEditing && mounted) {
        final resolved = resolveDefaultSportTypeId(lastSportType);
        if (resolved != _sportTypeId) {
          setState(() {
            _sportTypeId = resolved;
            final l10n = AppLocalizations.of(context)!;
            final autoName =
                _nameSync.nameForSportChange(sportTypeLabel(l10n, resolved));
            if (autoName != null) {
              _setNameProgrammatically(autoName);
            }
          });
        }
      }
    }
  }

  void _applyLastEquipmentForSport(String sportTypeId) {
    final lastIds = _lastEquipmentBySport[sportTypeId];
    if (lastIds == null || lastIds.isEmpty) {
      _selectedEquipmentIds = [];
      return;
    }
    final existingIds = _userEquipment.map((item) => item.id).toSet();
    _selectedEquipmentIds =
        lastIds.where((id) => existingIds.contains(id)).toList();
  }

  @override
  void dispose() {
    if (_showRecordTab) {
      _recorder.removeListener(_onRecorderChanged);
    }
    if (!_isEditing) {
      _nameController.removeListener(_onNameEdited);
    }
    _tabController?.dispose();
    _nameController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  void _onNameEdited() {
    if (_updatingNameProgrammatically || _isEditing) {
      return;
    }
    _nameSync.onUserEdited();
  }

  void _setNameProgrammatically(String value) {
    _updatingNameProgrammatically = true;
    _nameController.value = TextEditingValue(
      text: value,
      selection: TextSelection.collapsed(offset: value.length),
    );
    _updatingNameProgrammatically = false;
  }

  void _onRecorderChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  DateTime get _startDateTime {
    return DateTime(
      _date.year,
      _date.month,
      _date.day,
      _time.hour,
      _time.minute,
      _startSeconds,
    );
  }

  void _applyStartDateTime(DateTime local) {
    _date = DateTime(local.year, local.month, local.day);
    _time = TimeOfDay(hour: local.hour, minute: local.minute);
    _startSeconds = local.second;
  }

  Future<bool> _confirmDiscardRecording() async {
    if (!_recorder.isActive) {
      return true;
    }
    final l10n = AppLocalizations.of(context)!;
    final discard = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.discardRecordingTitle),
        content: Text(l10n.discardRecordingMessage),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.discardRecordingConfirm),
          ),
        ],
      ),
    );
    if (discard == true) {
      await _recorder.discardRecording();
    }
    return discard == true;
  }

  Future<void> _handleCancel() async {
    if (!await _confirmDiscardRecording()) {
      return;
    }
    if (!mounted) return;
    Navigator.pop(context);
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
      setState(() {
        _time = picked;
        _startSeconds = 0;
      });
    }
  }

  Future<void> _pickDuration() async {
    final result = await showDurationPickerDialog(
      context,
      initialSeconds: _durationSeconds,
    );
    if (result != null) {
      setState(() => _durationSeconds = result);
    }
  }

  Future<void> _pickDistance() async {
    final result = await showDistancePickerDialog(
      context,
      initialDistanceKm: _distanceKm,
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
      setState(() {
        _sportTypeId = selected;
        _applyLastEquipmentForSport(selected);
        final autoName = _nameSync.nameForSportChange(sportTypeLabel(l10n, selected));
        if (autoName != null) {
          _setNameProgrammatically(autoName);
        }
      });
    }
  }

  Future<void> _pickEquipment() async {
    final selected = await showEquipmentPickerSheet(
      context,
      equipment: _userEquipment,
      selectedIds: _selectedEquipmentIds,
    );
    if (selected != null) {
      setState(() => _selectedEquipmentIds = selected);
    }
  }

  void _removeEquipment(String id) {
    setState(() {
      _selectedEquipmentIds =
          _selectedEquipmentIds.where((itemId) => itemId != id).toList();
    });
  }

  Future<void> _pickPhotos() async {
    if (_isPickingPhotos || _isSubmitting) {
      return;
    }
    setState(() => _isPickingPhotos = true);
    try {
      final picked = await pickWorkoutPhotos();
      if (!mounted || picked.isEmpty) {
        return;
      }
      setState(() {
        _selectedPhotos = [
          ..._selectedPhotos,
          ...picked.map((item) => (filename: item.filename, bytes: item.bytes)),
        ];
      });
    } catch (e) {
      if (!mounted) return;
      debugPrint('Photo pick failed: $e');
    } finally {
      if (mounted) {
        setState(() => _isPickingPhotos = false);
      }
    }
  }

  void _removePhoto(int index) {
    setState(() {
      _selectedPhotos = [
        for (var i = 0; i < _selectedPhotos.length; i++)
          if (i != index) _selectedPhotos[i],
      ];
    });
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

    try {
      final picked = await pickTrackFile();
      if (picked == null) {
        return;
      }
      await _applyTrackFile(picked.filename, picked.bytes);
    } finally {
      _isPickingFile = false;
    }
  }

  Future<void> _applyTrackFile(String filename, List<int> bytes) async {
    if (_isSubmitting) {
      return;
    }

    final l10n = AppLocalizations.of(context)!;

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

    try {
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
          _applyStartDateTime(metadata.startDate!.toLocal());
        }
        if (metadata.durationSeconds != null) {
          _durationSeconds = metadata.durationSeconds!;
        }
        if (metadata.durationTotalSeconds != null) {
          _durationTotalSeconds = metadata.durationTotalSeconds;
        }
        if (metadata.distanceMeters != null) {
          _distanceKm = metadata.distanceMeters! / 1000;
        }
        _speedMaxKmh = metadata.speedMaxKmh;
        _speedAvgKmh = metadata.speedAvgKmh;
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
      debugPrint('Track apply failed: $e');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToParseTrack)),
      );
    } finally {
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

  void _applyRecordedTrack(RecordedTrack track) {
    setState(() {
      _trackFilename = 'track.gpx';
      _trackBytes = track.gpxBytes;
      _applyStartDateTime(track.startTime.toLocal());
      _durationSeconds = track.durationSeconds;
      _durationTotalSeconds = track.durationTotalSeconds;
      _distanceKm = track.distanceMeters / 1000;
      _speedMaxKmh = track.speedMaxKmh;
      _speedAvgKmh = track.speedAvgKmh;
    });
    _tabController?.animateTo(0);
    final l10n = AppLocalizations.of(context)!;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.trackMetadataApplied)),
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
        durationTotalSeconds: _durationTotalSeconds,
        distanceKm: _distanceKm,
        speedMaxKmh: _speedMaxKmh,
        speedAvgKmh: _speedAvgKmh,
        equipmentIds: _selectedEquipmentIds,
      );

      final Workout saved;
      if (_isEditing) {
        saved = await _api.updateWorkout(
          token: token,
          workoutId: widget.workout!.id,
          body: draft.toJson(),
        );
      } else {
        final fields = {
          'name': draft.name,
          if (draft.description.isNotEmpty) 'description': draft.description,
          'sport_type': draft.sportType,
          'start_date': draft.startDate.toUtc().toIso8601String(),
          'duration_seconds': draft.durationSeconds.toString(),
          if (draft.durationTotalSeconds != null)
            'duration_total_seconds': draft.durationTotalSeconds.toString(),
          'distance': (draft.distanceKm * 1000).toString(),
          if (draft.speedMaxKmh != null)
            'speed_max_kmh': draft.speedMaxKmh!.toStringAsFixed(2),
          if (draft.speedAvgKmh != null)
            'speed_avg_kmh': draft.speedAvgKmh!.toStringAsFixed(2),
          'equipment_ids': jsonEncode(draft.equipmentIds),
        };

        final hasTrack = _trackBytes != null && _trackFilename != null;
        final hasPhotos = _selectedPhotos.isNotEmpty;

        if (hasTrack || hasPhotos) {
          saved = await _api.createWorkoutMultipart(
            token: token,
            fields: fields,
            trackBytes: hasTrack ? _trackBytes : null,
            trackFilename: hasTrack ? _trackFilename : null,
            photos: hasPhotos ? _selectedPhotos : null,
          );
        } else {
          saved = await _api.createWorkout(token: token, body: draft.toJson());
        }
      }

      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.workoutSaved)),
      );
      Navigator.pop(context, saved);
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

  Widget _buildManualTab({required bool showTitle}) {
    return ManualWorkoutForm(
      formKey: _formKey,
      title: _isEditing
          ? AppLocalizations.of(context)!.editWorkoutTitle
          : AppLocalizations.of(context)!.addWorkout,
      nameController: _nameController,
      descriptionController: _descriptionController,
      sportTypeId: _sportTypeId,
      date: _date,
      time: _time,
      durationSeconds: _durationSeconds,
      distanceKm: _distanceKm,
      trackFilename: _trackFilename,
      equipment: _userEquipment,
      selectedEquipmentIds: _selectedEquipmentIds,
      selectedPhotos: _selectedPhotos,
      isSubmitting: _isSubmitting,
      isPickingFile: _isPickingFile,
      isParsingTrack: _isParsingTrack,
      isPickingPhotos: _isPickingPhotos,
      showTitle: showTitle,
      trackReadOnly: _isEditing,
      hidePhotos: _isEditing,
      onPickSportType: _pickSportType,
      onPickDate: _pickDate,
      onPickTime: _pickTime,
      onPickDuration: _pickDuration,
      onPickDistance: _pickDistance,
      onPickEquipment: _pickEquipment,
      onRemoveEquipment: _removeEquipment,
      onPickPhotos: _pickPhotos,
      onRemovePhoto: _removePhoto,
      onPickTrack: _pickTrack,
      onRemoveTrack: _removeTrack,
      onCancel: _handleCancel,
      onSubmit: _submit,
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) async {
        if (didPop) {
          return;
        }
        await _handleCancel();
      },
      child: Padding(
        padding: EdgeInsets.only(
          left: 24,
          right: 24,
          top: 24,
          bottom: 24 + MediaQuery.viewInsetsOf(context).bottom,
        ),
        child: _showRecordTab
            ? Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    _isEditing ? l10n.editWorkoutTitle : l10n.addWorkout,
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  TabBar(
                    controller: _tabController,
                    onTap: (index) {
                      if (_recorder.isActive && index == 0) {
                        _tabController?.index = 1;
                      }
                    },
                    tabs: [
                      Tab(
                        child: Opacity(
                          opacity: _recorder.isActive ? 0.38 : 1.0,
                          child: Text(l10n.tabManual),
                        ),
                      ),
                      Tab(text: l10n.tabRecord),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Expanded(
                    child: TabBarView(
                      controller: _tabController,
                      physics: _recorder.isActive
                          ? const NeverScrollableScrollPhysics()
                          : null,
                      children: [
                        _buildManualTab(showTitle: false),
                        RecordWorkoutTab(onFinished: _applyRecordedTrack),
                      ],
                    ),
                  ),
                ],
              )
            : _buildManualTab(showTitle: true),
      ),
    );
  }
}
