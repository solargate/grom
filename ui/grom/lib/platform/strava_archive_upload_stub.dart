import 'package:grom/api_request.dart';

import 'strava_archive_types.dart';

Future<Map<String, dynamic>> uploadStravaArchivePick({
  required ApiRequest api,
  required String token,
  required StravaArchivePick pick,
  void Function(double progress)? onProgress,
}) {
  throw UnsupportedError('Strava archive upload is not supported on this platform');
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
