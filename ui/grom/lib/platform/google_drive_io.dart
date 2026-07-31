import 'dart:io';

import 'package:extension_google_sign_in_as_googleapis_auth/extension_google_sign_in_as_googleapis_auth.dart';
import 'package:flutter/foundation.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:googleapis/drive/v3.dart' as drive;

export 'google_drive_stub.dart'
    show
        GoogleDriveError,
        GoogleDriveException,
        GoogleDriveFileEntry,
        GoogleDriveFolder;

import 'google_drive_stub.dart'
    show
        GoogleDriveError,
        GoogleDriveException,
        GoogleDriveFileEntry,
        GoogleDriveFolder;

bool get _isAndroidDriveSupported =>
    !kIsWeb && Platform.isAndroid;

final GoogleSignIn _googleSignIn = GoogleSignIn(
  scopes: const [drive.DriveApi.driveReadonlyScope],
);

Future<drive.DriveApi?> _driveApi() async {
  if (!_isAndroidDriveSupported) {
    throw GoogleDriveException(GoogleDriveError.unsupported);
  }

  var account = await _googleSignIn.signInSilently();
  account ??= await _googleSignIn.signIn();
  if (account == null) {
    throw GoogleDriveException(GoogleDriveError.cancelled);
  }

  final client = await _googleSignIn.authenticatedClient();
  if (client == null) {
    throw GoogleDriveException(GoogleDriveError.signInFailed);
  }

  return drive.DriveApi(client);
}

GoogleDriveException _mapDriveError(Object error) {
  final message = error.toString().toLowerCase();
  if (message.contains('access_denied') || message.contains('403')) {
    return GoogleDriveException(GoogleDriveError.accessDenied, message: error.toString());
  }
  if (message.contains('sign_in') || message.contains('sign in')) {
    return GoogleDriveException(GoogleDriveError.signInFailed, message: error.toString());
  }
  return GoogleDriveException(GoogleDriveError.signInFailed, message: error.toString());
}

Future<void> ensureGoogleDriveSignedIn() async {
  try {
    await _driveApi();
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}

Future<GoogleDriveFolder?> _firstMatchingFolder(String query) async {
  final api = await _driveApi();
  if (api == null) {
    return null;
  }

  final response = await api.files.list(
    q: query,
    spaces: 'drive',
    $fields: 'files(id,name)',
    pageSize: 1,
  );

  final files = response.files;
  if (files == null || files.isEmpty) {
    return null;
  }

  final file = files.first;
  final id = file.id;
  final name = file.name;
  if (id == null || name == null) {
    return null;
  }
  return GoogleDriveFolder(id: id, name: name);
}

Future<GoogleDriveFolder?> findHealthSyncFolderByNameContains() async {
  try {
    return _firstMatchingFolder(
      "mimeType='application/vnd.google-apps.folder' and name contains 'Health Sync' and trashed=false",
    );
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}

Future<GoogleDriveFolder?> findFolderByExactName(String name) async {
  final trimmed = name.trim();
  if (trimmed.isEmpty) {
    return null;
  }

  try {
    final escaped = trimmed.replaceAll("'", r"\'");
    return _firstMatchingFolder(
      "mimeType='application/vnd.google-apps.folder' and name = '$escaped' and trashed=false",
    );
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}

Future<List<GoogleDriveFileEntry>> listFolderFiles(String folderId) async {
  try {
    final api = await _driveApi();
    if (api == null) {
      return const [];
    }

    final entries = <GoogleDriveFileEntry>[];
    String? pageToken;

    do {
      final response = await api.files.list(
        q: "'$folderId' in parents and trashed=false",
        spaces: 'drive',
        $fields: 'nextPageToken,files(id,name)',
        pageSize: 200,
        pageToken: pageToken,
      );

      for (final file in response.files ?? const []) {
        final id = file.id;
        final name = file.name;
        if (id != null && name != null) {
          entries.add(GoogleDriveFileEntry(id: id, name: name));
        }
      }
      pageToken = response.nextPageToken;
    } while (pageToken != null);

    return entries;
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}

Future<List<int>> downloadDriveFile(String fileId) async {
  try {
    final api = await _driveApi();
    if (api == null) {
      throw GoogleDriveException(GoogleDriveError.unsupported);
    }

    final media = await api.files.get(
      fileId,
      downloadOptions: drive.DownloadOptions.fullMedia,
    ) as drive.Media;

    final chunks = <int>[];
    await for (final chunk in media.stream) {
      chunks.addAll(chunk);
    }
    return chunks;
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}
