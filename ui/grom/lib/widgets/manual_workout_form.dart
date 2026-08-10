import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/equipment.dart';
import 'package:grom/models/sport_types.dart';

import 'equipment_picker_field.dart';

/// Matches server `workouts.MaxPhotosPerWorkout`.
const kMaxPhotosPerWorkout = 20;

class ManualWorkoutForm extends StatelessWidget {
  const ManualWorkoutForm({
    super.key,
    required this.formKey,
    required this.title,
    required this.nameController,
    required this.descriptionController,
    required this.sportTypeId,
    required this.date,
    required this.time,
    required this.durationSeconds,
    required this.distanceKm,
    required this.trackFilename,
    required this.equipment,
    required this.selectedEquipmentIds,
    required this.selectedPhotos,
    this.existingPhotoFilenames = const [],
    this.existingPhotoPreviewUrl,
    this.authToken = '',
    required this.isSubmitting,
    required this.isPickingFile,
    required this.isParsingTrack,
    required this.isPickingPhotos,
    required this.showTitle,
    this.trackReadOnly = false,
    required this.onPickSportType,
    required this.onPickDate,
    required this.onPickTime,
    required this.onPickDuration,
    required this.onPickDistance,
    required this.onPickEquipment,
    required this.onRemoveEquipment,
    required this.onPickPhotos,
    required this.onRemovePhoto,
    this.onRemoveExistingPhoto,
    required this.onPickTrack,
    required this.onRemoveTrack,
    required this.onCancel,
    required this.onSubmit,
  });

  final GlobalKey<FormState> formKey;
  final String title;
  final TextEditingController nameController;
  final TextEditingController descriptionController;
  final String sportTypeId;
  final DateTime date;
  final TimeOfDay time;
  final int durationSeconds;
  final double distanceKm;
  final String? trackFilename;
  final List<Equipment> equipment;
  final List<String> selectedEquipmentIds;
  final List<({String filename, Uint8List bytes})> selectedPhotos;
  final List<String> existingPhotoFilenames;
  final String Function(String filename)? existingPhotoPreviewUrl;
  final String authToken;
  final bool isSubmitting;
  final bool isPickingFile;
  final bool isParsingTrack;
  final bool isPickingPhotos;
  final bool showTitle;
  final bool trackReadOnly;
  final VoidCallback onPickSportType;
  final VoidCallback onPickDate;
  final VoidCallback onPickTime;
  final VoidCallback onPickDuration;
  final VoidCallback onPickDistance;
  final VoidCallback onPickEquipment;
  final ValueChanged<String> onRemoveEquipment;
  final VoidCallback onPickPhotos;
  final ValueChanged<int> onRemovePhoto;
  final ValueChanged<String>? onRemoveExistingPhoto;
  final VoidCallback onPickTrack;
  final VoidCallback onRemoveTrack;
  final VoidCallback onCancel;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = l10n.localeName;
    final dateText = DateFormat.yMMMd(locale).format(date);
    final timeText = time.format(context);
    final durationText = formatDuration(l10n, durationSeconds);
    final distanceText = distanceKm == 0
        ? l10n.distanceZero
        : l10n.distanceKilometers(
            distanceKm >= 10
                ? distanceKm.toStringAsFixed(1)
                : distanceKm.toStringAsFixed(2),
          );
    final selectedSport = sportTypeById(sportTypeId);

    return Form(
      key: formKey,
      child: ScrollConfiguration(
        behavior: ScrollConfiguration.of(context).copyWith(scrollbars: false),
        child: SingleChildScrollView(
          // Space for OutlineInputBorder floating labels when title is outside
          // (mobile add sheet with TabBar); otherwise labels clip at the scroll top.
          padding: EdgeInsets.only(top: showTitle ? 0 : 8),
          child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (showTitle) ...[
              Text(
                title,
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 20),
            ],
            TextFormField(
              controller: nameController,
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
              controller: descriptionController,
              decoration: InputDecoration(
                labelText: l10n.workoutDescription,
                border: const OutlineInputBorder(),
              ),
              minLines: 3,
              maxLines: 5,
            ),
            const SizedBox(height: 16),
            InkWell(
              onTap: onPickSportType,
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
                      child: Text(sportTypeLabel(l10n, sportTypeId)),
                    ),
                    const Icon(Icons.arrow_drop_down),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            InkWell(
              onTap: onPickDate,
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
              onTap: onPickTime,
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
              onTap: onPickDuration,
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
              onTap: onPickDistance,
              child: InputDecorator(
                decoration: InputDecoration(
                  labelText: l10n.workoutDistance,
                  border: const OutlineInputBorder(),
                ),
                child: Text(distanceText),
              ),
            ),
            const SizedBox(height: 16),
            EquipmentPickerField(
              equipment: equipment,
              selectedIds: selectedEquipmentIds,
              isSubmitting: isSubmitting,
              onPick: onPickEquipment,
              onRemove: onRemoveEquipment,
            ),
            const SizedBox(height: 16),
            _WorkoutPhotosField(
              existingFilenames: existingPhotoFilenames,
              existingPreviewUrl: existingPhotoPreviewUrl,
              authToken: authToken,
              photos: selectedPhotos,
              canAddPhotos: existingPhotoFilenames.length + selectedPhotos.length <
                  kMaxPhotosPerWorkout,
              isSubmitting: isSubmitting,
              isPickingPhotos: isPickingPhotos,
              onPickPhotos: onPickPhotos,
              onRemovePhoto: onRemovePhoto,
              onRemoveExistingPhoto: onRemoveExistingPhoto,
            ),
            const SizedBox(height: 16),
            if (!trackReadOnly || trackFilename != null) ...[
              _TrackPickerField(
                trackFilename: trackFilename,
                isSubmitting: isSubmitting,
                isPickingFile: isPickingFile,
                isParsingTrack: isParsingTrack,
                readOnly: trackReadOnly,
                onPickTrack: onPickTrack,
                onRemoveTrack: onRemoveTrack,
              ),
              const SizedBox(height: 24),
            ] else
              const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: OutlinedButton(
                    onPressed: isSubmitting ? null : onCancel,
                    child: Text(l10n.cancel),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: FilledButton(
                    onPressed: isSubmitting ? null : onSubmit,
                    child: isSubmitting
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

class _WorkoutPhotosField extends StatelessWidget {
  const _WorkoutPhotosField({
    required this.existingFilenames,
    required this.existingPreviewUrl,
    required this.authToken,
    required this.photos,
    required this.canAddPhotos,
    required this.isSubmitting,
    required this.isPickingPhotos,
    required this.onPickPhotos,
    required this.onRemovePhoto,
    this.onRemoveExistingPhoto,
  });

  final List<String> existingFilenames;
  final String Function(String filename)? existingPreviewUrl;
  final String authToken;
  final List<({String filename, Uint8List bytes})> photos;
  final bool canAddPhotos;
  final bool isSubmitting;
  final bool isPickingPhotos;
  final VoidCallback onPickPhotos;
  final ValueChanged<int> onRemovePhoto;
  final ValueChanged<String>? onRemoveExistingPhoto;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final totalCount = existingFilenames.length + photos.length;
    final headers = authToken.isEmpty
        ? null
        : {'Authorization': 'Bearer $authToken'};

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        FilledButton.tonalIcon(
          onPressed: (isSubmitting || isPickingPhotos || !canAddPhotos)
              ? null
              : onPickPhotos,
          icon: isPickingPhotos
              ? const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.add_photo_alternate_outlined),
          label: Text(l10n.addPhotos),
        ),
        if (totalCount > 0) ...[
          const SizedBox(height: 8),
          Text(l10n.photosSelected(totalCount)),
          const SizedBox(height: 8),
          SizedBox(
            height: 72,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: totalCount,
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemBuilder: (context, index) {
                if (index < existingFilenames.length) {
                  final filename = existingFilenames[index];
                  final url = existingPreviewUrl?.call(filename);
                  return _PhotoThumb(
                    isSubmitting: isSubmitting,
                    onRemove: onRemoveExistingPhoto == null
                        ? null
                        : () => onRemoveExistingPhoto!(filename),
                    removeTooltip: l10n.removePhoto,
                    child: url == null || url.isEmpty
                        ? const ColoredBox(
                            color: Color(0xFFE0E0E0),
                            child: Icon(Icons.broken_image_outlined),
                          )
                        : Image.network(
                            url,
                            headers: headers,
                            width: 72,
                            height: 72,
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stackTrace) {
                              return const ColoredBox(
                                color: Color(0xFFE0E0E0),
                                child: Icon(Icons.broken_image_outlined),
                              );
                            },
                          ),
                  );
                }
                final photo = photos[index - existingFilenames.length];
                final newIndex = index - existingFilenames.length;
                return _PhotoThumb(
                  isSubmitting: isSubmitting,
                  onRemove: () => onRemovePhoto(newIndex),
                  removeTooltip: l10n.removePhoto,
                  child: Image.memory(
                    photo.bytes,
                    width: 72,
                    height: 72,
                    fit: BoxFit.cover,
                  ),
                );
              },
            ),
          ),
        ],
      ],
    );
  }
}

class _PhotoThumb extends StatelessWidget {
  const _PhotoThumb({
    required this.child,
    required this.isSubmitting,
    required this.removeTooltip,
    this.onRemove,
  });

  final Widget child;
  final bool isSubmitting;
  final String removeTooltip;
  final VoidCallback? onRemove;

  @override
  Widget build(BuildContext context) {
    return Stack(
      clipBehavior: Clip.none,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: SizedBox(
            width: 72,
            height: 72,
            child: child,
          ),
        ),
        if (onRemove != null)
          Positioned(
            top: -8,
            right: -8,
            child: IconButton(
              tooltip: removeTooltip,
              visualDensity: VisualDensity.compact,
              onPressed: isSubmitting ? null : onRemove,
              icon: const Icon(Icons.cancel, size: 20),
            ),
          ),
      ],
    );
  }
}

class _TrackPickerField extends StatelessWidget {
  const _TrackPickerField({
    required this.trackFilename,
    required this.isSubmitting,
    required this.isPickingFile,
    required this.isParsingTrack,
    this.readOnly = false,
    required this.onPickTrack,
    required this.onRemoveTrack,
  });

  final String? trackFilename;
  final bool isSubmitting;
  final bool isPickingFile;
  final bool isParsingTrack;
  final bool readOnly;
  final VoidCallback onPickTrack;
  final VoidCallback onRemoveTrack;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    if (trackFilename != null) {
      return InputDecorator(
        decoration: InputDecoration(
          labelText: l10n.workoutTrack,
          border: const OutlineInputBorder(),
        ),
        child: Row(
          children: [
            Expanded(
              child: Text(
                l10n.trackFileSelected(trackFilename!),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            if (isParsingTrack)
              const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            else if (!readOnly)
              IconButton(
                tooltip: l10n.removeTrack,
                onPressed: isSubmitting ? null : onRemoveTrack,
                icon: const Icon(Icons.close),
              ),
          ],
        ),
      );
    }

    if (readOnly) {
      return const SizedBox.shrink();
    }

    return FilledButton.tonalIcon(
      onPressed: (isSubmitting || isPickingFile || isParsingTrack)
          ? null
          : onPickTrack,
      icon: const Icon(Icons.upload_file),
      label: Text(l10n.selectTrackFile),
    );
  }
}

Future<int?> showDurationPickerDialog(
  BuildContext context, {
  required int initialSeconds,
}) {
  return showDialog<int>(
    context: context,
    builder: (context) => _DurationPickerDialog(initialSeconds: initialSeconds),
  );
}

Future<double?> showDistancePickerDialog(
  BuildContext context, {
  required double initialDistanceKm,
}) {
  return showDialog<double>(
    context: context,
    builder: (context) =>
        _DistancePickerDialog(initialDistanceKm: initialDistanceKm),
  );
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
