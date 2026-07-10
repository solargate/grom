import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:image/image.dart' as img;
import 'package:image_cropper/image_cropper.dart';

import '../platform/avatar_image_picker.dart';
import '../widgets/avatar_crop_page.dart';

class AvatarImageService {
  static const avatarSize = 256;

  static Future<Uint8List?> pickCropAndEncode(
    BuildContext context, {
    BuildContext? navigatorContext,
  }) async {
    final picked = await pickAvatarImage();
    if (picked == null) {
      return null;
    }

    final cropContext = navigatorContext ?? context;
    if (!cropContext.mounted) {
      return null;
    }

    if (kIsWeb) {
      return _cropOnWeb(cropContext, picked.bytes);
    }

    final cropped = await ImageCropper().cropImage(
      sourcePath: picked.path,
      aspectRatio: const CropAspectRatio(ratioX: 1, ratioY: 1),
      compressQuality: 95,
      uiSettings: [
        AndroidUiSettings(
          toolbarTitle: 'Avatar',
          cropStyle: CropStyle.circle,
          lockAspectRatio: true,
          aspectRatioPresets: [CropAspectRatioPreset.square],
        ),
        IOSUiSettings(
          title: 'Avatar',
          cropStyle: CropStyle.circle,
          aspectRatioLockEnabled: true,
          aspectRatioPresets: [CropAspectRatioPreset.square],
        ),
      ],
    );

    if (cropped == null) {
      return null;
    }

    final bytes = await cropped.readAsBytes();
    return encodeAvatarBytes(bytes);
  }

  static Future<Uint8List?> _cropOnWeb(
    BuildContext cropContext,
    Uint8List imageBytes,
  ) async {
    final cropped = await Navigator.of(cropContext).push<Uint8List>(
      MaterialPageRoute(
        fullscreenDialog: true,
        builder: (context) => AvatarCropPage(imageBytes: imageBytes),
      ),
    );
    if (cropped == null) {
      return null;
    }
    return encodeAvatarBytes(cropped);
  }

  static Future<Uint8List> encodeAvatarBytes(Uint8List croppedBytes) async {
    final decoded = img.decodeImage(croppedBytes);
    if (decoded == null) {
      throw StateError('Failed to decode cropped image');
    }

    final square = img.copyResizeCropSquare(decoded, size: avatarSize);
    return Uint8List.fromList(img.encodePng(square));
  }
}
