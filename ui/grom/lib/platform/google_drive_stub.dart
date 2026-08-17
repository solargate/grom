class GoogleDriveFolder {
  const GoogleDriveFolder({required this.id, required this.name});

  final String id;
  final String name;
}

class GoogleDriveFileEntry {
  const GoogleDriveFileEntry({required this.id, required this.name});

  final String id;
  final String name;
}

enum GoogleDriveError {
  cancelled,
  signInFailed,
  accessDenied,
  unsupported,
}

class GoogleDriveException implements Exception {
  GoogleDriveException(this.error, {this.message});

  final GoogleDriveError error;
  final String? message;

  @override
  String toString() => message ?? error.name;
}

Future<void> ensureGoogleDriveSignedIn() async {
  throw GoogleDriveException(GoogleDriveError.unsupported);
}

Future<void> disconnectGoogleDrive() async {}

Future<GoogleDriveFolder?> findHealthSyncFolderByNameContains() async {
  return null;
}

Future<GoogleDriveFolder?> findFolderByExactName(String name) async {
  return null;
}

Future<List<GoogleDriveFileEntry>> listFolderFiles(String folderId) async {
  return const [];
}

Future<List<int>> downloadDriveFile(String fileId) async {
  throw GoogleDriveException(GoogleDriveError.unsupported);
}
