class DownloadedTrack {
  const DownloadedTrack({
    required this.bytes,
    required this.filename,
  });

  final List<int> bytes;
  final String filename;
}
