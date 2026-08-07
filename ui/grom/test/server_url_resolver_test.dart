import 'package:flutter_test/flutter_test.dart';
import 'package:grom/server_url_resolver.dart';

void main() {
  group('needsSchemeProbe', () {
    test('true for bare host and host with path', () {
      expect(needsSchemeProbe('grom.msk.ru'), isTrue);
      expect(needsSchemeProbe('grom.example/prefix'), isTrue);
      expect(needsSchemeProbe('  grom.example  '), isTrue);
    });

    test('false when scheme or port is explicit', () {
      expect(needsSchemeProbe('https://grom.example'), isFalse);
      expect(needsSchemeProbe('http://grom.example'), isFalse);
      expect(needsSchemeProbe('HTTP://grom.example'), isFalse);
      expect(needsSchemeProbe('grom.example:8443'), isFalse);
      expect(needsSchemeProbe('ftp://grom.example'), isFalse);
      expect(needsSchemeProbe(''), isFalse);
    });

    test('IPv6 brackets: probe when no port, skip when port set', () {
      // No explicit port → probe is allowed (hasPort is false).
      expect(needsSchemeProbe('[::1]'), isTrue);
      expect(needsSchemeProbe('[::1]:8080'), isFalse);
    });
  });

  group('baseUrlFromStatusUri', () {
    test('strips status path and keeps redirect host/scheme', () {
      expect(
        baseUrlFromStatusUri(
          Uri.parse('https://www.example.com/api/v1/status'),
        ),
        'https://www.example.com',
      );
      expect(
        baseUrlFromStatusUri(
          Uri.parse('https://grom.example/prefix/api/v1/status'),
        ),
        'https://grom.example/prefix',
      );
      expect(
        baseUrlFromStatusUri(
          Uri.parse('http://192.168.1.10:8080/api/v1/status'),
        ),
        'http://192.168.1.10:8080',
      );
    });
  });

  group('resolveServerBaseUrl', () {
    test('skips probe when scheme or port provided', () async {
      var probed = false;
      final resolved = await resolveServerBaseUrl(
        'https://grom.example/',
        probe: (_) async {
          probed = true;
          return const SchemeProbeResult(reached: false);
        },
      );
      expect(probed, isFalse);
      expect(resolved, 'https://grom.example');

      final withPort = await resolveServerBaseUrl(
        'grom.example:8443',
        probe: (_) async {
          probed = true;
          return const SchemeProbeResult(reached: false);
        },
      );
      expect(withPort, 'https://grom.example:8443');
    });

    test('uses https when https probe reaches', () async {
      final resolved = await resolveServerBaseUrl(
        'grom.msk.ru',
        probe: (uri) async {
          expect(uri.scheme, 'https');
          expect(uri.path, '/api/v1/status');
          return SchemeProbeResult(reached: true, finalUri: uri);
        },
      );
      expect(resolved, 'https://grom.msk.ru');
    });

    test('uses https when TLS is present but handshake fails', () async {
      final resolved = await resolveServerBaseUrl(
        'grom.example',
        probe: (uri) async {
          if (uri.scheme == 'https') {
            return const SchemeProbeResult(reached: false, tlsPresent: true);
          }
          fail('should not probe http after tlsPresent');
        },
      );
      expect(resolved, 'https://grom.example');
    });

    test('falls back to http when https unreachable', () async {
      final resolved = await resolveServerBaseUrl(
        '192.168.1.10',
        probe: (uri) async {
          if (uri.scheme == 'https') {
            return const SchemeProbeResult(reached: false);
          }
          return SchemeProbeResult(
            reached: true,
            finalUri: Uri.parse('http://192.168.1.10/api/v1/status'),
          );
        },
      );
      expect(resolved, 'http://192.168.1.10');
    });

    test('falls back to https when both probes fail', () async {
      final resolved = await resolveServerBaseUrl(
        'offline.example',
        probe: (_) async => const SchemeProbeResult(reached: false),
      );
      expect(resolved, 'https://offline.example');
    });

    test('preserves path prefix and follows redirect final URI', () async {
      final resolved = await resolveServerBaseUrl(
        'grom.example/app',
        probe: (uri) async {
          expect(uri.toString(), 'https://grom.example/app/api/v1/status');
          return SchemeProbeResult(
            reached: true,
            finalUri: Uri.parse('https://cdn.example/app/api/v1/status'),
          );
        },
      );
      expect(resolved, 'https://cdn.example/app');
    });

    test('http probe redirect to https uses final base', () async {
      final resolved = await resolveServerBaseUrl(
        'grom.example',
        probe: (uri) async {
          if (uri.scheme == 'https') {
            return const SchemeProbeResult(reached: false);
          }
          return SchemeProbeResult(
            reached: true,
            finalUri: Uri.parse('https://grom.example/api/v1/status'),
          );
        },
      );
      expect(resolved, 'https://grom.example');
    });
  });
}
