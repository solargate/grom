import 'dart:async';
import 'dart:io';

import 'server_scheme_probe_result.dart';

const _probeTimeout = Duration(seconds: 3);

/// Probes `GET statusUri` with a short timeout; follows redirects.
Future<SchemeProbeResult> defaultSchemeProbe(Uri statusUri) async {
  final client = HttpClient();
  client.connectionTimeout = _probeTimeout;
  client.idleTimeout = _probeTimeout;

  try {
    final request = await client.getUrl(statusUri).timeout(_probeTimeout);
    request.followRedirects = true;
    request.maxRedirects = 5;

    final response = await request.close().timeout(_probeTimeout);

    var effective = statusUri;
    for (final redirect in response.redirects) {
      effective = effective.resolveUri(redirect.location);
    }

    await response.drain<void>().timeout(_probeTimeout);

    return SchemeProbeResult(
      reached: true,
      finalUri: effective,
    );
  } on HandshakeException {
    return const SchemeProbeResult(reached: false, tlsPresent: true);
  } on CertificateException {
    return const SchemeProbeResult(reached: false, tlsPresent: true);
  } on TlsException {
    return const SchemeProbeResult(reached: false, tlsPresent: true);
  } on TimeoutException {
    return const SchemeProbeResult(reached: false);
  } on SocketException {
    return const SchemeProbeResult(reached: false);
  } on HttpException {
    return const SchemeProbeResult(reached: false);
  } catch (_) {
    return const SchemeProbeResult(reached: false);
  } finally {
    client.close(force: true);
  }
}
