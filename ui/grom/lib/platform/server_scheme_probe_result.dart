/// Outcome of probing `GET {base}/api/v1/status`.
class SchemeProbeResult {
  const SchemeProbeResult({
    required this.reached,
    this.tlsPresent = false,
    this.finalUri,
  });

  /// Any HTTP response was received (including non-2xx).
  final bool reached;

  /// TLS/certificate failed but the endpoint appears to speak TLS — treat as HTTPS.
  final bool tlsPresent;

  /// Final URI after redirects when [reached] is true.
  final Uri? finalUri;
}
