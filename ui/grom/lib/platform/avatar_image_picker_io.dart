import 'package:image_picker/image_picker.dart';

import 'avatar_image_picker_stub.dart';

Future<AvatarPickResult?> pickAvatarImage() async {
  final picked = await ImagePicker().pickImage(
    source: ImageSource.gallery,
    imageQuality: 95,
  );
  if (picked == null) {
    return null;
  }

  final bytes = await picked.readAsBytes();
  return (path: picked.path, bytes: bytes);
}
