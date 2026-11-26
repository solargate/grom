part of 'api_request.dart';

mixin _IntegrationsApi on _ApiTransport {
  Future<Map<String, dynamic>> uploadStravaArchiveRaw({
    required String token,
    required List<int> bytes,
    void Function(double progress)? onProgress,
  }) {
    return uploadStravaArchiveRawStream(
      token: token,
      stream: Stream.value(bytes),
      length: bytes.length,
      onProgress: onProgress,
    );
  }

  Future<Map<String, dynamic>> uploadStravaArchiveRawStream({
    required String token,
    required Stream<List<int>> stream,
    required int length,
    void Function(double progress)? onProgress,
  }) async {
    final abort = Completer<void>();
    var sent = 0;
    final request = _ArchiveRequest(
      _uri('/api/v1/integrations/strava/import'),
      stream.map((chunk) {
        sent += chunk.length;
        if (length > 0) onProgress?.call(sent / length);
        return chunk;
      }),
      abort.future,
    )
      ..headers.addAll(_bearerHeaders(token, contentType: 'application/zip'))
      ..contentLength = length;

    onProgress?.call(0);
    final response = await _send(
      request,
      timeout: _archiveTimeout,
      onTimeout: abort.complete,
      onHeaders: () => onProgress?.call(1),
    );
    if (response.statusCode == 202) {
      return _decode<Map<String, dynamic>>(response);
    }
    throw _parseError(response, authenticated: true);
  }

  Future<Map<String, dynamic>> uploadStravaArchive({
    required String token,
    required List<int> bytes,
    required String filename,
    void Function(double progress)? onProgress,
  }) async {
    return uploadStravaArchiveRaw(
      token: token,
      bytes: bytes,
      onProgress: onProgress,
    );
  }

  Future<Map<String, dynamic>> uploadStravaArchiveFromStream({
    required String token,
    required Stream<List<int>> stream,
    required int length,
    required String filename,
    void Function(double progress)? onProgress,
  }) async {
    return uploadStravaArchiveRawStream(
      token: token,
      stream: stream,
      length: length,
      onProgress: onProgress,
    );
  }

  Future<Map<String, dynamic>> getStravaImportStatus(String token) async {
    final response = await _request(
      'GET',
      _uri('/api/v1/integrations/strava/import/status'),
      headers: _bearerHeaders(token),
    );

    if (response.statusCode == 200) {
      return _decode<Map<String, dynamic>>(response);
    }

    throw _parseError(response, authenticated: true);
  }
}

// Hand the source stream to the HTTP client for backpressure and cancellation.
class _ArchiveRequest extends http.BaseRequest with http.Abortable {
  _ArchiveRequest(Uri uri, this._source, this.abortTrigger)
      : super('POST', uri);

  final Stream<List<int>> _source;

  @override
  final Future<void> abortTrigger;

  @override
  http.ByteStream finalize() {
    super.finalize();
    return http.ByteStream(_source);
  }
}
