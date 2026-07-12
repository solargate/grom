import 'dart:convert';
import 'dart:io';

import 'package:grom/api_request.dart';
import 'package:grom/server_storage.dart';
import 'package:http/http.dart' as http;

import 'strava_archive_types.dart';

Future<Map<String, dynamic>> uploadStravaArchivePick({
  required ApiRequest api,
  required String token,
  required StravaArchivePick pick,
  void Function(double progress)? onProgress,
}) async {
  if (pick.path != null) {
    return _uploadRawFromPath(
      token: token,
      path: pick.path!,
      onProgress: onProgress,
    );
  }
  if (pick.bytes != null) {
    return api.uploadStravaArchiveRaw(
      token: token,
      bytes: pick.bytes!,
      onProgress: onProgress,
    );
  }
  if (pick.stream != null && pick.size != null && pick.size! > 0) {
    return api.uploadStravaArchiveRawStream(
      token: token,
      stream: pick.stream!,
      length: pick.size!,
      onProgress: onProgress,
    );
  }
  throw ApiException('Selected archive file is unavailable');
}

Future<Map<String, dynamic>> _uploadRawFromPath({
  required String token,
  required String path,
  void Function(double progress)? onProgress,
}) async {
  final file = File(path);
  final length = await file.length();
  if (length == 0) {
    throw ApiException('Selected archive file is empty');
  }

  final request = http.StreamedRequest(
    'POST',
    _uri('/api/v1/integrations/strava/import'),
  );
  request.headers['Authorization'] = 'Bearer $token';
  request.headers['Content-Type'] = 'application/zip';
  request.contentLength = length;

  var sent = 0;
  await for (final chunk in file.openRead()) {
    request.sink.add(chunk);
    sent += chunk.length;
    if (onProgress != null && length > 0) {
      onProgress(sent / length);
    }
  }
  await request.sink.close();

  final response = await http.Response.fromStream(await request.send());
  if (response.statusCode == 202) {
    return jsonDecode(response.body) as Map<String, dynamic>;
  }
  throw ApiException(
    _parseErrorMessage(response),
    statusCode: response.statusCode,
  );
}

Uri _uri(String path) {
  final base = ServerStorage.cachedBaseUrl;
  if (base == null || base.isEmpty) {
    return Uri.parse(path);
  }
  return Uri.parse(base).resolve(path.startsWith('/') ? path.substring(1) : path);
}

String _parseErrorMessage(http.Response response) {
  try {
    final json = jsonDecode(response.body) as Map<String, dynamic>;
    return json['error'] as String? ?? 'Unknown error';
  } catch (_) {
    return 'Request failed (${response.statusCode})';
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
