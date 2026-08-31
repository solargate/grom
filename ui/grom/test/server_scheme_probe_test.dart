import 'package:flutter_test/flutter_test.dart';
import 'package:grom/platform/server_scheme_probe_result.dart';
import 'package:grom/platform/server_scheme_probe_stub.dart';

void main() {
  test('SchemeProbeResult stores reach and TLS flags', () {
    const ok = SchemeProbeResult(
      reached: true,
      finalUri: null,
      tlsPresent: false,
    );
    expect(ok.reached, isTrue);

    const tls = SchemeProbeResult(reached: false, tlsPresent: true);
    expect(tls.tlsPresent, isTrue);
  });

  test('defaultSchemeProbe stub returns unreachable on non-IO', () async {
    final result = await defaultSchemeProbe(Uri.parse('http://127.0.0.1:1/status'));
    expect(result.reached, isFalse);
  });
}
