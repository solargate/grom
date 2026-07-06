import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';
import 'package:travka/models/sport_types.dart';
import 'package:travka/platform/is_mobile_client.dart';
import 'package:travka/platform/track_file_picker.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/recorded_track.dart';
import '../models/workout.dart';
import '../services/track_recording_service.dart';
import 'manual_workout_form.dart';
import 'record_workout_tab.dart';

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
    isDismissible: false,
    enableDrag: false,
    builder: (context) => const AddWorkoutSheet(),
  );
}

class AddWorkoutSheet extends StatefulWidget {
  const AddWorkoutSheet({super.key});

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

  TabController? _tabController;

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

  bool get _showRecordTab => isMobileClient;

  @override
  void initState() {
    super.initState();
    final now = DateTime.now();
    _date = DateTime(now.year, now.month, now.day);
    _time = TimeOfDay.fromDateTime(now);
    if (_showRecordTab) {
      _tabController = TabController(length: 2, vsync: this);
    }
  }

  @override
  void dispose() {
    _tabController?.dispose();
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
    Navigator.pop(context, false);
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

  void _applyRecordedTrack(RecordedTrack track) {
    final localStart = track.startTime.toLocal();
    setState(() {
      _trackFilename = 'track.gpx';
      _trackBytes = track.gpxBytes;
      _date = DateTime(localStart.year, localStart.month, localStart.day);
      _time = TimeOfDay(hour: localStart.hour, minute: localStart.minute);
      _durationSeconds = track.durationSeconds;
      _distanceKm = track.distanceMeters / 1000;
    });
    _tabController?.animateTo(1);
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

  Widget _buildManualTab({required bool showTitle}) {
    return ManualWorkoutForm(
      formKey: _formKey,
      nameController: _nameController,
      descriptionController: _descriptionController,
      sportTypeId: _sportTypeId,
      date: _date,
      time: _time,
      durationSeconds: _durationSeconds,
      distanceKm: _distanceKm,
      trackFilename: _trackFilename,
      isSubmitting: _isSubmitting,
      isPickingFile: _isPickingFile,
      isParsingTrack: _isParsingTrack,
      showTitle: showTitle,
      onPickSportType: _pickSportType,
      onPickDate: _pickDate,
      onPickTime: _pickTime,
      onPickDuration: _pickDuration,
      onPickDistance: _pickDistance,
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
                    l10n.addWorkout,
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  TabBar(
                    controller: _tabController,
                    onTap: (index) async {
                      if (index == 0 ||
                          !_recorder.isActive ||
                          _recorder.state == TrackRecordingState.paused) {
                        return;
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
                      } else {
                        _tabController?.index = 0;
                      }
                    },
                    tabs: [
                      Tab(text: l10n.tabRecord),
                      Tab(text: l10n.tabManual),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Expanded(
                    child: TabBarView(
                      controller: _tabController,
                      children: [
                        RecordWorkoutTab(onFinished: _applyRecordedTrack),
                        _buildManualTab(showTitle: false),
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
