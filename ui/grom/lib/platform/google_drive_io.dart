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

bool get _isAndroidDriveSupported => !kIsWeb && Platform.isAndroid;

const _driveScopes = <String>[
  drive.DriveApi.driveReadonlyScope,
];

final GoogleSignIn _googleSignIn = GoogleSignIn(
  scopes: _driveScopes,
);

bool _looksLikeInvalidOrDeniedToken(Object error) {
  final message = error.toString().toLowerCase();
  return message.contains('invalid_token') ||
      message.contains('access was denied') ||
      message.contains('access_denied') ||
      message.contains('403');
}

Future<void> _ensureDriveScopesGranted() async {
  // Android-only path: do not call canAccessScopes (web-only; throws
  // UnimplementedError on mobile). Scopes are also set on GoogleSignIn;
  // requestScopes is a no-op when already granted.
  final granted = await _googleSignIn.requestScopes(_driveScopes);
  if (!granted) {
    throw GoogleDriveException(GoogleDriveError.accessDenied);
  }
}

Future<drive.DriveApi> _createDriveApi({required bool forceInteractive}) async {
  if (!_isAndroidDriveSupported) {
    throw GoogleDriveException(GoogleDriveError.unsupported);
  }

  GoogleSignInAccount? account;
  if (!forceInteractive) {
    account = await _googleSignIn.signInSilently();
  }
  account ??= await _googleSignIn.signIn();
  if (account == null) {
    throw GoogleDriveException(GoogleDriveError.cancelled);
  }

  await _ensureDriveScopesGranted();

  // Re-read auth after requestScopes so Drive gets a token that includes the scope.
  await _googleSignIn.currentUser?.authentication;

  final client = await _googleSignIn.authenticatedClient();
  if (client == null) {
    throw GoogleDriveException(
      GoogleDriveError.signInFailed,
      message: 'authenticatedClient returned null',
    );
  }
  return drive.DriveApi(client);
}

Future<void> _disconnectQuietly() async {
  try {
    await _googleSignIn.disconnect();
  } catch (_) {
    // Best-effort revoke of a stale / incomplete grant.
  }
}

/// Revokes the app's Google Drive grant. No-op off Android.
Future<void> disconnectGoogleDrive() async {
  if (!_isAndroidDriveSupported) {
    return;
  }
  await _disconnectQuietly();
}

/// Runs [run] with a Drive client. On invalid/denied token, disconnects once and
/// retries with an interactive sign-in + explicit Drive scope consent.
Future<T> _withDriveApi<T>(Future<T> Function(drive.DriveApi api) run) async {
  Future<T> attempt({required bool forceInteractive}) async {
    final api = await _createDriveApi(forceInteractive: forceInteractive);
    return run(api);
  }

  try {
    return await attempt(forceInteractive: false);
  } catch (error) {
    if (error is GoogleDriveException) {
      if (error.error == GoogleDriveError.cancelled ||
          error.error == GoogleDriveError.unsupported ||
          error.error == GoogleDriveError.signInFailed) {
        rethrow;
      }
      if (error.error != GoogleDriveError.accessDenied &&
          !_looksLikeInvalidOrDeniedToken(error)) {
        rethrow;
      }
    } else if (!_looksLikeInvalidOrDeniedToken(error)) {
      throw _mapDriveError(error);
    }

    await _disconnectQuietly();

    try {
      return await attempt(forceInteractive: true);
    } on GoogleDriveException {
      rethrow;
    } catch (retryError) {
      throw _mapDriveError(retryError);
    }
  }
}

GoogleDriveException _mapDriveError(Object error) {
  final message = error.toString().toLowerCase();
  if (_looksLikeInvalidOrDeniedToken(error)) {
    return GoogleDriveException(
      GoogleDriveError.accessDenied,
      message: error.toString(),
    );
  }
  if (message.contains('sign_in') || message.contains('sign in')) {
    return GoogleDriveException(
      GoogleDriveError.signInFailed,
      message: error.toString(),
    );
  }
  return GoogleDriveException(
    GoogleDriveError.signInFailed,
    message: error.toString(),
  );
}

Future<void> ensureGoogleDriveSignedIn() async {
  try {
    await _withDriveApi((_) async {});
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}

Future<GoogleDriveFolder?> _firstMatchingFolder(String query) async {
  return _withDriveApi((api) async {
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
  });
}

Future<GoogleDriveFolder?> findHealthSyncFolderByNameContains() async {
  try {
    return await _firstMatchingFolder(
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
    return await _firstMatchingFolder(
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
    return await _withDriveApi((api) async {
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
    });
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}

Future<List<int>> downloadDriveFile(String fileId) async {
  try {
    return await _withDriveApi((api) async {
      final media = await api.files.get(
        fileId,
        downloadOptions: drive.DownloadOptions.fullMedia,
      ) as drive.Media;

      final chunks = <int>[];
      await for (final chunk in media.stream) {
        chunks.addAll(chunk);
      }
      return chunks;
    });
  } on GoogleDriveException {
    rethrow;
  } catch (error) {
    throw _mapDriveError(error);
  }
}
