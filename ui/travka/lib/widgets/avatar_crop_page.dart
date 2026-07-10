import 'dart:typed_data';

import 'package:crop_your_image/crop_your_image.dart';
import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

class AvatarCropPage extends StatefulWidget {
  const AvatarCropPage({
    super.key,
    required this.imageBytes,
  });

  final Uint8List imageBytes;

  @override
  State<AvatarCropPage> createState() => _AvatarCropPageState();
}

class _AvatarCropPageState extends State<AvatarCropPage> {
  final CropController _controller = CropController();

  void _onCropped(CropResult result) {
    final l10n = AppLocalizations.of(context)!;
    switch (result) {
      case CropSuccess(:final croppedImage):
        Navigator.of(context).pop(croppedImage);
      case CropFailure(:final cause):
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('${l10n.failedToUploadAvatar}: $cause')),
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.cropAvatarTitle),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => Navigator.of(context).pop(),
        ),
        actions: [
          TextButton(
            onPressed: _controller.cropCircle,
            child: Text(l10n.cropAvatarDone),
          ),
        ],
      ),
      body: Crop(
        image: widget.imageBytes,
        controller: _controller,
        withCircleUi: true,
        interactive: true,
        baseColor: theme.colorScheme.surface,
        maskColor: Colors.black.withValues(alpha: 0.55),
        onCropped: _onCropped,
      ),
    );
  }
}
