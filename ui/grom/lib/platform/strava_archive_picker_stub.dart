import 'strava_archive_types.dart';

Future<StravaArchivePick?> pickStravaArchiveFile() async {
  throw UnsupportedError('Strava archive picker is not supported on this platform');
}

Future<StravaArchivePick?> pickStravaArchiveFileImpl() => pickStravaArchiveFile();
