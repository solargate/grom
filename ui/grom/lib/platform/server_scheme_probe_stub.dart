import 'server_scheme_probe_result.dart';

export 'server_scheme_probe_result.dart';

/// Web / non-IO stub — probing is mobile-only.
Future<SchemeProbeResult> defaultSchemeProbe(Uri statusUri) async {
  return const SchemeProbeResult(reached: false);
}
