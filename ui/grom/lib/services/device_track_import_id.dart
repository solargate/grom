import 'package:crypto/crypto.dart';

/// Namespace for device file-picker imports (`external_id.name`).
const deviceImportExternalIDName = 'device-import';

/// Stable dedup id: `{basename_lower}:{sha256_16}`.
String deviceImportExternalID({
  required String filename,
  required List<int> bytes,
}) {
  final basename = filename.trim().toLowerCase();
  final digest = sha256.convert(bytes);
  final hash16 = digest.toString().substring(0, 16);
  return '$basename:$hash16';
}

bool isTrackFilename(String filename) {
  final lower = filename.trim().toLowerCase();
  return lower.endsWith('.gpx') || lower.endsWith('.fit');
}

String trackBasenameWithoutExtension(String filename) {
  final name = filename.trim();
  final lower = name.toLowerCase();
  if (lower.endsWith('.gpx') || lower.endsWith('.fit')) {
    return name.substring(0, name.length - 4);
  }
  return name;
}
