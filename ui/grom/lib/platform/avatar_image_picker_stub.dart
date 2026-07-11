import 'dart:typed_data';

typedef AvatarPickResult = ({String path, Uint8List bytes});

Future<AvatarPickResult?> pickAvatarImage() =>
    throw UnsupportedError('Cannot pick avatar without platform implementation');
