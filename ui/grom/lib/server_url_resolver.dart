import 'server_storage.dart';
import 'platform/server_scheme_probe.dart';

export 'platform/server_scheme_probe_result.dart';

typedef SchemeProbe = Future<SchemeProbeResult> Function(Uri statusUri);

const _statusPathSuffix = '/api/v1/status';

/// Whether scheme auto-detection should run for [input].
///
/// Skips when the user already provided `http(s)://` or an explicit port.
bool needsSchemeProbe(String input) {
  final trimmed = input.trim();
  if (trimmed.isEmpty) {
    return false;
  }

  final lower = trimmed.toLowerCase();
  if (lower.startsWith('http://') || lower.startsWith('https://')) {
    return false;
  }
  if (trimmed.contains('://')) {
    return false;
  }

  final provisional = Uri.tryParse('https://$trimmed');
  if (provisional == null || provisional.host.isEmpty) {
    return false;
  }
  if (provisional.hasPort) {
    return false;
  }

  return true;
}

Uri statusUriForBase(String baseUrl) {
  final normalized = ServerStorage.normalizeBaseUrl(baseUrl);
  return Uri.parse('$normalized$_statusPathSuffix');
}

/// Derives a server base URL from a final `/api/v1/status` URI (after redirects).
String baseUrlFromStatusUri(Uri statusUri) {
  var path = statusUri.path;
  if (path.endsWith('/')) {
    path = path.substring(0, path.length - 1);
  }
  if (path.endsWith(_statusPathSuffix)) {
    path = path.substring(0, path.length - _statusPathSuffix.length);
  }

  final buffer = StringBuffer('${statusUri.scheme}://${statusUri.host}');
  if (statusUri.hasPort) {
    buffer.write(':${statusUri.port}');
  }
  if (path.isNotEmpty && path != '/') {
    buffer.write(path);
  }

  return ServerStorage.normalizeBaseUrl(buffer.toString());
}

/// Resolves a user-entered server locator to a normalized base URL.
///
/// When [needsSchemeProbe] is true, tries HTTPS then HTTP against `/api/v1/status`.
/// TLS/certificate errors count as HTTPS. If neither responds, falls back to HTTPS.
Future<String> resolveServerBaseUrl(
  String input, {
  SchemeProbe? probe,
}) async {
  final trimmed = input.trim();
  if (trimmed.isEmpty) {
    return trimmed;
  }

  if (!needsSchemeProbe(trimmed)) {
    return ServerStorage.normalizeBaseUrl(trimmed);
  }

  final probeFn = probe ?? defaultSchemeProbe;
  final httpsBase = ServerStorage.normalizeBaseUrl('https://$trimmed');
  final httpsResult = await probeFn(statusUriForBase(httpsBase));
  if (httpsResult.reached || httpsResult.tlsPresent) {
    if (httpsResult.reached && httpsResult.finalUri != null) {
      return baseUrlFromStatusUri(httpsResult.finalUri!);
    }
    return httpsBase;
  }

  final httpBase = ServerStorage.normalizeBaseUrl('http://$trimmed');
  final httpResult = await probeFn(statusUriForBase(httpBase));
  if (httpResult.reached) {
    if (httpResult.finalUri != null) {
      return baseUrlFromStatusUri(httpResult.finalUri!);
    }
    return httpBase;
  }

  return httpsBase;
}
