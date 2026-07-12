import 'dart:convert';
import 'dart:js_interop';

import 'package:grom/api_request.dart';
import 'package:grom/server_storage.dart';
import 'package:web/web.dart' as web;

import 'strava_archive_types.dart';

Future<Map<String, dynamic>> uploadStravaArchivePick({
  required ApiRequest api,
  required String token,
  required StravaArchivePick pick,
  void Function(double progress)? onProgress,
}) async {
  final nativeFile = pick.nativeFile;
  if (nativeFile is web.File) {
    if (onProgress != null) {
      onProgress(0);
    }

    final headers = web.Headers();
    headers.set('Authorization', 'Bearer $token');
    headers.set('Content-Type', 'application/zip');

    final response = await web.window.fetch(
      _url('/api/v1/integrations/strava/import').toJS,
      web.RequestInit(
        method: 'POST',
        headers: headers,
        body: nativeFile,
      ),
    ).toDart;

    if (onProgress != null) {
      onProgress(1);
    }

    final body = (await response.text().toDart).toDart;
    if (response.status == 202) {
      return jsonDecode(body) as Map<String, dynamic>;
    }
    throw ApiException(
      _parseErrorMessage(body, response.status),
      statusCode: response.status,
    );
  }

  if (pick.bytes != null) {
    return api.uploadStravaArchiveRaw(
      token: token,
      bytes: pick.bytes!,
      onProgress: onProgress,
    );
  }

  throw ApiException('Selected archive file is unavailable');
}

String _url(String path) {
  final base = ServerStorage.cachedBaseUrl;
  if (base == null || base.isEmpty) {
    return path;
  }
  return Uri.parse(base).resolve(path.startsWith('/') ? path.substring(1) : path).toString();
}

String _parseErrorMessage(String body, int statusCode) {
  try {
    final json = jsonDecode(body) as Map<String, dynamic>;
    return json['error'] as String? ?? 'Unknown error';
  } catch (_) {
    return 'Request failed ($statusCode)';
  }
}

Future<Map<String, dynamic>> uploadStravaArchivePickImpl({
  required ApiRequest api,
  required String token,
  required StravaArchivePick pick,
  void Function(double progress)? onProgress,
}) =>
    uploadStravaArchivePick(
      api: api,
      token: token,
      pick: pick,
      onProgress: onProgress,
    );
